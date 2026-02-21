package config

import (
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"tailscale.com/tailcfg"
)

const (
	TSNetListenModeListener = "listener"
	TSNetListenModeService  = "service"

	deviceNameKey          = "device-name"
	legacyTailscaleNameKey = "tailscale-name"
	listenModeKey          = "listen-mode"
	legacyListenModeKey    = "tsnet-listen-mode"
	serviceNameKey         = "service-name"
	legacyServiceNameKey   = "tsnet-service-name"
)

// Config holds the parsed and validated configuration
type Config struct {
	Port             int
	TailscaleName    string
	Funnel           bool
	FunnelAllowlist  []netip.Prefix
	Verbose          bool
	JSON             bool
	LogFile          string
	AuthKey          string
	ForceTsnet       bool
	SetPath          string
	ServePort        int
	UseHTTPS         bool
	NoTUI            bool
	NoUI             bool
	UIPort           int
	Version          bool
	Mock             bool
	CleanupServe     bool
	TSNetListenMode  string
	TSNetServiceName string
}

// Parse parses command line arguments and returns a validated configuration
func Parse() (*Config, error) {
	return ParseArgs(os.Args[1:])
}

// ParseArgs parses command line arguments and returns a validated configuration.
// Exposed for tests.
func ParseArgs(args []string) (*Config, error) {
	v := viper.New()
	if err := configureViper(v); err != nil {
		return nil, err
	}

	state := &parseState{}
	cmd, err := newRootCommand(v, state)
	if err != nil {
		return nil, err
	}

	cmd.SetArgs(args)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		return nil, err
	}

	if helpRequested(cmd, args) {
		return nil, pflag.ErrHelp
	}

	port := v.GetInt("port")
	if state.portSet {
		port = state.port
	}

	deviceName, err := resolveAliasedSetting(v, deviceNameKey, legacyTailscaleNameKey, strings.TrimSpace)
	if err != nil {
		return nil, err
	}
	listenMode, err := resolveAliasedSetting(v, listenModeKey, legacyListenModeKey, func(raw string) string {
		return strings.ToLower(strings.TrimSpace(raw))
	})
	if err != nil {
		return nil, err
	}
	serviceName, err := resolveAliasedSetting(v, serviceNameKey, legacyServiceNameKey, strings.TrimSpace)
	if err != nil {
		return nil, err
	}

	if deviceName == "" {
		deviceName = "tgate"
	}
	if listenMode == "" {
		listenMode = TSNetListenModeListener
	}
	if serviceName == "" {
		serviceName = "svc:tgate"
	}

	funnelAllowlist, err := parseFunnelAllowlist(normalizeList(v.Get("funnel-allowlist")))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Port:             port,
		TailscaleName:    deviceName,
		Funnel:           v.GetBool("funnel"),
		FunnelAllowlist:  funnelAllowlist,
		Verbose:          v.GetBool("verbose"),
		JSON:             v.GetBool("json"),
		LogFile:          v.GetString("log-file"),
		AuthKey:          v.GetString("auth-key"),
		ForceTsnet:       v.GetBool("force-tsnet"),
		SetPath:          v.GetString("set-path"),
		ServePort:        v.GetInt("serve-port"),
		UseHTTPS:         v.GetBool("use-https"),
		NoTUI:            v.GetBool("no-tui"),
		NoUI:             v.GetBool("no-ui"),
		UIPort:           v.GetInt("ui-port"),
		Version:          v.GetBool("version"),
		Mock:             v.GetBool("mock"),
		CleanupServe:     v.GetBool("cleanup-serve"),
		TSNetListenMode:  listenMode,
		TSNetServiceName: serviceName,
	}

	// Handle version flag
	if cfg.Version {
		return cfg, nil
	}

	// Handle cleanup flag
	if cfg.CleanupServe {
		return cfg, nil
	}

	// Validate arguments
	if cfg.Mock && cfg.Port != 0 {
		return nil, fmt.Errorf("cannot specify both port and --mock flag%s", usageSuffix)
	}

	if !cfg.Mock && cfg.Port == 0 {
		return nil, fmt.Errorf("port argument is required (or use --mock for testing mode)%s", usageSuffix)
	}

	if cfg.Port < 0 {
		return nil, fmt.Errorf("port must be a positive integer")
	}

	// Auto-configure options
	cfg.applyAutoConfiguration()
	if err := cfg.validateTSNetServiceConfig(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// applyAutoConfiguration applies automatic configuration rules
func (c *Config) applyAutoConfiguration() {
	// Auto-enable funnel for mock mode unless explicitly disabled
	if c.Mock && !c.Funnel {
		c.Funnel = true
		c.UseHTTPS = true
	}

	// If funnel is enabled, automatically enable HTTPS since funnel requires it
	if c.Funnel {
		c.UseHTTPS = true
	}
}

// GetSetPath returns the mount path with default fallback
func (c *Config) GetSetPath() string {
	if c.SetPath == "" {
		return "/"
	}
	return c.SetPath
}

// GetServePort returns the serve port with protocol-based defaults
func (c *Config) GetServePort() int {
	if c.ServePort == 0 {
		if c.UseHTTPS {
			return 443
		}
		return 80
	}
	return c.ServePort
}

// HasFunnelAllowlist reports whether Funnel allowlist enforcement is active.
func (c *Config) HasFunnelAllowlist() bool {
	return c.Funnel && len(c.FunnelAllowlist) > 0
}

// UseFunnelProxyProtocol reports whether Funnel traffic should use PROXY v2.
// We only enable this for root-path serving because TCP forwarding does not
// support mount-point routing semantics from serve web handlers.
func (c *Config) UseFunnelProxyProtocol() bool {
	return c.HasFunnelAllowlist() && c.GetSetPath() == "/"
}

// EffectiveTSNetListenMode returns the runtime tsnet listen mode once
// compatibility fallbacks are applied.
func (c *Config) EffectiveTSNetListenMode() string {
	mode := c.TSNetListenMode
	if mode == "" {
		mode = TSNetListenModeListener
	}
	return mode
}

const usageSuffix = "\nUsage: tgate <port> [flags]     (proxy mode)\n       tgate --mock [flags]     (mock/testing mode)\n       tgate --version\n       tgate --cleanup-serve"

type parseState struct {
	port    int
	portSet bool
}

func configureViper(v *viper.Viper) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".tgate", "config.yml")
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("TGATE")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()
	v.SetDefault("funnel-allowlist", []string{})

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil
		}
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	return nil
}

