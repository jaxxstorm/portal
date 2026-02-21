// internal/tailscale/client.go
package tailscale

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"slices"
	"strings"

	"go.uber.org/zap"
	"tailscale.com/client/local"
	"tailscale.com/ipn"
	"tailscale.com/tailcfg"

	"github.com/jaxxstorm/tgate/internal/logging"
)

// Config holds configuration for Tailscale setup
type Config struct {
	MountPath           string
	EnableFunnel        bool
	EnableProxyProtocol bool
	UseHTTPS            bool
	ServePort           int
	ProxyPort           int
	ListenMode          string
	ServiceName         string
}

// ServiceInfo holds information about the configured service
type ServiceInfo struct {
	URL       string // The Tailscale URL where the service is accessible
	LocalURL  string // Local proxy URL (for development/testing)
	DNSName   string // Tailscale DNS name
	ServePort int    // Port used by Tailscale serve
	ProxyPort int    // Local proxy port
	IsFunnel  bool   // Whether funnel is enabled (internet accessible)
	IsHTTPS   bool   // Whether HTTPS is enabled
	MountPath string // Mount path for the service
}

// Client wraps the Tailscale local client with additional functionality
type Client struct {
	lc     *local.Client
	logger *zap.Logger
}

// NewClient creates a new Tailscale client with structured logging
func NewClient(logger *zap.Logger) *Client {
	return &Client{
		lc:     &local.Client{},
		logger: logger,
	}
}

// IsAvailable checks if local Tailscale is available
func (c *Client) IsAvailable(ctx context.Context) bool {
	c.logger.Info(logging.MsgTailscaleAvailability,
		logging.Component("tailscale_client"),
	)

	_, err := c.lc.Status(ctx)
	available := err == nil

	if available {
		c.logger.Info(logging.MsgTailscaleDaemonReady,
			logging.Status("ready"),
		)
	} else {
		c.logger.Info(logging.MsgTailscaleDaemonMissing,
			logging.Status("not_found"),
			logging.Error(err),
		)
	}

	return available
}

