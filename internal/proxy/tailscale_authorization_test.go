package proxy

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/jaxxstorm/portal/internal/model"
)

func TestTailscaleAuthorizerCompletesLoginAndRejectsStateReuse(t *testing.T) {
	key := testRSAKey(t)
	now := time.Date(2026, time.March, 31, 12, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id_token": testIDToken(t, key, "portal-client", r.FormValue("code"), now),
			})
		default:
			_ = json.NewEncoder(w).Encode(jwksDocument{Keys: []jwk{testJWK(key)}})
		}
	}))
	defer server.Close()

	a := NewTailscaleAuthorizer(TailscaleAuthorizationConfig{
		ClientID: "portal-client", ClientSecret: "app-secret", SessionSecret: "session-secret",
		Secure: true, HTTPClient: server.Client(), JWKSURL: server.URL, AuthURL: server.URL + "/authorize", TokenURL: server.URL + "/token", Now: func() time.Time { return now },
	}).(*tailscaleAuthorizer)

	loginURL, stateCookie := startLogin(t, a, "/hook?name=build")
	params := loginURL.Query()
	if got, want := params.Get("response_type"), "code"; got != want {
		t.Fatalf("expected response type %q, got %q", want, got)
	}
	if got, want := params.Get("scope"), "openid"; got != want {
		t.Fatalf("expected scope %q, got %q", want, got)
	}

	callback := httptest.NewRequest(http.MethodGet, "https://portal.example/_portal/auth/tailscale/callback?state="+url.QueryEscape(params.Get("state"))+"&code="+url.QueryEscape(params.Get("nonce")), nil)
	callback.AddCookie(stateCookie)
	callbackRecorder := httptest.NewRecorder()
	a.callback(callbackRecorder, callback)
	if got, want := callbackRecorder.Code, http.StatusFound; got != want {
		t.Fatalf("expected successful callback status %d, got %d: %s", want, got, callbackRecorder.Body.String())
	}
	if got, want := callbackRecorder.Header().Get("Location"), "/hook?name=build"; got != want {
		t.Fatalf("expected callback redirect %q, got %q", want, got)
	}
	sessionCookie := responseCookie(t, callbackRecorder, tailscaleSessionCookie)
	if !sessionCookie.HttpOnly || !sessionCookie.Secure || sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("session cookie missing secure attributes: %+v", sessionCookie)
	}

	authorized := httptest.NewRequest(http.MethodGet, "https://portal.example/hook", nil)
	authorized.AddCookie(sessionCookie)
	if !a.Authorize(httptest.NewRecorder(), authorized) {
		t.Fatalf("expected signed session to authorize request")
	}

	reused := httptest.NewRequest(http.MethodGet, "https://portal.example/_portal/auth/tailscale/callback?state="+url.QueryEscape(params.Get("state"))+"&code="+url.QueryEscape(params.Get("nonce")), nil)
	reused.AddCookie(stateCookie)
	reusedRecorder := httptest.NewRecorder()
	a.callback(reusedRecorder, reused)
	if got, want := reusedRecorder.Code, http.StatusUnauthorized; got != want {
		t.Fatalf("expected reused state status %d, got %d", want, got)
	}
}

func TestTailscaleAuthorizerRejectsTamperedSessionAndExternalReturnURL(t *testing.T) {
	a := NewTailscaleAuthorizer(TailscaleAuthorizationConfig{ClientID: "client", SessionSecret: "session"}).(*tailscaleAuthorizer)
	request := httptest.NewRequest(http.MethodGet, "http://portal.example/path", nil)
	request.AddCookie(&http.Cookie{Name: tailscaleSessionCookie, Value: "tampered"})
	response := httptest.NewRecorder()
	if a.Authorize(response, request) {
		t.Fatalf("expected tampered session to be rejected")
	}
	if got, want := response.Code, http.StatusFound; got != want {
		t.Fatalf("expected login redirect %d, got %d", want, got)
	}

	loginRequest := httptest.NewRequest(http.MethodGet, "http://portal.example/_portal/auth/tailscale/login?return_to=https://attacker.example", nil)
	loginResponse := httptest.NewRecorder()
	a.login(loginResponse, loginRequest)
	location, err := url.Parse(loginResponse.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	stateCookie := responseCookie(t, loginResponse, tailscaleStateCookie)
	var state signedAuthState
	if err := a.verify(stateCookie.Value, &state); err != nil {
		t.Fatal(err)
	}
	if got, want := state.ReturnTo, "/"; got != want {
		t.Fatalf("expected safe return target %q, got %q", want, got)
	}
	if location.Query().Get("redirect_uri") != "http://portal.example/_portal/auth/tailscale/callback" {
		t.Fatalf("unexpected callback URL %q", location.Query().Get("redirect_uri"))
	}
}

func TestTailscaleAuthorizationPreventsBackendDispatchInTailnetAndFunnelModes(t *testing.T) {
	for _, funnel := range []bool{false, true} {
		t.Run(map[bool]string{false: "tailnet", true: "funnel"}[funnel], func(t *testing.T) {
			server := NewServer(Config{
				Mode: model.ModeMock, Logger: zap.NewNop(), FunnelEnabled: funnel,
				Authorizer: NewTailscaleAuthorizer(TailscaleAuthorizationConfig{ClientID: "client", SessionSecret: "session"}),
			})
			response := httptest.NewRecorder()
			server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
			if got, want := response.Code, http.StatusFound; got != want {
				t.Fatalf("expected login redirect %d, got %d", want, got)
			}
			if got := response.Header().Get("X-portal-mode"); got != "" {
				t.Fatalf("expected request not to reach mock backend, got %q", got)
			}
		})
	}
}

func startLogin(t *testing.T, a *tailscaleAuthorizer, returnTo string) (*url.URL, *http.Cookie) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "https://portal.example/_portal/auth/tailscale/login?return_to="+url.QueryEscape(returnTo), nil)
	response := httptest.NewRecorder()
	a.login(response, request)
	if got, want := response.Code, http.StatusFound; got != want {
		t.Fatalf("expected login redirect %d, got %d", want, got)
	}
	location, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	return location, responseCookie(t, response, tailscaleStateCookie)
}

func responseCookie(t *testing.T, response *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("response did not set cookie %q", name)
	return nil
}

func testRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func testJWK(key *rsa.PrivateKey) jwk {
	return jwk{KID: "test-key", KTY: "RSA", N: base64.RawURLEncoding.EncodeToString(key.N.Bytes()), E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())}
}

func testIDToken(t *testing.T, key *rsa.PrivateKey, audience, nonce string, now time.Time) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": "RS256", "kid": "test-key"})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := json.Marshal(map[string]any{"iss": tailscaleIssuer, "aud": audience, "exp": now.Add(time.Minute).Unix(), "nonce": nonce, "sub": "user-123"})
	if err != nil {
		t.Fatal(err)
	}
	signed := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(signed))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return signed + "." + base64.RawURLEncoding.EncodeToString(signature)
}
