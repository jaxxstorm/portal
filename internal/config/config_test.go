package config

import (
	"errors"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	homeDir, err := os.MkdirTemp("", "portal-config-test-home-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(homeDir)
	_ = os.Setenv("HOME", homeDir)

	for _, entry := range os.Environ() {
		key, _, found := strings.Cut(entry, "=")
		if found && strings.HasPrefix(key, "PORTAL_") {
			_ = os.Unsetenv(key)
		}
	}
	os.Exit(m.Run())
}

func TestParseArgsCLICompatibility(t *testing.T) {
	cfg, err := ParseArgs([]string{"8080", "--funnel", "--verbose"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", cfg.Port)
	}
	if !cfg.Funnel {
		t.Fatalf("expected funnel true")
	}
	if !cfg.UseHTTPS {
		t.Fatalf("expected use https to auto-enable for funnel")
	}
	if !cfg.Verbose {
		t.Fatalf("expected verbose true")
	}
}

func TestParseArgsMockCLICompatibility(t *testing.T) {
	cfg, err := ParseArgs([]string{"--mock"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !cfg.Mock {
		t.Fatalf("expected mock true")
	}
	if cfg.Funnel {
		t.Fatalf("expected funnel false by default in mock mode")
	}
	if cfg.UseHTTPS {
		t.Fatalf("expected use https false by default in mock mode")
	}
}

func TestParseArgsMockWithExplicitFunnelCLI(t *testing.T) {
	cfg, err := ParseArgs([]string{"--mock", "--funnel"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !cfg.Mock {
		t.Fatalf("expected mock true")
	}
	if !cfg.Funnel {
		t.Fatalf("expected funnel true when explicitly enabled in mock mode")
	}
	if !cfg.UseHTTPS {
		t.Fatalf("expected use https true when funnel is enabled")
	}
}

func TestParseArgsMockAndFunnelResolvedIndependentlyFromEnv(t *testing.T) {
	t.Setenv("PORTAL_MOCK", "true")
	t.Setenv("PORTAL_FUNNEL", "false")

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !cfg.Mock {
		t.Fatalf("expected mock true from env")
	}
	if cfg.Funnel {
		t.Fatalf("expected funnel false from env when explicitly disabled")
	}
	if cfg.UseHTTPS {
		t.Fatalf("expected use https false without funnel")
	}
}

func TestParseArgsMockWithExplicitFunnelFromEnv(t *testing.T) {
	t.Setenv("PORTAL_MOCK", "true")
	t.Setenv("PORTAL_FUNNEL", "true")

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !cfg.Mock {
		t.Fatalf("expected mock true from env")
	}
	if !cfg.Funnel {
		t.Fatalf("expected funnel true from env")
	}
	if !cfg.UseHTTPS {
		t.Fatalf("expected use https true when funnel is enabled")
	}
}

func TestParseArgsCleanupCLICompatibility(t *testing.T) {
	cfg, err := ParseArgs([]string{"--cleanup-serve"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !cfg.CleanupServe {
		t.Fatalf("expected cleanup-serve true")
	}
}

func TestParseArgsReadsDefaultConfigFilePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfigFile(t, home, `
port: 9090
funnel: true
set-path: /api
`)

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Port != 9090 {
		t.Fatalf("expected port 9090, got %d", cfg.Port)
	}
	if !cfg.Funnel {
		t.Fatalf("expected funnel true from config file")
	}
	if cfg.GetSetPath() != "/api" {
		t.Fatalf("expected set-path /api, got %q", cfg.GetSetPath())
	}
}

func TestParseArgsLoadsFunnelAllowlistFromConfigFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfigFile(t, home, `
port: 9090
funnel: true
funnel-allowlist:
  - 203.0.113.10
  - 198.51.100.0/24
`)

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := funnelAllowlistStrings(cfg.FunnelAllowlist), []string{"203.0.113.10/32", "198.51.100.0/24"}; !slices.Equal(got, want) {
		t.Fatalf("expected funnel allowlist %v from config file, got %v", want, got)
	}
}

func TestParseArgsBindsEnvironmentVariables(t *testing.T) {
	t.Setenv("PORTAL_PORT", "7070")
	t.Setenv("PORTAL_FUNNEL", "true")
	t.Setenv("PORTAL_FUNNEL_ALLOWLIST", "203.0.113.10, 198.51.100.0/24")
	t.Setenv("PORTAL_VERBOSE", "true")
	t.Setenv("PORTAL_NO_TUI", "true")

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Port != 7070 {
		t.Fatalf("expected port 7070, got %d", cfg.Port)
	}
	if !cfg.Funnel {
		t.Fatalf("expected funnel true from env")
	}
	if !cfg.Verbose {
		t.Fatalf("expected verbose true from env")
	}
	if !cfg.NoTUI {
		t.Fatalf("expected no-tui true from env")
	}
	if got, want := funnelAllowlistStrings(cfg.FunnelAllowlist), []string{"203.0.113.10/32", "198.51.100.0/24"}; !slices.Equal(got, want) {
		t.Fatalf("expected funnel allowlist %v from env, got %v", want, got)
	}
}

func TestParseArgsPrecedenceCLIOverEnvOverConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfigFile(t, home, `
port: 9000
serve-port: 80
funnel: false
funnel-allowlist:
  - 203.0.113.10
`)
	t.Setenv("PORTAL_SERVE_PORT", "443")
	t.Setenv("PORTAL_FUNNEL", "true")
	t.Setenv("PORTAL_FUNNEL_ALLOWLIST", "198.51.100.0/24,192.0.2.25")

	cfg, err := ParseArgs([]string{"8080", "--serve-port", "8443"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Port != 8080 {
		t.Fatalf("expected CLI port to win with 8080, got %d", cfg.Port)
	}
	if cfg.ServePort != 8443 {
		t.Fatalf("expected CLI serve-port to win with 8443, got %d", cfg.ServePort)
	}
	if !cfg.Funnel {
		t.Fatalf("expected env funnel=true to override config")
	}
	if got, want := funnelAllowlistStrings(cfg.FunnelAllowlist), []string{"198.51.100.0/24", "192.0.2.25/32"}; !slices.Equal(got, want) {
		t.Fatalf("expected env funnel allowlist %v to override config, got %v", want, got)
	}
}

func TestParseArgsDeviceNameDefaultsToTgate(t *testing.T) {
	cfg, err := ParseArgs([]string{"8080"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got, want := cfg.TailscaleName, "portal"; got != want {
		t.Fatalf("expected default device name %q, got %q", want, got)
	}
}

func TestParseArgsDeviceNamePrecedenceCLIOverEnvOverConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfigFile(t, home, `
port: 9000
device-name: cfg-device
`)
	t.Setenv("PORTAL_DEVICE_NAME", "env-device")

	cfg, err := ParseArgs([]string{"8080", "--device-name", "cli-device"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got, want := cfg.TailscaleName, "cli-device"; got != want {
		t.Fatalf("expected CLI device name %q, got %q", want, got)
	}
}

func TestParseArgsSupportsLegacyTailscaleName(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_TAILSCALE_NAME", "legacy-node")

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got, want := cfg.TailscaleName, "legacy-node"; got != want {
		t.Fatalf("expected legacy tailscale name %q, got %q", want, got)
	}
}

func TestParseArgsRejectsConflictingDeviceAndLegacyTailscaleName(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_DEVICE_NAME", "device-node")
	t.Setenv("PORTAL_TAILSCALE_NAME", "legacy-node")

	_, err := ParseArgs([]string{})
	if err == nil {
		t.Fatalf("expected conflicting device-name error")
	}
	if got, want := err.Error(), "conflicting configuration: device-name"; !strings.Contains(got, want) {
		t.Fatalf("expected error containing %q, got %q", want, got)
	}
}

func TestParseArgsTSNetListenModeDefaults(t *testing.T) {
	cfg, err := ParseArgs([]string{"8080"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := cfg.TSNetListenMode, TSNetListenModeListener; got != want {
		t.Fatalf("expected default tsnet listen mode %q, got %q", want, got)
	}
	if got, want := cfg.TSNetServiceName, "svc:portal"; got != want {
		t.Fatalf("expected default tsnet service name %q, got %q", want, got)
	}
	if cfg.IsServiceMode() {
		t.Fatalf("expected listener default mode to not be service mode")
	}
}

func TestGetServePortDefaultsToTargetPortInServiceMode(t *testing.T) {
	cfg, err := ParseArgs([]string{"8080", "--listen-mode", "service", "--service-name", "svc:portal"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := cfg.GetServePort(), 8080; got != want {
		t.Fatalf("expected default serve-port %d in service mode, got %d", want, got)
	}
}

func TestGetServePortDefaultsToEightyInListenerMode(t *testing.T) {
	cfg, err := ParseArgs([]string{"8080", "--listen-mode", "listener"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := cfg.GetServePort(), 80; got != want {
		t.Fatalf("expected default serve-port %d in listener mode, got %d", want, got)
	}
}

func TestGetServePortUsesExplicitOverrideInServiceMode(t *testing.T) {
	cfg, err := ParseArgs([]string{"8080", "--listen-mode", "service", "--service-name", "svc:portal", "--serve-port", "9090"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := cfg.GetServePort(), 9090; got != want {
		t.Fatalf("expected explicit serve-port %d to win, got %d", want, got)
	}
}

func TestGetServePortDefaultsToHTTPSPortInServiceMode(t *testing.T) {
	cfg, err := ParseArgs([]string{"8080", "--listen-mode", "service", "--service-name", "svc:portal", "--use-https"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := cfg.GetServePort(), 443; got != want {
		t.Fatalf("expected default HTTPS serve-port %d, got %d", want, got)
	}
}

func TestParseArgsTSNetListenModePrecedenceCLIOverEnvOverConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfigFile(t, home, `
port: 9000
listen-mode: listener
service-name: svc:cfg
`)
	t.Setenv("PORTAL_LISTEN_MODE", "service")
	t.Setenv("PORTAL_SERVICE_NAME", "svc:env")

	cfg, err := ParseArgs([]string{"8080", "--listen-mode", "listener", "--service-name", "svc:cli"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := cfg.TSNetListenMode, TSNetListenModeListener; got != want {
		t.Fatalf("expected CLI tsnet listen mode %q, got %q", want, got)
	}
	if got, want := cfg.TSNetServiceName, "svc:cli"; got != want {
		t.Fatalf("expected CLI tsnet service name %q, got %q", want, got)
	}
}

func TestParseArgsTSNetServiceModeFromEnvironment(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_LISTEN_MODE", "service")
	t.Setenv("PORTAL_SERVICE_NAME", "svc:from-env")

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := cfg.TSNetListenMode, TSNetListenModeService; got != want {
		t.Fatalf("expected tsnet listen mode %q, got %q", want, got)
	}
	if got, want := cfg.TSNetServiceName, "svc:from-env"; got != want {
		t.Fatalf("expected tsnet service name %q, got %q", want, got)
	}
	if !cfg.IsServiceMode() {
		t.Fatalf("expected service listen mode to be reported as service mode")
	}
}

func TestParseArgsSupportsLegacyTSNetModeNames(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_TSNET_LISTEN_MODE", "service")
	t.Setenv("PORTAL_TSNET_SERVICE_NAME", "svc:legacy")

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got, want := cfg.TSNetListenMode, TSNetListenModeService; got != want {
		t.Fatalf("expected listen mode %q from legacy env, got %q", want, got)
	}
	if got, want := cfg.TSNetServiceName, "svc:legacy"; got != want {
		t.Fatalf("expected service name %q from legacy env, got %q", want, got)
	}
}

func TestParseArgsSupportsLegacyTSNetModeNamesInConfigFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfigFile(t, home, `
port: 8080
tsnet-listen-mode: service
tsnet-service-name: svc:legacy-config
`)

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got, want := cfg.TSNetListenMode, TSNetListenModeService; got != want {
		t.Fatalf("expected listen mode %q from legacy config, got %q", want, got)
	}
	if got, want := cfg.TSNetServiceName, "svc:legacy-config"; got != want {
		t.Fatalf("expected service name %q from legacy config, got %q", want, got)
	}
}

func TestParseArgsRejectsInvalidTSNetListenMode(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_LISTEN_MODE", "bogus")

	_, err := ParseArgs([]string{})
	if err == nil {
		t.Fatalf("expected listen mode validation error")
	}
	if got, want := err.Error(), `invalid listen-mode "bogus"`; !strings.Contains(got, want) {
		t.Fatalf("expected error containing %q, got %q", want, got)
	}
}

func TestParseArgsRejectsInvalidTSNetServiceName(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_LISTEN_MODE", "service")
	t.Setenv("PORTAL_SERVICE_NAME", "not-a-service")

	_, err := ParseArgs([]string{})
	if err == nil {
		t.Fatalf("expected tsnet service name validation error")
	}
	if got, want := err.Error(), `invalid service-name "not-a-service"`; !strings.Contains(got, want) {
		t.Fatalf("expected error containing %q, got %q", want, got)
	}
}

func TestParseArgsRejectsServiceModeWithFunnel(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_LISTEN_MODE", "service")
	t.Setenv("PORTAL_FUNNEL", "true")

	_, err := ParseArgs([]string{})
	if err == nil {
		t.Fatalf("expected mutual exclusivity validation error")
	}
	if got, want := err.Error(), "cannot be combined with funnel=true"; !strings.Contains(got, want) {
		t.Fatalf("expected error containing %q, got %q", want, got)
	}
}

func TestParseArgsRejectsConflictingCanonicalAndLegacyListenMode(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_LISTEN_MODE", "listener")
	t.Setenv("PORTAL_TSNET_LISTEN_MODE", "service")

	_, err := ParseArgs([]string{})
	if err == nil {
		t.Fatalf("expected conflicting mode-name error")
	}
	if got, want := err.Error(), "conflicting configuration: listen-mode"; !strings.Contains(got, want) {
		t.Fatalf("expected error containing %q, got %q", want, got)
	}
}

func TestParseArgsRejectsConflictingCanonicalAndLegacyServiceName(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_LISTEN_MODE", "service")
	t.Setenv("PORTAL_SERVICE_NAME", "svc:canonical")
	t.Setenv("PORTAL_TSNET_SERVICE_NAME", "svc:legacy")

	_, err := ParseArgs([]string{})
	if err == nil {
		t.Fatalf("expected conflicting service-name error")
	}
	if got, want := err.Error(), "conflicting configuration: service-name"; !strings.Contains(got, want) {
		t.Fatalf("expected error containing %q, got %q", want, got)
	}
}

func TestParseArgsTailnetDefaultFromConfigFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfigFile(t, home, `
port: 8081
`)

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Funnel {
		t.Fatalf("expected funnel false by default")
	}
	if cfg.UseHTTPS {
		t.Fatalf("expected use-https false by default")
	}
}

func TestUseFunnelProxyProtocolEnabledForRootPathAllowlist(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_FUNNEL", "true")
	t.Setenv("PORTAL_FUNNEL_ALLOWLIST", "203.0.113.10")

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !cfg.UseFunnelProxyProtocol() {
		t.Fatalf("expected funnel proxy protocol to be enabled")
	}
}

func TestUseFunnelProxyProtocolDisabledForNonRootSetPath(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_FUNNEL", "true")
	t.Setenv("PORTAL_FUNNEL_ALLOWLIST", "203.0.113.10")
	t.Setenv("PORTAL_SET_PATH", "/api")

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.UseFunnelProxyProtocol() {
		t.Fatalf("expected funnel proxy protocol to be disabled for non-root set-path")
	}
}

func TestParseArgsTailnetDefaultFromCLI(t *testing.T) {
	cfg, err := ParseArgs([]string{"8080"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Funnel {
		t.Fatalf("expected funnel false by default for CLI mode")
	}
}

func TestParseArgsTailnetDefaultFromEnvironment(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Funnel {
		t.Fatalf("expected funnel false by default for env mode")
	}
}

func TestParseArgsAuthorizationDefaultsToNone(t *testing.T) {
	cfg, err := ParseArgs([]string{"8080"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := cfg.AuthMode, AuthModeNone; got != want {
		t.Fatalf("expected auth mode %q, got %q", want, got)
	}
	if cfg.AuthUsername != "" || cfg.AuthPassword != "" {
		t.Fatalf("expected empty default authorization credentials")
	}
}

func TestParseArgsLoadsBasicAuthorizationFromConfigFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfigFile(t, home, `
port: 8080
auth-mode: basic
auth-username: alice
auth-password: config-secret
`)

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := cfg.AuthMode, AuthModeBasic; got != want {
		t.Fatalf("expected auth mode %q, got %q", want, got)
	}
	if got, want := cfg.AuthUsername, "alice"; got != want {
		t.Fatalf("expected username %q, got %q", want, got)
	}
	if got, want := cfg.AuthPassword, "config-secret"; got != want {
		t.Fatalf("expected password %q, got %q", want, got)
	}
}

func TestParseArgsLoadsBasicAuthorizationFromEnvironment(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_AUTH_MODE", "basic")
	t.Setenv("PORTAL_AUTH_USERNAME", "alice")
	t.Setenv("PORTAL_AUTH_PASSWORD", "environment-secret")

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := cfg.AuthMode, AuthModeBasic; got != want {
		t.Fatalf("expected auth mode %q, got %q", want, got)
	}
	if got, want := cfg.AuthUsername, "alice"; got != want {
		t.Fatalf("expected username %q, got %q", want, got)
	}
	if got, want := cfg.AuthPassword, "environment-secret"; got != want {
		t.Fatalf("expected password %q, got %q", want, got)
	}
}

func TestParseArgsAuthorizationCLIOverridesEnvironmentAndConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfigFile(t, home, `
port: 8080
auth-mode: basic
auth-username: config-user
auth-password: config-password
`)
	t.Setenv("PORTAL_AUTH_USERNAME", "env-user")
	t.Setenv("PORTAL_AUTH_PASSWORD", "env-password")

	cfg, err := ParseArgs([]string{"--auth-username", "cli-user", "--auth-password", "cli-password"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := cfg.AuthUsername, "cli-user"; got != want {
		t.Fatalf("expected CLI username %q, got %q", want, got)
	}
	if got, want := cfg.AuthPassword, "cli-password"; got != want {
		t.Fatalf("expected CLI password %q, got %q", want, got)
	}
}

func TestParseArgsAuthorizationValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing username",
			args: []string{"8080", "--auth-mode", "basic", "--auth-password", "secret"},
			want: "auth-username is required",
		},
		{
			name: "missing password",
			args: []string{"8080", "--auth-mode", "basic", "--auth-username", "alice"},
			want: "auth-password is required",
		},
		{
			name: "credentials without basic mode",
			args: []string{"8080", "--auth-username", "alice", "--auth-password", "secret"},
			want: "require auth-mode=basic",
		},
		{
			name: "unsupported mode",
			args: []string{"8080", "--auth-mode", "oauth"},
			want: `invalid auth-mode "oauth"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArgs(tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestParseArgsLoadsTailscaleAuthorizationFromEnvironment(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_AUTH_MODE", "tailscale")
	t.Setenv("PORTAL_TAILSCALE_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("PORTAL_TAILSCALE_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("PORTAL_TAILSCALE_SESSION_SECRET", "session-secret")

	cfg, err := ParseArgs(nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got, want := cfg.AuthMode, AuthModeTailscale; got != want {
		t.Fatalf("expected auth mode %q, got %q", want, got)
	}
	if !cfg.UseHTTPS {
		t.Fatalf("expected Tailscale login to enable HTTPS for secure session cookies")
	}
}

func TestParseArgsTailscaleAuthorizationCLIOverridesEnvironmentAndConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfigFile(t, home, `
port: 8080
auth-mode: tailscale
tailscale-oauth-client-id: config-client
tailscale-oauth-client-secret: config-secret
tailscale-session-secret: config-session
`)
	t.Setenv("PORTAL_TAILSCALE_OAUTH_CLIENT_ID", "env-client")

	cfg, err := ParseArgs([]string{"--tailscale-oauth-client-id", "cli-client"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got, want := cfg.TailscaleOAuthClientID, "cli-client"; got != want {
		t.Fatalf("expected client ID %q, got %q", want, got)
	}
}

func TestParseArgsTailscaleAuthorizationValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing client ID",
			args: []string{"8080", "--auth-mode", "tailscale", "--tailscale-oauth-client-secret", "secret", "--tailscale-session-secret", "session"},
			want: "tailscale-oauth-client-id is required",
		},
		{
			name: "missing client secret",
			args: []string{"8080", "--auth-mode", "tailscale", "--tailscale-oauth-client-id", "id", "--tailscale-session-secret", "session"},
			want: "tailscale-oauth-client-secret is required",
		},
		{
			name: "missing session secret",
			args: []string{"8080", "--auth-mode", "tailscale", "--tailscale-oauth-client-id", "id", "--tailscale-oauth-client-secret", "secret"},
			want: "tailscale-session-secret is required",
		},
		{
			name: "credentials without tailscale mode",
			args: []string{"8080", "--tailscale-oauth-client-id", "id"},
			want: "Tailscale OAuth credentials require auth-mode=tailscale",
		},
		{
			name: "basic credentials cannot be combined",
			args: []string{"8080", "--auth-mode", "tailscale", "--auth-username", "alice", "--tailscale-oauth-client-id", "id", "--tailscale-oauth-client-secret", "secret", "--tailscale-session-secret", "session"},
			want: "cannot be combined with auth-mode=tailscale",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArgs(tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestParseArgsRejectsMockWithPortAcrossSources(t *testing.T) {
	t.Setenv("PORTAL_MOCK", "true")
	t.Setenv("PORTAL_PORT", "8080")

	_, err := ParseArgs([]string{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestParseArgsRequiresPortUnlessMock(t *testing.T) {
	_, err := ParseArgs([]string{})
	if err == nil {
		t.Fatalf("expected missing port error")
	}
}

func TestParseArgsRejectsInvalidFunnelAllowlistEntry(t *testing.T) {
	t.Setenv("PORTAL_PORT", "8080")
	t.Setenv("PORTAL_FUNNEL_ALLOWLIST", "203.0.113.10,not-an-ip")

	_, err := ParseArgs([]string{})
	if err == nil {
		t.Fatalf("expected invalid funnel allowlist error")
	}
	if got, want := err.Error(), `invalid funnel allowlist entry "not-an-ip"`; !strings.Contains(got, want) {
		t.Fatalf("expected error containing %q, got %q", want, got)
	}
}

func TestParseArgsAllowsVersionAndCleanupWithoutPort(t *testing.T) {
	versionCfg, err := ParseArgs([]string{"--version"})
	if err != nil {
		t.Fatalf("expected no error for --version, got %v", err)
	}
	if !versionCfg.Version {
		t.Fatalf("expected version flag to be set")
	}

	cleanupCfg, err := ParseArgs([]string{"--cleanup-serve"})
	if err != nil {
		t.Fatalf("expected no error for --cleanup-serve, got %v", err)
	}
	if !cleanupCfg.CleanupServe {
		t.Fatalf("expected cleanup-serve flag to be set")
	}
}

func TestParseArgsHelpReturnsHelpError(t *testing.T) {
	_, err := ParseArgs([]string{"--help"})
	if !errors.Is(err, pflag.ErrHelp) {
		t.Fatalf("expected pflag.ErrHelp, got %v", err)
	}
}

func TestParseArgsCommandOrientedServing(t *testing.T) {
	serve, err := ParseArgs([]string{"serve", "8080"})
	if err != nil {
		t.Fatalf("serve command: %v", err)
	}
	if got, want := serve.Port, 8080; got != want || serve.Mock || serve.Funnel {
		t.Fatalf("unexpected serve config: %+v", serve)
	}

	t.Setenv("PORTAL_PORT", "8080")
	mock, err := ParseArgs([]string{"mock"})
	if err != nil {
		t.Fatalf("mock command: %v", err)
	}
	if !mock.Mock || mock.Port != 0 || mock.Funnel {
		t.Fatalf("unexpected mock config: %+v", mock)
	}
}

func TestParseArgsNodeCommandsSelectIsolatedTailscaleNode(t *testing.T) {
	serve, err := ParseArgs([]string{"serve", "node", "8080", "--device-name", "portal-test"})
	if err != nil {
		t.Fatalf("serve node command: %v", err)
	}
	if !serve.ForceTsnet || serve.Mock || serve.Port != 8080 || serve.TailscaleName != "portal-test" {
		t.Fatalf("unexpected serve node config: %+v", serve)
	}

	mock, err := ParseArgs([]string{"mock", "node"})
	if err != nil {
		t.Fatalf("mock node command: %v", err)
	}
	if !mock.ForceTsnet || !mock.Mock || mock.Port != 0 {
		t.Fatalf("unexpected mock node config: %+v", mock)
	}
}

func TestCommandOrientedHelpHidesAdvancedFlags(t *testing.T) {
	cmd, err := newRootCommand(viper.New(), &parseState{}, commandArgs{serve: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"auth-key", deviceNameKey, listenModeKey, "set-path", "serve-port", "ui-port", "auth-password", "tailscale-session-secret"} {
		flag := cmd.Flags().Lookup(name)
		if flag == nil || !flag.Hidden {
			t.Fatalf("expected %q to be hidden from serve help", name)
		}
	}
	if flag := cmd.Flags().Lookup("funnel"); flag == nil || flag.Hidden {
		t.Fatalf("expected funnel to remain visible in serve help")
	}
}

func TestPublicCommandHelpScopesFlags(t *testing.T) {
	root := newCommandHelpRoot()
	if root.Flags().Lookup("funnel") != nil || root.Flags().Lookup("auth-key") != nil {
		t.Fatalf("expected root help to contain no serving flags")
	}
	if root.PersistentFlags().Lookup("version") == nil {
		t.Fatalf("expected version to be the global flag")
	}

	serve, _, err := root.Find([]string{"serve"})
	if err != nil {
		t.Fatal(err)
	}
	if serve.Flags().Lookup("funnel") == nil || serve.Flags().Lookup("auth-mode") == nil {
		t.Fatalf("expected common serving flags")
	}
	if serve.Flags().Lookup("auth-key") != nil || serve.Flags().Lookup("serve-port") != nil {
		t.Fatalf("expected advanced flags to stay out of public serve help")
	}
	node, _, err := root.Find([]string{"serve", "node"})
	if err != nil {
		t.Fatal(err)
	}
	if node.Flags().Lookup("auth-key") == nil || node.Flags().Lookup(deviceNameKey) == nil {
		t.Fatalf("expected isolated-node settings in node help")
	}

	cleanup, _, err := root.Find([]string{"cleanup"})
	if err != nil {
		t.Fatal(err)
	}
	if cleanup.Flags().Lookup("funnel") != nil || cleanup.Flags().Lookup("auth-key") != nil {
		t.Fatalf("cleanup should not expose serving flags")
	}
}

func TestParseArgsMaintenanceCommandsSkipServingValidation(t *testing.T) {
	t.Setenv("PORTAL_PORT", "-1")
	cleanup, err := ParseArgs([]string{"cleanup"})
	if err != nil {
		t.Fatalf("cleanup command: %v", err)
	}
	if !cleanup.CleanupServe {
		t.Fatalf("expected cleanup command to set CleanupServe")
	}

	init, err := ParseArgs([]string{"config", "init", "--force"})
	if err != nil {
		t.Fatalf("config init command: %v", err)
	}
	if !init.ConfigInit || !init.ConfigInitForce {
		t.Fatalf("unexpected config init command: %+v", init)
	}

	oauthApp, err := ParseArgs([]string{"oauth-app", "create", "--help"})
	if err != nil {
		t.Fatalf("oauth-app command: %v", err)
	}
	if got, want := oauthApp.OAuthAppArgs, []string{"oauth-app", "create", "--help"}; !slices.Equal(got, want) {
		t.Fatalf("unexpected OAuth App args: got %#v want %#v", got, want)
	}
}

func TestWriteConfigTemplate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath, err := WriteConfigTemplate(false)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "funnel: false") || strings.Contains(string(contents), "tailscale-oauth-client-secret: \"tskey") {
		t.Fatalf("unexpected template contents: %s", contents)
	}
	if info, err := os.Stat(configPath); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("expected user-only config file permissions, info=%v err=%v", info, err)
	}

	if _, err := WriteConfigTemplate(false); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected overwrite protection error, got %v", err)
	}
	if err := os.WriteFile(configPath, []byte("port: 9999\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteConfigTemplate(true); err != nil {
		t.Fatal(err)
	}
	contents, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "port: 8080") {
		t.Fatalf("force did not replace template: %s", contents)
	}
}

func writeConfigFile(t *testing.T, homeDir, content string) {
	t.Helper()

	configDir := filepath.Join(homeDir, ".portal")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yml")
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}

func funnelAllowlistStrings(prefixes []netip.Prefix) []string {
	result := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		result = append(result, prefix.String())
	}
	return result
}
