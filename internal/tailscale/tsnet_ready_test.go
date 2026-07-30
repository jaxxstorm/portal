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

func TestTSNetFunnelReadinessStatus(t *testing.T) {
	tests := []struct {
		name                  string
		enableFunnel          bool
		ready                 bool
		wantCertificateStatus string
		wantFunnelStatus      string
	}{
		{
			name: "tailnet", wantCertificateStatus: "not_required", wantFunnelStatus: "disabled",
		},
		{
			name: "funnel ready", enableFunnel: true, ready: true, wantCertificateStatus: "verified", wantFunnelStatus: "ready",
		},
		{
			name: "funnel failed", enableFunnel: true, wantCertificateStatus: "not_verified", wantFunnelStatus: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &TSNetServer{config: TSNetConfig{EnableFunnel: tt.enableFunnel}}
			certificateStatus, funnelStatus := server.funnelReadinessStatus(tt.ready)
			if got, want := certificateStatus, tt.wantCertificateStatus; got != want {
				t.Fatalf("unexpected certificate status: got %q want %q", got, want)
			}
			if got, want := funnelStatus, tt.wantFunnelStatus; got != want {
				t.Fatalf("unexpected funnel status: got %q want %q", got, want)
			}
		})
	}
}
