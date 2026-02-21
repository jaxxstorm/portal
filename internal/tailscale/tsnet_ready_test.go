package tailscale

import "testing"

func TestEmitReadyInvokesCallback(t *testing.T) {
	server := &TSNetServer{}

	var got TSNetReadyInfo
	server.SetReadyCallback(func(info TSNetReadyInfo) {
		got = info
	})

	server.emitReady(TSNetReadyInfo{
		ServiceURL:           "https://node.ts.net",
		ConfiguredListenMode: TSNetListenModeService,
		EffectiveListenMode:  TSNetListenModeListener,
	})
	if got.ServiceURL != "https://node.ts.net" {
		t.Fatalf("expected callback URL to be propagated, got %q", got.ServiceURL)
	}
	if got.ConfiguredListenMode != TSNetListenModeService {
		t.Fatalf("expected configured listen mode to be propagated")
	}
}

func TestEmitReadySkipsEmptyURL(t *testing.T) {
	server := &TSNetServer{}

	called := false
	server.SetReadyCallback(func(info TSNetReadyInfo) {
		called = true
	})

	server.emitReady(TSNetReadyInfo{})
	if called {
		t.Fatalf("expected callback to be skipped for empty URL")
	}
}
