package tailscale

import "testing"

func TestEmitReadyInvokesCallback(t *testing.T) {
	server := &TSNetServer{}

	var got string
	server.SetReadyCallback(func(serviceURL string) {
		got = serviceURL
	})

	server.emitReady("https://node.ts.net")
	if got != "https://node.ts.net" {
		t.Fatalf("expected callback URL to be propagated, got %q", got)
	}
}

func TestEmitReadySkipsEmptyURL(t *testing.T) {
	server := &TSNetServer{}

	called := false
	server.SetReadyCallback(func(serviceURL string) {
		called = true
	})

	server.emitReady("")
	if called {
		t.Fatalf("expected callback to be skipped for empty URL")
	}
}