// GetDNSName returns the DNS name of this Tailscale node
func (c *Client) GetDNSName(ctx context.Context) (string, error) {
	c.logger.Debug("Getting Tailscale DNS name",
		logging.Component("tailscale_client"),
		logging.Operation("get_dns_name"),
	)

	status, err := c.lc.Status(ctx)
	if err != nil {
		c.logger.Error("Failed to get Tailscale status",
			logging.Component("tailscale_client"),
			logging.Operation("get_dns_name"),
			logging.Error(err),
		)
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	dnsName := strings.TrimSuffix(status.Self.DNSName, ".")
	c.logger.Debug("Retrieved DNS name",
		logging.Component("tailscale_client"),
		zap.String("dns_name", dnsName),
	)

	return dnsName, nil
}

// SetupServe configures Tailscale serve for the given configuration
func (c *Client) SetupServe(ctx context.Context, config Config) (*ServiceInfo, error) {
	// Use shorter log messages to avoid TUI truncation issues
	c.logger.Info("Tailscale serve setup starting",
		logging.Component("tailscale_serve"),
		logging.ServePort(config.ServePort),
		logging.ProxyPort(config.ProxyPort),
	)

	// Get current serve config
	sc, err := c.lc.GetServeConfig(ctx)
	if err != nil {
		c.logger.Error("Failed to get serve config",
			logging.Component("tailscale_serve"),
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to get serve config: %w", err)
	}
	if sc == nil {
		sc = new(ipn.ServeConfig)
	}

	status, err := c.lc.Status(ctx)
	if err != nil {
		c.logger.Error("Failed to get Tailscale status",
			logging.Component("tailscale_serve"),
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	dnsName := strings.TrimSuffix(status.Self.DNSName, ".")
	magicDNSSuffix := status.MagicDNSSuffix
	if status.CurrentTailnet != nil && strings.TrimSpace(status.CurrentTailnet.MagicDNSSuffix) != "" {
		magicDNSSuffix = status.CurrentTailnet.MagicDNSSuffix
	}
	listenMode := normalizeServeListenMode(config.ListenMode)
	serviceName := strings.TrimSpace(config.ServiceName)

	// Set up HTTP handler for the proxy target
	h := &ipn.HTTPHandler{
		Proxy: fmt.Sprintf("http://localhost:%d", config.ProxyPort),
	}

	// Clean mount path
	mountPath := config.MountPath
	if mountPath == "" {
		mountPath = "/"
	}
	if !strings.HasPrefix(mountPath, "/") {
		mountPath = "/" + mountPath
	}

	// Determine serve port and TLS usage
	var srvPort uint16
	var useTLS bool

	if config.ServePort != 0 {
		srvPort = uint16(config.ServePort)
		useTLS = config.UseHTTPS || config.ServePort == 443
	} else {
		if config.UseHTTPS {
			srvPort = 443
			useTLS = true
		} else {
			srvPort = 80
			useTLS = false
		}
	}

	c.logger.Info("Tailscale serve configuration",
		logging.Component("tailscale_serve"),
		logging.ServePort(int(srvPort)),
		zap.Bool("use_tls", useTLS),
	)

	var serviceNameTag tailcfg.ServiceName
	if listenMode == TSNetListenModeService {
		serviceNameTag = tailcfg.ServiceName(serviceName)
		if err := serviceNameTag.Validate(); err != nil {
			return nil, fmt.Errorf("invalid service name %q: %w", serviceName, err)
		}
		if strings.TrimSpace(magicDNSSuffix) == "" {
			return nil, fmt.Errorf("cannot configure service %q: missing tailnet MagicDNS suffix", serviceName)
		}
	}

	// Check if port is already in use in node-level or service-level config.
	portInUse := sc.IsTCPForwardingOnPort(srvPort, "") || sc.IsServingWeb(srvPort, "")
	if serviceNameTag != "" {
		portInUse = portInUse || sc.IsTCPForwardingOnPort(srvPort, serviceNameTag) || sc.IsServingWeb(srvPort, serviceNameTag)
	}
	if portInUse {
		c.logger.Error("Port already in use for serve",
			logging.Component("tailscale_serve"),
			logging.ServePort(int(srvPort)),
			zap.String("listen_mode", listenMode),
			zap.String("service_name", serviceName),
		)
		return nil, fmt.Errorf("port %d is already in use by tailscale serve", srvPort)
	}

	useFunnelProxyProtocol := config.EnableFunnel && config.EnableProxyProtocol

	if listenMode == TSNetListenModeService {
		if config.EnableFunnel {
			return nil, fmt.Errorf("service mode is mutually exclusive with funnel")
		}
		sc.SetWebHandler(h, serviceName, srvPort, mountPath, useTLS, magicDNSSuffix)
	} else if useFunnelProxyProtocol {
		c.logger.Info("Setting up TLS-terminated TCP forwarding with PROXY protocol v2",
			logging.Component("tailscale_serve"),
			logging.ServePort(int(srvPort)),
		)
		sc.SetTCPForwarding(srvPort, fmt.Sprintf("127.0.0.1:%d", config.ProxyPort), true, 2, dnsName)
	} else {
		// Set web handler
		sc.SetWebHandler(h, dnsName, srvPort, mountPath, useTLS, "")
	}

	// If using HTTPS/TLS, verify HTTPS feature support
	if useTLS {
		if err := c.enableHTTPSFeature(ctx); err != nil {
			c.logger.Warn("HTTPS feature check failed",
				logging.Component("tailscale_serve"),
				logging.Error(err),
				logging.Status("https_may_not_work"),
			)
		}
	}

	// Enable funnel if requested (only works with HTTPS/443)
	if config.EnableFunnel {
		if !useTLS || srvPort != 443 {
			c.logger.Error("Funnel configuration invalid",
				logging.Component("tailscale_serve"),
				logging.ServePort(int(srvPort)),
				zap.Bool("use_tls", useTLS),
				logging.Status("funnel_requires_https_443"),
			)
			return nil, fmt.Errorf("funnel requires HTTPS on port 443")
		}

		// Enable HTTPS feature first if needed
		if err := c.enableHTTPSFeature(ctx); err != nil {
			c.logger.Error("Failed to enable HTTPS feature for funnel",
				logging.Component("tailscale_serve"),
				logging.Error(err),
				logging.Status("funnel_setup_failed"),
			)
			return nil, fmt.Errorf("HTTPS certificates not enabled: %w", err)
		}

		// Check certificate status before enabling funnel
		c.logger.Info("Checking HTTPS certificate status for funnel",
			logging.Component("tailscale_serve"),
			zap.String("dns_name", dnsName),
		)

		if err := c.checkTailscaleCertificates(ctx, dnsName); err != nil {
			c.logger.Error("Certificate check failed for funnel",
				logging.Component("tailscale_serve"),
				zap.String("dns_name", dnsName),
				logging.Error(err),
				logging.Status("funnel_cert_check_failed"),
			)
			return nil, fmt.Errorf("cannot enable funnel: %w", err)
		}

		sc.SetFunnel(dnsName, srvPort, true)
		c.logger.Info("Funnel enabled successfully",
			logging.Component("tailscale_serve"),
			logging.Status("internet_accessible"),
		)
	}

	if listenMode == TSNetListenModeService {
		if err := c.ensureServiceAdvertised(ctx, serviceNameTag); err != nil {
			c.logger.Error("Failed to advertise Tailscale service",
				logging.Component("tailscale_serve"),
				zap.String("service_name", serviceName),
				logging.Error(err),
			)
			return nil, err
		}
	}

	// Apply the serve config
	err = c.lc.SetServeConfig(ctx, sc)
	if err != nil {
		c.logger.Error("Failed to apply serve config",
			logging.Component("tailscale_serve"),
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to set serve config: %w", err)
	}

	// Display URL information
	scheme := "http"
	if useTLS {
		scheme = "https"
	}

	portPart := ""
	if (scheme == "http" && srvPort != 80) || (scheme == "https" && srvPort != 443) {
		portPart = fmt.Sprintf(":%d", srvPort)
	}

	hostForURL := dnsName
	if listenMode == TSNetListenModeService {
		hostForURL = fmt.Sprintf("%s.%s", serviceNameTag.WithoutPrefix(), magicDNSSuffix)
	}
	url := fmt.Sprintf("%s://%s%s%s", scheme, hostForURL, portPart, mountPath)

	if config.EnableFunnel {
		c.logger.Info("Tailscale serve success - internet accessible",
			logging.Component("tailscale_serve"),
			logging.URL(url),
		)
	} else {
		c.logger.Info("Tailscale serve success - tailnet only",
			logging.Component("tailscale_serve"),
			logging.URL(url),
		)
	}

	// Create ServiceInfo to return
	serviceInfo := &ServiceInfo{
		URL:       url,
		LocalURL:  fmt.Sprintf("http://localhost:%d", config.ProxyPort),
		DNSName:   dnsName,
		ServePort: int(srvPort),
		ProxyPort: config.ProxyPort,
		IsFunnel:  config.EnableFunnel,
		IsHTTPS:   useTLS,
		MountPath: mountPath,
	}

	return serviceInfo, nil
}

func normalizeServeListenMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		return TSNetListenModeListener
	}
	return mode
}

func (c *Client) ensureServiceAdvertised(ctx context.Context, svcName tailcfg.ServiceName) error {
	prefs, err := c.lc.GetPrefs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current tailscale prefs: %w", err)
	}

	current := prefs.AdvertiseServices
	svc := svcName.String()
	if slices.Contains(current, svc) {
		return nil
	}

	updated := append(append([]string{}, current...), svc)
	_, err = c.lc.EditPrefs(ctx, &ipn.MaskedPrefs{
		AdvertiseServicesSet: true,
		Prefs: ipn.Prefs{
			AdvertiseServices: updated,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to advertise service %q: %w", svc, err)
	}

	c.logger.Info("Advertised Tailscale service",
		logging.Component("tailscale_serve"),
		zap.String("service_name", svc),
	)
	return nil
}

// SetupUIServe sets up Tailscale serve for the UI dashboard
func (c *Client) SetupUIServe(ctx context.Context, uiPort int) (uint16, string, error) {
	c.logger.Info("Setting up Tailscale UI serve",
		logging.Component("tailscale_ui_serve"),
		logging.UIPort(uiPort),
	)

	// Get current serve config
	sc, err := c.lc.GetServeConfig(ctx)
	if err != nil {
		c.logger.Error("Failed to get serve config for UI",
			logging.Component("tailscale_ui_serve"),
			logging.Error(err),
		)
		return 0, "", fmt.Errorf("failed to get serve config: %w", err)
	}
	if sc == nil {
		sc = new(ipn.ServeConfig)
	}

	// Get DNS name
	dnsName, err := c.GetDNSName(ctx)
	if err != nil {
		return 0, "", err
	}

	// Find available Tailscale port starting from a random port
	tailscalePort, err := c.findAvailableTailscalePort(sc, 8080)
	if err != nil {
		c.logger.Error("Failed to find available Tailscale port for UI",
			logging.Component("tailscale_ui_serve"),
			logging.Error(err),
		)
		return 0, "", fmt.Errorf("failed to find available Tailscale port: %w", err)
	}

	c.logger.Info("Allocated Tailscale port for UI",
		logging.Component("tailscale_ui_serve"),
		logging.TailscalePort(int(tailscalePort)),
		logging.UIPort(uiPort),
	)

	uiHandler := &ipn.HTTPHandler{
		Proxy: fmt.Sprintf("http://localhost:%d", uiPort),
	}

	sc.SetWebHandler(uiHandler, dnsName, tailscalePort, "/ui/", false, "") // HTTP only, no TLS

	// Apply the serve config
	err = c.lc.SetServeConfig(ctx, sc)
	if err != nil {
		c.logger.Error("Failed to apply UI serve config",
			logging.Component("tailscale_ui_serve"),
			logging.Error(err),
		)
		return 0, "", fmt.Errorf("failed to set UI serve config: %w", err)
	}

	uiURL := fmt.Sprintf("http://%s:%d/ui/", dnsName, tailscalePort)

	c.logger.Info("Web UI serve configured successfully",
		logging.Component("tailscale_ui_serve"),
		logging.URL(uiURL),
		logging.TailscalePort(int(tailscalePort)),
		logging.Status("tailnet_accessible"),
	)

	return tailscalePort, uiURL, nil
}

// Cleanup removes Tailscale serve configuration
func (c *Client) Cleanup(ctx context.Context, config Config) error {
	c.logger.Info(logging.MsgCleanupStarting,
		logging.Component("tailscale_serve"),
		logging.ProxyPort(config.ProxyPort),
	)

	sc, err := c.lc.GetServeConfig(ctx)
	if err != nil || sc == nil {
		c.logger.Debug("No serve config to clean up",
			logging.Component("tailscale_serve"),
			logging.Error(err),
		)
		return nil // Nothing to clean up
	}

	dnsName, err := c.GetDNSName(ctx)
	if err != nil {
		c.logger.Warn("Failed to get DNS name during cleanup",
			logging.Component("tailscale_serve"),
			logging.Error(err),
		)
		return err
	}

	// Create a completely empty serve config to clear everything
	emptyConfig := &ipn.ServeConfig{}

	c.logger.Info("Clearing all Tailscale serve configurations",
		logging.Component("tailscale_serve"),
		zap.String("dns_name", dnsName),
	)

	// Apply the empty config to clear all serve configurations
	err = c.lc.SetServeConfig(ctx, emptyConfig)
	if err != nil {
		c.logger.Warn("Failed to clear serve config",
			logging.Component("tailscale_serve"),
			logging.Error(err),
		)
		return err
	} else {
		c.logger.Info(logging.MsgCleanupComplete,
			logging.Component("tailscale_serve"),
		)
	}

	return nil
}

// CleanupAll removes all Tailscale serve configurations (not just from this session)
func (c *Client) CleanupAll(ctx context.Context) error {
	c.logger.Info("Clearing ALL Tailscale serve configurations",
		logging.Component("tailscale_serve"),
		logging.Operation("cleanup_all"),
	)

	// Create a completely empty serve config to clear everything
	emptyConfig := &ipn.ServeConfig{}

	// Apply the empty config to clear all serve configurations
	err := c.lc.SetServeConfig(ctx, emptyConfig)
	if err != nil {
		c.logger.Error("Failed to clear all serve configurations",
			logging.Component("tailscale_serve"),
			logging.Error(err),
		)
		return err
	}

	c.logger.Info("All Tailscale serve configurations cleared",
		logging.Component("tailscale_serve"),
		logging.Status("all_cleared"),
	)

	return nil
}

// enableHTTPSFeature enables HTTPS capability for the tailnet
func (c *Client) enableHTTPSFeature(ctx context.Context) error {
	c.logger.Debug("Checking HTTPS capability",
		logging.Component("tailscale_https"),
		logging.Operation("enable_https_feature"),
	)

	// Check if HTTPS is already enabled
	status, err := c.lc.Status(ctx)
	if err != nil {
		c.logger.Error("Failed to get status for HTTPS check",
			logging.Component("tailscale_https"),
			logging.Error(err),
		)
		return fmt.Errorf("failed to get status: %w", err)
	}

	if status.Self.HasCap(tailcfg.CapabilityHTTPS) {
		c.logger.Info("HTTPS capability already enabled",
			logging.Component("tailscale_https"),
			logging.Status("already_enabled"),
		)
		return nil
	}

	c.logger.Warn("HTTPS capability not enabled",
		logging.Component("tailscale_https"),
		logging.Status("requires_manual_setup"),
	)

	return fmt.Errorf("HTTPS capability needs to be enabled in your Tailscale admin console")
}

// checkTailscaleCertificates checks if HTTPS certificates are available
func (c *Client) checkTailscaleCertificates(ctx context.Context, dnsName string) error {
	c.logger.Debug("Checking Tailscale certificates",
		logging.Component("tailscale_https"),
		zap.String("dns_name", dnsName),
		logging.Operation("check_certificates"),
	)

	// Check if HTTPS certificates are available
	status, err := c.lc.Status(ctx)
	if err != nil {
		c.logger.Error("Failed to get status for certificate check",
			logging.Component("tailscale_https"),
			logging.Error(err),
		)
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Check if the node has HTTPS capability
	if !status.Self.HasCap(tailcfg.CapabilityHTTPS) {
		c.logger.Error("Node does not have HTTPS capability",
			logging.Component("tailscale_https"),
			logging.Status("https_not_enabled"),
		)
		return fmt.Errorf("HTTPS certificates not enabled for this tailnet")
	}

	c.logger.Info("HTTPS capability confirmed",
		logging.Component("tailscale_https"),
		logging.Status("capability_enabled"),
	)

	// Check certificate status
	if len(status.CertDomains) == 0 {
		c.logger.Warn("No certificate domains found",
			logging.Component("tailscale_https"),
			logging.Status("no_cert_domains"),
		)
		return fmt.Errorf("no certificate domains available")
	}

	// Check if our domain is in the cert domains
	found := false
	for _, domain := range status.CertDomains {
		if strings.Contains(domain, strings.Split(dnsName, ".")[0]) {
			found = true
			break
		}
	}

	if !found {
		c.logger.Warn("Certificate not found for domain",
			logging.Component("tailscale_https"),
			zap.String("dns_name", dnsName),
			zap.Strings("available_domains", status.CertDomains),
			logging.Status("cert_not_available"),
		)
		return fmt.Errorf("certificate not available for domain %s", dnsName)
	}

	c.logger.Info("Certificate confirmed for domain",
		logging.Component("tailscale_https"),
		zap.String("dns_name", dnsName),
		logging.Status("cert_available"),
	)
	return nil
}

// findAvailableTailscalePort finds an available port for Tailscale serve
func (c *Client) findAvailableTailscalePort(sc *ipn.ServeConfig, startPort uint16) (uint16, error) {
	// Use a random starting port to avoid conflicts
	randomOffset := rand.IntN(100) // Random offset 0-99
	actualStartPort := startPort + uint16(randomOffset)

	c.logger.Debug("Finding available Tailscale port",
		logging.Component("tailscale_port_allocation"),
		logging.StartPort(int(actualStartPort)),
	)

	for port := actualStartPort; port < actualStartPort+200; port++ {
		if !sc.IsTCPForwardingOnPort(port, "") {
			// Also check if web handlers exist on this port
			available := true
			if sc.Web != nil {
				for hostPort := range sc.Web {
					if strings.Contains(string(hostPort), fmt.Sprintf(":%d", port)) {
						available = false
						break
					}
				}
			}
			if available {
				c.logger.Debug("Found available Tailscale port",
					logging.Component("tailscale_port_allocation"),
					logging.Port(int(port)),
				)
				return port, nil
			}
		}
	}

	c.logger.Error("No available Tailscale port found",
		logging.Component("tailscale_port_allocation"),
		logging.StartPort(int(actualStartPort)),
	)
	return 0, fmt.Errorf("no available Tailscale port found starting from %d", actualStartPort)
}

// FindAvailableLocalPort finds an available local port starting from a random port
func FindAvailableLocalPort() (int, error) {
	// Use a random starting port in the ephemeral port range (49152-65535)
	startPort := 49152 + rand.IntN(10000) // Random port between 49152 and 59151

	for port := startPort; port < startPort+1000; port++ {
		if port > 65535 {
			// Wrap around if we exceed the port range
			port = 49152 + (port - 65535)
		}

		ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found starting from random port %d", startPort)
}