func newRootCommand(v *viper.Viper, state *parseState) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "tgate [port]",
		Short: "Expose local services over Tailscale",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}

			port, err := strconv.Atoi(args[0])
			if err != nil || port <= 0 {
				return fmt.Errorf("invalid port %q: must be a positive integer", args[0])
			}

			state.port = port
			state.portSet = true
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringP(deviceNameKey, "n", "", "Tailscale device name (only used with tsnet mode) (default: tgate)")
	flags.String(legacyTailscaleNameKey, "", "Deprecated alias for --device-name")
	flags.BoolP("funnel", "f", false, "Enable Tailscale funnel (public internet access)")
	flags.BoolP("verbose", "v", false, "Enable verbose logging")
	flags.BoolP("json", "j", false, "Output logs in JSON format")
	flags.String("log-file", "", "Log file path (optional)")
	flags.String("auth-key", "", "Tailscale auth key to create separate tsnet device")
	flags.Bool("force-tsnet", false, "Force tsnet mode even if local Tailscale is available")
	flags.String("set-path", "", "Set custom path for serve (default: /)")
	flags.Int("serve-port", 0, "Tailscale serve port (default: 80 for HTTP, 443 for HTTPS)")
	flags.Bool("use-https", false, "Use HTTPS instead of HTTP for Tailscale serve")
	flags.Bool("no-tui", false, "Disable TUI and use simple console output")
	flags.Bool("no-ui", false, "Disable web UI dashboard")
	flags.Int("ui-port", 0, "Custom port for web UI (default: 4040 or next available)")
	flags.Bool("version", false, "Show version information")
	flags.BoolP("mock", "m", false, "Enable mock/testing mode (no backing server required, enables funnel by default)")
	flags.Bool("cleanup-serve", false, "Clear all Tailscale serve configurations and exit")
	flags.String(listenModeKey, "", "Listen mode: listener or service (default: listener)")
	flags.String(serviceNameKey, "", "Service name used when listen-mode=service (default: svc:tgate)")
	flags.String(legacyListenModeKey, "", "Deprecated alias for --listen-mode")
	flags.String(legacyServiceNameKey, "", "Deprecated alias for --service-name")
	_ = flags.MarkDeprecated(legacyTailscaleNameKey, "use --device-name instead")
	_ = flags.MarkDeprecated(legacyListenModeKey, "use --listen-mode instead")
	_ = flags.MarkDeprecated(legacyServiceNameKey, "use --service-name instead")
	_ = flags.MarkHidden(legacyTailscaleNameKey)
	_ = flags.MarkHidden(legacyListenModeKey)
	_ = flags.MarkHidden(legacyServiceNameKey)

	keys := []string{
		"port",
		deviceNameKey,
		legacyTailscaleNameKey,
		"funnel",
		"funnel-allowlist",
		"verbose",
		"json",
		"log-file",
		"auth-key",
		"force-tsnet",
		"set-path",
		"serve-port",
		"use-https",
		"no-tui",
		"no-ui",
		"ui-port",
		"version",
		"mock",
		"cleanup-serve",
		listenModeKey,
		serviceNameKey,
		legacyListenModeKey,
		legacyServiceNameKey,
	}

	for _, key := range keys {
		if flag := flags.Lookup(key); flag != nil {
			if err := v.BindPFlag(key, flag); err != nil {
				return nil, fmt.Errorf("failed to bind flag %s: %w", key, err)
			}
		}
		if err := v.BindEnv(key); err != nil {
			return nil, fmt.Errorf("failed to bind env for %s: %w", key, err)
		}
	}

	return cmd, nil
}

