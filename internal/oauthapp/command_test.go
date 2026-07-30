package oauthapp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateCommandPrintsAuthorizationFlags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"management-token","token_type":"Bearer"}`))
		case "/oauth-apps":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"app-id","clientSecret":"app-secret"}`))
		default:
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	var output bytes.Buffer
	cmd := newCreateCommand(&output, server.Client(), server.URL+"/token", server.URL+"/oauth-apps")
	cmd.SetArgs([]string{
		"--oauth-client-id", "management-id",
		"--oauth-client-secret", "management-secret",
		"--hostname", "portal.example",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	got := output.String()
	for _, flag := range []string{
		"--auth-mode=tailscale",
		`--tailscale-oauth-client-id="app-id"`,
		`--tailscale-oauth-client-secret="app-secret"`,
		"--tailscale-session-secret=",
	} {
		if !strings.Contains(got, flag) {
			t.Fatalf("expected output to contain %q, got %q", flag, got)
		}
	}
	if strings.Contains(got, "Start Portal:") || strings.Contains(got, "portal --") {
		t.Fatalf("expected flags only, got %q", got)
	}
}

func TestCreateUsesManagementOAuthClientAndGeneratesSecrets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			clientID, clientSecret, ok := r.BasicAuth()
			if !ok || clientID != "management-id" || clientSecret != "management-secret" {
				t.Fatalf("unexpected management credentials: id=%q secret=%q ok=%t", clientID, clientSecret, ok)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"management-token","token_type":"Bearer"}`))
		case "/oauth-apps":
			username, password, ok := r.BasicAuth()
			if !ok || username != "management-token" || password != "" {
				t.Fatalf("unexpected API token credentials: username=%q password=%q ok=%t", username, password, ok)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if got, want := payload["redirectURIs"], []any{"https://portal.example/_portal/auth/tailscale/callback"}; !equalValues(got, want) {
				t.Fatalf("unexpected redirect URIs: got %#v want %#v", got, want)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"app-id","clientSecret":"app-secret"}`))
		default:
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	result, err := Create(context.Background(), server.Client(), server.URL+"/token", server.URL+"/oauth-apps", CreateRequest{
		ManagementClientID:     "management-id",
		ManagementClientSecret: "management-secret",
		Name:                   "portal",
		Description:            "Portal browser login",
		Hostname:               "portal.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := result.ClientID, "app-id"; got != want {
		t.Fatalf("unexpected app client ID: got %q want %q", got, want)
	}
	if got, want := result.ClientSecret, "app-secret"; got != want {
		t.Fatalf("unexpected app client secret: got %q want %q", got, want)
	}
	if len(result.SessionSecret) != 64 {
		t.Fatalf("expected 32-byte hex session secret, got %q", result.SessionSecret)
	}
}

func TestCreateRejectsInvalidHostname(t *testing.T) {
	_, err := Create(context.Background(), nil, "", "", CreateRequest{
		ManagementClientID:     "id",
		ManagementClientSecret: "secret",
		Name:                   "portal",
		Hostname:               "https://portal.example/callback?state=bad",
	})
	if err == nil || !strings.Contains(err.Error(), "invalid hostname") {
		t.Fatalf("expected hostname validation error, got %v", err)
	}
}

func equalValues(got, want any) bool {
	gotJSON, _ := json.Marshal(got)
	wantJSON, _ := json.Marshal(want)
	return string(gotJSON) == string(wantJSON)
}
