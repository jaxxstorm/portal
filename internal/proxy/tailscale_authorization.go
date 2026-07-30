package proxy

import (
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	tailscaleAuthorizationEndpoint = "https://login.tailscale.com/a/oauth_authorize"
	tailscaleTokenEndpoint         = "https://login.tailscale.com/api/v2/oauth/token"
	tailscaleJWKSURL               = "https://login.tailscale.com/.well-known/jwks.json"
	tailscaleIssuer                = "https://login.tailscale.com"
	tailscaleAuthPrefix            = "/_portal/auth/tailscale"
	tailscaleStateCookie           = "portal_tailscale_oauth_state"
	tailscaleSessionCookie         = "portal_tailscale_session"
	tailscaleStateLifetime         = 10 * time.Minute
	tailscaleSessionLifetime       = 8 * time.Hour
)

// TailscaleAuthorizationConfig configures Portal's internal OAuth App login
// endpoints.
type TailscaleAuthorizationConfig struct {
	ClientID      string
	ClientSecret  string
	SessionSecret string
	MountPath     string
	Secure        bool
	HTTPClient    *http.Client
	Now           func() time.Time
	JWKSURL       string
	AuthURL       string
	TokenURL      string
	Logger        *zap.Logger
}

type tailscaleAuthorizer struct {
	clientID      string
	clientSecret  string
	sessionSecret []byte
	mountPath     string
	secure        bool
	httpClient    *http.Client
	now           func() time.Time
	jwksURL       string
	authURL       string
	tokenURL      string
	logger        *zap.Logger

	mu     sync.Mutex
	states map[string]time.Time
}

type signedAuthState struct {
	State    string `json:"state"`
	Nonce    string `json:"nonce"`
	ReturnTo string `json:"return_to"`
	Expires  int64  `json:"expires"`
}

type signedSession struct {
	Expires int64 `json:"expires"`
}

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KID string `json:"kid"`
	KTY string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// NewTailscaleAuthorizer creates a browser-only authorizer backed by a
// Tailscale OAuth App's OIDC implicit flow.
func NewTailscaleAuthorizer(config TailscaleAuthorizationConfig) Authorizer {
	now := config.Now
	if now == nil {
		now = time.Now
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	jwksURL := config.JWKSURL
	if jwksURL == "" {
		jwksURL = tailscaleJWKSURL
	}
	authURL := config.AuthURL
	if authURL == "" {
		authURL = tailscaleAuthorizationEndpoint
	}
	tokenURL := config.TokenURL
	if tokenURL == "" {
		tokenURL = tailscaleTokenEndpoint
	}
	logger := config.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &tailscaleAuthorizer{
		clientID:      config.ClientID,
		clientSecret:  config.ClientSecret,
		sessionSecret: []byte(config.SessionSecret),
		mountPath:     normalizedMountPath(config.MountPath),
		secure:        config.Secure,
		httpClient:    httpClient,
		now:           now,
		jwksURL:       jwksURL,
		authURL:       authURL,
		tokenURL:      tokenURL,
		logger:        logger,
		states:        make(map[string]time.Time),
	}
}

func (a *tailscaleAuthorizer) Authorize(w http.ResponseWriter, r *http.Request) bool {
	if a.handleAuthRoute(w, r) {
		return false
	}
	if a.validSession(r) {
		return true
	}

	returnTo := r.URL.RequestURI()
	if !a.validReturnTo(returnTo) {
		returnTo = a.mountPath
	}
	http.Redirect(w, r, a.route("/login")+"?return_to="+url.QueryEscape(returnTo), http.StatusFound)
	return false
}

func (a *tailscaleAuthorizer) handleAuthRoute(w http.ResponseWriter, r *http.Request) bool {
	switch r.URL.Path {
	case a.route("/login"):
		a.login(w, r)
		return true
	case a.route("/callback"):
		a.callback(w, r)
		return true
	case a.route("/logout"):
		a.logout(w, r)
		return true
	default:
		return false
	}
}

func (a *tailscaleAuthorizer) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	returnTo := r.URL.Query().Get("return_to")
	if !a.validReturnTo(returnTo) {
		returnTo = a.mountPath
	}
	state, err := randomValue()
	if err != nil {
		http.Error(w, "Unable to start login", http.StatusInternalServerError)
		return
	}
	nonce, err := randomValue()
	if err != nil {
		http.Error(w, "Unable to start login", http.StatusInternalServerError)
		return
	}
	expires := a.now().Add(tailscaleStateLifetime)
	payload, err := json.Marshal(signedAuthState{State: state, Nonce: nonce, ReturnTo: returnTo, Expires: expires.Unix()})
	if err != nil {
		http.Error(w, "Unable to start login", http.StatusInternalServerError)
		return
	}
	a.mu.Lock()
	a.states[state] = expires
	a.removeExpiredStatesLocked()
	a.mu.Unlock()
	a.setCookie(w, tailscaleStateCookie, a.sign(payload), int(tailscaleStateLifetime.Seconds()))

	http.Redirect(w, r, a.oauthConfig(r).AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce)), http.StatusFound)
}

