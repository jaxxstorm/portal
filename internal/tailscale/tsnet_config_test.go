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
