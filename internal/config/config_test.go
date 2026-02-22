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