func (a *tailscaleAuthorizer) callback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Query().Get("error") != "" {
		http.Error(w, "Tailscale authorization failed", http.StatusUnauthorized)
		return
	}
	state := r.URL.Query().Get("state")
	authState, err := a.consumeState(r, state)
	if err != nil {
		http.Error(w, "Invalid OAuth state", http.StatusUnauthorized)
		return
	}
	a.setCookie(w, tailscaleStateCookie, "", -1)
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusUnauthorized)
		return
	}
	idToken, err := a.exchangeCode(r.Context(), code, a.callbackURL(r))
	if err != nil {
		a.logger.Warn("Tailscale OAuth token exchange failed", zap.Error(err))
		http.Error(w, "Tailscale token exchange failed", http.StatusUnauthorized)
		return
	}
	if err := a.verifyIDToken(r.Context(), idToken, authState.Nonce); err != nil {
		a.logger.Warn("Tailscale OAuth ID token validation failed", zap.Error(err))
		http.Error(w, "Invalid identity token", http.StatusUnauthorized)
		return
	}
	payload, err := json.Marshal(signedSession{Expires: a.now().Add(tailscaleSessionLifetime).Unix()})
	if err != nil {
		http.Error(w, "Unable to create session", http.StatusInternalServerError)
		return
	}
	a.setCookie(w, tailscaleSessionCookie, a.sign(payload), int(tailscaleSessionLifetime.Seconds()))
	http.Redirect(w, r, authState.ReturnTo, http.StatusFound)
}

func (a *tailscaleAuthorizer) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	a.setCookie(w, tailscaleSessionCookie, "", -1)
	http.Redirect(w, r, a.mountPath, http.StatusFound)
}

func (a *tailscaleAuthorizer) consumeState(r *http.Request, state string) (signedAuthState, error) {
	var authState signedAuthState
	if state == "" {
		return authState, errors.New("missing state")
	}
	cookie, err := r.Cookie(tailscaleStateCookie)
	if err != nil {
		return authState, err
	}
	if err := a.verify(cookie.Value, &authState); err != nil || authState.Expires < a.now().Unix() || subtle.ConstantTimeCompare([]byte(authState.State), []byte(state)) != 1 {
		return signedAuthState{}, errors.New("invalid state")
	}
	a.mu.Lock()
	expires, ok := a.states[state]
	delete(a.states, state)
	a.removeExpiredStatesLocked()
	a.mu.Unlock()
	if !ok || expires.Before(a.now()) {
		return signedAuthState{}, errors.New("expired or reused state")
	}
	return authState, nil
}

func (a *tailscaleAuthorizer) validSession(r *http.Request) bool {
	cookie, err := r.Cookie(tailscaleSessionCookie)
	if err != nil {
		return false
	}
	var session signedSession
	return a.verify(cookie.Value, &session) == nil && session.Expires >= a.now().Unix()
}

func (a *tailscaleAuthorizer) verifyIDToken(ctx context.Context, token, nonce string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("malformed token")
	}
	var header struct {
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
	}
	if err := decodeJWTPart(parts[0], &header); err != nil || header.Algorithm != "RS256" || header.KeyID == "" {
		return errors.New("unsupported token header")
	}
	publicKey, err := a.fetchPublicKey(ctx, header.KeyID)
	if err != nil {
		return err
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return err
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature); err != nil {
		return err
	}
	var claims struct {
		Issuer   string          `json:"iss"`
		Audience json.RawMessage `json:"aud"`
		Expires  int64           `json:"exp"`
		Nonce    string          `json:"nonce"`
		Subject  string          `json:"sub"`
	}
	if err := decodeJWTPart(parts[1], &claims); err != nil {
		return err
	}
	if claims.Issuer != tailscaleIssuer || claims.Expires < a.now().Unix() || claims.Nonce != nonce || claims.Subject == "" || !audienceContains(claims.Audience, a.clientID) {
		return errors.New("invalid token claims")
	}
	return nil
}

