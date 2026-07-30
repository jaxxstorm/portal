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
	AuthModeNone            = "none"
	AuthModeBasic           = "basic"
	AuthModeTailscale       = "tailscale"

	deviceNameKey          = "device-name"
	legacyTailscaleNameKey = "tailscale-name"
	listenModeKey          = "listen-mode"
	legacyListenModeKey    = "tsnet-listen-mode"
	serviceNameKey         = "service-name"
	legacyServiceNameKey   = "tsnet-service-name"
)

// Config holds the parsed and validated configuration
type Config struct {
	Port                       int
	TailscaleName              string
	Funnel                     bool
	FunnelAllowlist            []netip.Prefix
	Verbose                    bool
	JSON                       bool
	LogFile                    string
	AuthKey                    string
	ForceTsnet                 bool
	SetPath                    string
	ServePort                  int
	UseHTTPS                   bool
	NoTUI                      bool
	NoUI                       bool
	UIPort                     int
	Version                    bool
	Mock                       bool
	CleanupServe               bool
	TSNetListenMode            string
	TSNetServiceName           string
	AuthMode                   string
	AuthUsername               string
	AuthPassword               string
	TailscaleOAuthClientID     string
	TailscaleOAuthClientSecret string
	TailscaleSessionSecret     string
	ConfigInit                 bool
	ConfigInitForce            bool
	OAuthAppArgs               []string
}

// Parse parses command line arguments and returns a validated configuration
func Parse() (*Config, error) {
	return ParseArgs(os.Args[1:])
}

// ParseArgs parses command line arguments and returns a validated configuration.
// Exposed for tests.
func ParseArgs(args []string) (*Config, error) {
	if commandHelpRequested(args) {
		cmd := newCommandHelpRoot()
		cmd.SetArgs(args)
		_ = cmd.Execute()
		return nil, pflag.ErrHelp
	}
	args, command, err := normalizeCommandArgs(args)
	if err != nil {
		return nil, err
	}
	if command.configInit {
		return &Config{ConfigInit: true, ConfigInitForce: command.configInitForce}, nil
	}
	if command.cleanup {
		return &Config{CleanupServe: true}, nil
	}
	if command.oauthApp {
		return &Config{OAuthAppArgs: command.oauthAppArgs}, nil
	}

	v := viper.New()
	if err := configureViper(v); err != nil {
		return nil, err
	}

	state := &parseState{}
	cmd, err := newRootCommand(v, state, command)
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
		deviceName = "portal"
	}
	if listenMode == "" {
		listenMode = TSNetListenModeListener
	}
	if serviceName == "" {
		serviceName = "svc:portal"
	}
	authMode := strings.ToLower(strings.TrimSpace(v.GetString("auth-mode")))
	if authMode == "" {
		authMode = AuthModeNone
	}

	funnelAllowlist, err := parseFunnelAllowlist(normalizeList(v.Get("funnel-allowlist")))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Port:                       port,
		TailscaleName:              deviceName,
		Funnel:                     v.GetBool("funnel"),
		FunnelAllowlist:            funnelAllowlist,
		Verbose:                    v.GetBool("verbose"),
		JSON:                       v.GetBool("json"),
		LogFile:                    v.GetString("log-file"),
		AuthKey:                    v.GetString("auth-key"),
		ForceTsnet:                 v.GetBool("force-tsnet"),
		SetPath:                    v.GetString("set-path"),
		ServePort:                  v.GetInt("serve-port"),
		UseHTTPS:                   v.GetBool("use-https"),
		NoTUI:                      v.GetBool("no-tui"),
		NoUI:                       v.GetBool("no-ui"),
		UIPort:                     v.GetInt("ui-port"),
		Version:                    v.GetBool("version"),
		Mock:                       v.GetBool("mock"),
		CleanupServe:               v.GetBool("cleanup-serve"),
		TSNetListenMode:            listenMode,
		TSNetServiceName:           serviceName,
		AuthMode:                   authMode,
		AuthUsername:               v.GetString("auth-username"),
		AuthPassword:               v.GetString("auth-password"),
		TailscaleOAuthClientID:     v.GetString("tailscale-oauth-client-id"),
		TailscaleOAuthClientSecret: v.GetString("tailscale-oauth-client-secret"),
		TailscaleSessionSecret:     v.GetString("tailscale-session-secret"),
	}
	if command.node {
		cfg.ForceTsnet = true
	}
	if command.mock {
		cfg.Mock = true
		cfg.Port = 0
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
	if err := cfg.validateAuthorizationConfig(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func commandHelpRequested(args []string) bool {
	if len(args) == 1 {
		return args[0] == "--help" || args[0] == "help"
	}
	if len(args) == 2 {
		return args[1] == "--help" || args[1] == "-h"
	}
	if len(args) == 3 {
		isHelp := args[2] == "--help" || args[2] == "-h"
		return (args[0] == "config" && args[1] == "init" && isHelp) ||
			((args[0] == "serve" || args[0] == "mock") && args[1] == "node" && isHelp)
	}
	return false
}

func newCommandHelpRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "portal",
		Short: "Expose local services over Tailscale",
		Long:  "Portal exposes local services over Tailscale. Tailnet-private exposure is the default; Funnel is explicitly public.",
	}
	root.PersistentFlags().Bool("version", false, "Show version information")
	serveCmd := newCommandHelpServe("serve [port]", "Serve a local service through the local Tailscale daemon", false)
	serveCmd.AddCommand(newCommandHelpServe("node [port]", "Serve a local service through Portal's isolated Tailscale node", true))
	root.AddCommand(serveCmd)
	mockCmd := newCommandHelpServe("mock", "Serve Portal's mock backend through the local Tailscale daemon", false)
	mockCmd.AddCommand(newCommandHelpServe("node", "Serve Portal's mock backend through an isolated Tailscale node", true))
	root.AddCommand(mockCmd)
	root.AddCommand(&cobra.Command{Use: "cleanup", Short: "Clear this node's Tailscale Serve configuration", Run: func(*cobra.Command, []string) {}})
	configCmd := &cobra.Command{Use: "config", Short: "Manage Portal configuration"}
	initCmd := &cobra.Command{Use: "init", Short: "Create a configuration template", Run: func(*cobra.Command, []string) {}}
	initCmd.Flags().Bool("force", false, "Replace an existing configuration file")
	configCmd.AddCommand(initCmd)
	root.AddCommand(configCmd)
	root.AddCommand(&cobra.Command{Use: "oauth-app", Short: "Create and configure Tailscale OAuth Apps", Run: func(*cobra.Command, []string) {}})
	return root
}