func helpRequested(cmd *cobra.Command, args []string) bool {
	help, err := cmd.Flags().GetBool("help")
	if err == nil && help {
		return true
	}

	return len(args) > 0 && args[0] == "help"
}

func normalizeList(value any) []string {
	var values []string

	switch v := value.(type) {
	case nil:
		return nil
	case string:
		values = strings.Split(v, ",")
	case []string:
		values = append(values, v...)
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				values = append(values, s)
			}
		}
	}

	normalized := make([]string, 0, len(values))
	for _, value := range values {
		entry := strings.TrimSpace(value)
		if entry == "" {
			continue
		}
		normalized = append(normalized, entry)
	}

	return normalized
}

func parseFunnelAllowlist(entries []string) ([]netip.Prefix, error) {
	parsed := make([]netip.Prefix, 0, len(entries))
	for _, entry := range entries {
		prefix, err := parseAllowlistEntry(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid funnel allowlist entry %q: must be an IP address or CIDR block", entry)
		}
		parsed = append(parsed, prefix)
	}
	return parsed, nil
}

func parseAllowlistEntry(entry string) (netip.Prefix, error) {
	if strings.Contains(entry, "/") {
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			return netip.Prefix{}, err
		}
		return prefix.Masked(), nil
	}

	addr, err := netip.ParseAddr(entry)
	if err != nil {
		return netip.Prefix{}, err
	}
	addr = addr.Unmap()

	if addr.Is4() {
		return netip.PrefixFrom(addr, 32), nil
	}
	return netip.PrefixFrom(addr, 128), nil
}

func resolveAliasedSetting(v *viper.Viper, canonicalKey, legacyKey string, normalize func(string) string) (string, error) {
	canonical := normalize(v.GetString(canonicalKey))
	legacy := normalize(v.GetString(legacyKey))

	if canonical != "" && legacy != "" && canonical != legacy {
		return "", fmt.Errorf("conflicting configuration: %s=%q conflicts with %s=%q", canonicalKey, canonical, legacyKey, legacy)
	}
	if canonical != "" {
		return canonical, nil
	}
	return legacy, nil
}

func (c *Config) validateTSNetServiceConfig() error {
	switch c.TSNetListenMode {
	case "", TSNetListenModeListener:
		c.TSNetListenMode = TSNetListenModeListener
		return nil
	case TSNetListenModeService:
		if c.Funnel {
			return fmt.Errorf("unsupported operating mode combination: listen-mode=service cannot be combined with funnel=true")
		}
		if c.TSNetServiceName == "" {
			return fmt.Errorf("service-name is required when listen-mode=service")
		}
		if err := tailcfg.ServiceName(c.TSNetServiceName).Validate(); err != nil {
			return fmt.Errorf("invalid service-name %q: %w", c.TSNetServiceName, err)
		}
		return nil
	default:
		return fmt.Errorf("invalid listen-mode %q: must be %q or %q", c.TSNetListenMode, TSNetListenModeListener, TSNetListenModeService)
	}
}