func (a *tailscaleAuthorizer) fetchPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.jwksURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS request returned %s", resp.Status)
	}
	var document jwksDocument
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&document); err != nil {
		return nil, err
	}
	for _, key := range document.Keys {
		if key.KID != keyID || key.KTY != "RSA" {
			continue
		}
		n, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			return nil, err
		}
		e, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			return nil, err
		}
		exponent := new(big.Int).SetBytes(e)
		if !exponent.IsInt64() || exponent.Int64() <= 1 {
			return nil, errors.New("invalid RSA exponent")
		}
		return &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(exponent.Int64())}, nil
	}
	return nil, errors.New("signing key not found")
}

func (a *tailscaleAuthorizer) setCookie(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: value, Path: a.mountPath, MaxAge: maxAge, HttpOnly: true, Secure: a.secure, SameSite: http.SameSiteLaxMode})
}

func (a *tailscaleAuthorizer) callbackURL(r *http.Request) string {
	scheme := "http"
	if a.secure {
		scheme = "https"
	}
	return scheme + "://" + r.Host + a.route("/callback")
}

func (a *tailscaleAuthorizer) oauthConfig(r *http.Request) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     a.clientID,
		ClientSecret: a.clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   a.authURL,
			TokenURL:  a.tokenURL,
			AuthStyle: oauth2.AuthStyleInHeader,
		},
		RedirectURL: a.callbackURL(r),
		Scopes:      []string{"openid"},
	}
}

func (a *tailscaleAuthorizer) exchangeCode(ctx context.Context, code, redirectURL string) (string, error) {
	form := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURL},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(a.clientID, a.clientSecret)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var response struct {
		IDToken          string `json:"id_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&response); err != nil {
		return "", err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if response.ErrorDescription != "" {
			return "", fmt.Errorf("OAuth provider returned %s: %s", response.Error, response.ErrorDescription)
		}
		return "", fmt.Errorf("OAuth provider returned %s", resp.Status)
	}
	if response.IDToken == "" {
		return "", errors.New("OAuth provider response missing id_token")
	}
	return response.IDToken, nil
}

func (a *tailscaleAuthorizer) route(path string) string {
	return strings.TrimSuffix(a.mountPath, "/") + tailscaleAuthPrefix + path
}

func (a *tailscaleAuthorizer) validReturnTo(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil && parsed.IsAbs() == false && parsed.Host == "" && strings.HasPrefix(parsed.Path, a.mountPath) && !strings.HasPrefix(value, "//")
}

func (a *tailscaleAuthorizer) sign(payload []byte) string {
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, a.sessionSecret)
	mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (a *tailscaleAuthorizer) verify(value string, target any) error {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return errors.New("invalid signed value")
	}
	mac := hmac.New(sha256.New, a.sessionSecret)
	mac.Write([]byte(parts[0]))
	expected := mac.Sum(nil)
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || subtle.ConstantTimeCompare(expected, signature) != 1 {
		return errors.New("invalid signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, target)
}

func (a *tailscaleAuthorizer) removeExpiredStatesLocked() {
	for state, expires := range a.states {
		if expires.Before(a.now()) {
			delete(a.states, state)
		}
	}
}

func randomValue() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func decodeJWTPart(part string, target any) error {
	decoded, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}
	return json.Unmarshal(decoded, target)
}

func audienceContains(raw json.RawMessage, audience string) bool {
	var single string
	if json.Unmarshal(raw, &single) == nil {
		return single == audience
	}
	var multiple []string
	if json.Unmarshal(raw, &multiple) != nil {
		return false
	}
	for _, value := range multiple {
		if value == audience {
			return true
		}
	}
	return false
}

func normalizedMountPath(path string) string {
	path = strings.TrimSuffix(strings.TrimSpace(path), "/")
	if path == "" || path == "/" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}