func newCommandHelpServe(use, short string, node bool) *cobra.Command {
	long := short + ".\n\nConfigure advanced UI, path, port, and credential settings in ~/.portal/config.yml or PORTAL_* environment variables."
	cmd := &cobra.Command{Use: use, Short: short, Long: long, Run: func(*cobra.Command, []string) {}}
	flags := cmd.Flags()
	flags.BoolP("funnel", "f", false, "Enable Tailscale Funnel (public internet access)")
	flags.String("auth-mode", AuthModeNone, "Endpoint authorization mode: none, basic, or tailscale")
	if node {
		flags.StringP("device-name", "n", "portal", "Name of Portal's isolated Tailscale node")
		flags.String("auth-key", "", "Tailscale auth key for the isolated node")
	}
	flags.BoolP("verbose", "v", false, "Enable verbose logging")
	flags.BoolP("json", "j", false, "Output logs in JSON format")
	flags.String("log-file", "", "Log file path")
	flags.Bool("no-tui", false, "Disable the terminal UI")
	flags.Bool("no-ui", false, "Disable the web UI dashboard")
	return cmd
}

// applyAutoConfiguration applies automatic configuration rules
func (c *Config) applyAutoConfiguration() {
	// Funnel and browser login both require HTTPS for secure cookies and callbacks.
	if c.Funnel || c.AuthMode == AuthModeTailscale {
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
		// In service mode, default to the target service port so the exposed
		// service endpoint aligns with operator-defined service endpoints.
		if c.EffectiveTSNetListenMode() == TSNetListenModeService && c.Port > 0 {
			return c.Port
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

// IsServiceMode reports whether runtime listen mode resolves to service mode.
func (c *Config) IsServiceMode() bool {
	return c.EffectiveTSNetListenMode() == TSNetListenModeService
}

const usageSuffix = "\nUsage: portal <port> [flags]     (proxy mode)\n       portal --mock [flags]     (mock/testing mode)\n       portal --version\n       portal --cleanup-serve"

type parseState struct {
	port    int
	portSet bool
}

type commandArgs struct {
	serve           bool
	mock            bool
	node            bool
	cleanup         bool
	configInit      bool
	configInitForce bool
	oauthApp        bool
	oauthAppArgs    []string
}

func normalizeCommandArgs(args []string) ([]string, commandArgs, error) {
	if len(args) == 0 {
		return args, commandArgs{}, nil
	}

	switch args[0] {
	case "serve":
		args, node := normalizeNodeArgs(args[1:])
		return args, commandArgs{serve: true, node: node}, nil
	case "mock":
		args, node := normalizeNodeArgs(args[1:])
		return append([]string{"--mock"}, args...), commandArgs{mock: true, node: node}, nil
	case "cleanup":
		if len(args) != 1 {
			return nil, commandArgs{}, fmt.Errorf("cleanup does not accept arguments")
		}
		return nil, commandArgs{cleanup: true}, nil
	case "oauth-app":
		return nil, commandArgs{oauthApp: true, oauthAppArgs: args}, nil
	case "config":
		if len(args) < 2 || args[1] != "init" {
			return nil, commandArgs{}, fmt.Errorf("usage: portal config init [--force]")
		}
		if len(args) > 3 || (len(args) == 3 && args[2] != "--force") {
			return nil, commandArgs{}, fmt.Errorf("usage: portal config init [--force]")
		}
		return nil, commandArgs{configInit: true, configInitForce: len(args) == 3}, nil
	default:
		return args, commandArgs{}, nil
	}
}

func normalizeNodeArgs(args []string) ([]string, bool) {
	if len(args) > 0 && args[0] == "node" {
		return args[1:], true
	}
	return args, false
}

// DefaultConfigPath returns Portal's default configuration file location.
func DefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	return filepath.Join(homeDir, ".portal", "config.yml"), nil
}

const configTemplate = `# Portal configuration. CLI flags override environment values, which override this file.
# Keep secrets in a secret manager or PORTAL_* environment variables.

# Local service port for "portal serve". Remove or change as needed.
# Plain serve and mock commands require the local tailscaled daemon.
port: 8080

# Tailnet-private by default. Set true only to enable public Tailscale Funnel.
funnel: false

# Optional Funnel source allowlist.
funnel-allowlist: []

# Tailscale serving settings. Use "portal serve node" or "portal mock node"
# to create an isolated Tailscale node with these settings.
device-name: portal
listen-mode: listener
service-name: svc:portal
# auth-key: ""
# serve-port: 0
# set-path: /
# use-https: false

# Runtime output settings.
# no-tui: false
# no-ui: false
# ui-port: 0
# json: false
# verbose: false

# Endpoint authorization. Use environment variables for credentials.
auth-mode: none
# auth-username: ""
# auth-password: ""
# tailscale-oauth-client-id: ""
# tailscale-oauth-client-secret: ""
# tailscale-session-secret: ""
`

// WriteConfigTemplate creates the default configuration template without
// persisting credentials.
func WriteConfigTemplate(force bool) (string, error) {
	configPath, err := DefaultConfigPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return "", fmt.Errorf("creating configuration directory: %w", err)
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if force {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	file, err := os.OpenFile(configPath, flags, 0o600)
	if err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("configuration file %s already exists; rerun with --force to replace it", configPath)
		}
		return "", fmt.Errorf("creating configuration template: %w", err)
	}
	defer file.Close()
	if _, err := file.WriteString(configTemplate); err != nil {
		return "", fmt.Errorf("writing configuration template: %w", err)
	}
	return configPath, nil
}

func configureViper(v *viper.Viper) error {
	configPath, err := DefaultConfigPath()
	if err != nil {
		return err
	}
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("PORTAL")
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

func newRootCommand(v *viper.Viper, state *parseState, command commandArgs) (*cobra.Command, error) {
	use := "portal [port]"
	short := "Expose local services over Tailscale"
	if command.serve {
		use = "portal serve [port]"
		short = "Serve a local service over Tailscale"
	}
	if command.mock {
		use = "portal mock"
		short = "Serve Portal's mock backend over Tailscale"
	}
	long := "Portal exposes local services over Tailscale.\n\nCommands:\n  portal serve [port]   Serve a local service\n  portal mock           Serve the mock backend\n  portal oauth-app      Create a Tailscale OAuth App\n  portal cleanup        Clear Tailscale Serve configuration\n  portal config init    Create a configuration template"
	if command.serve || command.mock {
		long = short + ".\n\nCommon runtime flags are shown below. Configure advanced tsnet, UI, path, port, and credential settings in ~/.portal/config.yml or PORTAL_* environment variables."
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
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
	flags.StringP(deviceNameKey, "n", "", "Tailscale device name (only used with tsnet mode) (default: portal)")
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
	flags.BoolP("mock", "m", false, "Enable mock/testing mode (no backing server required)")
	flags.Bool("cleanup-serve", false, "Clear all Tailscale serve configurations and exit")
	flags.String(listenModeKey, "", "Listen mode: listener or service (default: listener; service mode requires tag-based identity)")
	flags.String(serviceNameKey, "", "Service name used when listen-mode=service (default: svc:portal; requires tagged host identity)")
	flags.String(legacyListenModeKey, "", "Deprecated alias for --listen-mode")
	flags.String(legacyServiceNameKey, "", "Deprecated alias for --service-name")
	flags.String("auth-mode", AuthModeNone, "Endpoint authorization mode: none, basic, or tailscale")
	flags.String("auth-username", "", "HTTP Basic authorization username")
	flags.String("auth-password", "", "HTTP Basic authorization password")
	flags.String("tailscale-oauth-client-id", "", "Tailscale OAuth App client ID")
	flags.String("tailscale-oauth-client-secret", "", "Tailscale OAuth App client secret")
	flags.String("tailscale-session-secret", "", "Secret used to sign Tailscale login sessions")
	if command.serve || command.mock {
		_ = flags.MarkHidden("cleanup-serve")
	}
	if command.serve || command.mock {
		_ = flags.MarkHidden("mock")
		for _, name := range []string{
			"auth-key",
			deviceNameKey,
			listenModeKey,
			serviceNameKey,
			"set-path",
			"serve-port",
			"use-https",
			"ui-port",
			"auth-username",
			"auth-password",
			"tailscale-oauth-client-id",
			"tailscale-oauth-client-secret",
			"tailscale-session-secret",
		} {
			_ = flags.MarkHidden(name)
		}
	}
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
		"auth-mode",
		"auth-username",
		"auth-password",
		"tailscale-oauth-client-id",
		"tailscale-oauth-client-secret",
		"tailscale-session-secret",
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

func (c *Config) validateAuthorizationConfig() error {
	hasBasicCredentials := strings.TrimSpace(c.AuthUsername) != "" || c.AuthPassword != ""
	hasTailscaleCredentials := strings.TrimSpace(c.TailscaleOAuthClientID) != "" || c.TailscaleOAuthClientSecret != "" || c.TailscaleSessionSecret != ""

	switch c.AuthMode {
	case AuthModeNone:
		if hasBasicCredentials {
			return fmt.Errorf("auth-username and auth-password require auth-mode=basic")
		}
		if hasTailscaleCredentials {
			return fmt.Errorf("Tailscale OAuth credentials require auth-mode=tailscale")
		}
		return nil
	case AuthModeBasic:
		if hasTailscaleCredentials {
			return fmt.Errorf("Tailscale OAuth credentials cannot be combined with auth-mode=basic")
		}
		if strings.TrimSpace(c.AuthUsername) == "" {
			return fmt.Errorf("auth-username is required when auth-mode=basic")
		}
		if c.AuthPassword == "" {
			return fmt.Errorf("auth-password is required when auth-mode=basic")
		}
		return nil
	case AuthModeTailscale:
		if hasBasicCredentials {
			return fmt.Errorf("auth-username and auth-password cannot be combined with auth-mode=tailscale")
		}
		if strings.TrimSpace(c.TailscaleOAuthClientID) == "" {
			return fmt.Errorf("tailscale-oauth-client-id is required when auth-mode=tailscale")
		}
		if c.TailscaleOAuthClientSecret == "" {
			return fmt.Errorf("tailscale-oauth-client-secret is required when auth-mode=tailscale")
		}
		if c.TailscaleSessionSecret == "" {
			return fmt.Errorf("tailscale-session-secret is required when auth-mode=tailscale")
		}
		return nil
	default:
		return fmt.Errorf("invalid auth-mode %q: must be %q, %q, or %q", c.AuthMode, AuthModeNone, AuthModeBasic, AuthModeTailscale)
	}
}
