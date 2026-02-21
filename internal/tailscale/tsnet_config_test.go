package tailscale

import "testing"

func TestResolveTSNetServePort(t *testing.T) {
	tests := []struct {
		name   string
		config TSNetConfig
		want   int
	}{
		{
			name:   "default",
			config: TSNetConfig{},
			want:   80,
		},
		{
			name: "https defaults to 443",
			config: TSNetConfig{
				UseHTTPS: true,
			},
			want: 443,
		},
		{
			name: "funnel defaults to 443",
			config: TSNetConfig{
				EnableFunnel: true,
			},
			want: 443,
		},
		{
			name: "explicit port wins",
			config: TSNetConfig{
				ServePort: 8443,
			},
			want: 8443,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveTSNetServePort(tt.config); got != tt.want {
				t.Fatalf("unexpected serve port: got %d want %d", got, tt.want)
			}
		})
	}
}

func TestShouldUseTSNetTLS(t *testing.T) {
	tests := []struct {
		name      string
		config    TSNetConfig
		servePort int
		want      bool
	}{
		{
			name:      "default http",
			config:    TSNetConfig{},
			servePort: 80,
			want:      false,
		},
		{
			name: "https enabled",
			config: TSNetConfig{
				UseHTTPS: true,
			},
			servePort: 443,
			want:      true,
		},
		{
			name: "funnel forces tls",
			config: TSNetConfig{
				EnableFunnel: true,
			},
			servePort: 443,
			want:      true,
		},
		{
			name:      "port 443 implies tls",
			config:    TSNetConfig{},
			servePort: 443,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldUseTSNetTLS(tt.config, tt.servePort); got != tt.want {
				t.Fatalf("unexpected tls setting: got %t want %t", got, tt.want)
			}
		})
	}
}

func TestValidateTSNetServeConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    TSNetConfig
		servePort int
		useTLS    bool
		wantErr   bool
	}{
		{
			name:      "non-funnel ignores funnel validation",
			config:    TSNetConfig{},
			servePort: 80,
			useTLS:    false,
			wantErr:   false,
		},
		{
			name: "funnel valid with 443",
			config: TSNetConfig{
				EnableFunnel: true,
			},
			servePort: 443,
			useTLS:    true,
			wantErr:   false,
		},
		{
			name: "funnel valid with 8443",
			config: TSNetConfig{
				EnableFunnel: true,
			},
			servePort: 8443,
			useTLS:    true,
			wantErr:   false,
		},
		{
			name: "funnel invalid without tls",
			config: TSNetConfig{
				EnableFunnel: true,
			},
			servePort: 443,
			useTLS:    false,
			wantErr:   true,
		},
		{
			name: "funnel invalid port",
			config: TSNetConfig{
				EnableFunnel: true,
			},
			servePort: 80,
			useTLS:    true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTSNetServeConfig(tt.config, tt.servePort, tt.useTLS)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: err=%v wantErr=%t", err, tt.wantErr)
			}
		})
	}
}

func TestBuildTSNetServiceURL(t *testing.T) {
	tests := []struct {
		name      string
		dnsName   string
		servePort int
		useTLS    bool
		want      string
	}{
		{
			name:      "http default port",
			dnsName:   "node.ts.net",
			servePort: 80,
			useTLS:    false,
			want:      "http://node.ts.net",
		},
		{
			name:      "https default port",
			dnsName:   "node.ts.net",
			servePort: 443,
			useTLS:    true,
			want:      "https://node.ts.net",
		},
		{
			name:      "https custom port",
			dnsName:   "node.ts.net",
			servePort: 8443,
			useTLS:    true,
			want:      "https://node.ts.net:8443",
		},
		{
			name:      "http custom port",
			dnsName:   "node.ts.net",
			servePort: 8080,
			useTLS:    false,
			want:      "http://node.ts.net:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildTSNetServiceURL(tt.dnsName, tt.servePort, tt.useTLS); got != tt.want {
				t.Fatalf("unexpected URL: got %q want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeTSNetListenMode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "default empty", in: "", want: TSNetListenModeListener},
		{name: "trims and lowercases", in: " Service ", want: TSNetListenModeService},
		{name: "listener", in: "listener", want: TSNetListenModeListener},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeTSNetListenMode(tt.in); got != tt.want {
				t.Fatalf("unexpected normalized mode: got %q want %q", got, tt.want)
			}
		})
	}
}

func TestEffectiveTSNetListenMode(t *testing.T) {
	tests := []struct {
		name          string
		configured    string
		funnelEnabled bool
		wantMode      string
		wantReason    string
	}{
		{
			name:          "service mode is unchanged",
			configured:    TSNetListenModeService,
			funnelEnabled: true,
			wantMode:      TSNetListenModeService,
			wantReason:    "",
		},
		{
			name:          "service remains service in tailnet mode",
			configured:    TSNetListenModeService,
			funnelEnabled: false,
			wantMode:      TSNetListenModeService,
			wantReason:    "",
		},
		{
			name:          "listener stays listener",
			configured:    TSNetListenModeListener,
			funnelEnabled: true,
			wantMode:      TSNetListenModeListener,
			wantReason:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, reason := effectiveTSNetListenMode(tt.configured, tt.funnelEnabled)
			if mode != tt.wantMode || reason != tt.wantReason {
				t.Fatalf("unexpected effective mode result: got mode=%q reason=%q want mode=%q reason=%q", mode, reason, tt.wantMode, tt.wantReason)
			}
		})
	}
}

func TestValidateTSNetListenConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  TSNetConfig
		mode    string
		wantErr bool
	}{
		{
			name: "listener mode valid",
			config: TSNetConfig{
				ListenMode: TSNetListenModeListener,
			},
			mode:    TSNetListenModeListener,
			wantErr: false,
		},
		{
			name: "service mode valid",
			config: TSNetConfig{
				ListenMode:  TSNetListenModeService,
				ServiceName: "svc:tgate",
				ServePort:   443,
			},
			mode:    TSNetListenModeService,
			wantErr: false,
		},
		{
			name: "service mode invalid service name",
			config: TSNetConfig{
				ListenMode:  TSNetListenModeService,
				ServiceName: "not-valid",
				ServePort:   443,
			},
			mode:    TSNetListenModeService,
			wantErr: true,
		},
		{
			name: "service mode invalid with funnel",
			config: TSNetConfig{
				ListenMode:   TSNetListenModeService,
				ServiceName:  "svc:tgate",
				EnableFunnel: true,
				ServePort:    443,
			},
			mode:    TSNetListenModeService,
			wantErr: true,
		},
		{
			name: "unknown mode invalid",
			config: TSNetConfig{
				ListenMode: "other",
			},
			mode:    "other",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTSNetListenConfig(tt.config, tt.mode)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: err=%v wantErr=%t", err, tt.wantErr)
			}
		})
	}
}
