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
	if !cfg.Funnel {
		t.Fatalf("expected funnel true in mock mode")
	}
	if !cfg.UseHTTPS {
		t.Fatalf("expected use https true in mock mode")
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
	t.Setenv("TGATE_PORT", "7070")
	t.Setenv("TGATE_FUNNEL", "true")
	t.Setenv("TGATE_FUNNEL_ALLOWLIST", "203.0.113.10, 198.51.100.0/24")
	t.Setenv("TGATE_VERBOSE", "true")
	t.Setenv("TGATE_NO_TUI", "true")

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
	t.Setenv("TGATE_SERVE_PORT", "443")
	t.Setenv("TGATE_FUNNEL", "true")
	t.Setenv("TGATE_FUNNEL_ALLOWLIST", "198.51.100.0/24,192.0.2.25")

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
	t.Setenv("TGATE_PORT", "8080")
	t.Setenv("TGATE_FUNNEL", "true")
	t.Setenv("TGATE_FUNNEL_ALLOWLIST", "203.0.113.10")

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !cfg.UseFunnelProxyProtocol() {
		t.Fatalf("expected funnel proxy protocol to be enabled")
	}
}

func TestUseFunnelProxyProtocolDisabledForNonRootSetPath(t *testing.T) {
	t.Setenv("TGATE_PORT", "8080")
	t.Setenv("TGATE_FUNNEL", "true")
	t.Setenv("TGATE_FUNNEL_ALLOWLIST", "203.0.113.10")
	t.Setenv("TGATE_SET_PATH", "/api")

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
	t.Setenv("TGATE_PORT", "8080")

	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Funnel {
		t.Fatalf("expected funnel false by default for env mode")
	}
}

func TestParseArgsRejectsMockWithPortAcrossSources(t *testing.T) {
	t.Setenv("TGATE_MOCK", "true")
	t.Setenv("TGATE_PORT", "8080")

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
	t.Setenv("TGATE_PORT", "8080")
	t.Setenv("TGATE_FUNNEL_ALLOWLIST", "203.0.113.10,not-an-ip")

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

	configDir := filepath.Join(homeDir, ".tgate")
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
