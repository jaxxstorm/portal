// internal/tailscale/tsnet.go
package tailscale

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"

	"go.uber.org/zap"
	"tailscale.com/ipn"
	"tailscale.com/tailcfg"
	"tailscale.com/tsnet"

	"github.com/jaxxstorm/tgate/internal/logging"
)

const (
	TSNetListenModeListener = "listener"
	TSNetListenModeService  = "service"
)

// TSNetConfig holds configuration for tsnet mode
type TSNetConfig struct {
	Hostname     string
	AuthKey      string
	EnableFunnel bool
	UseHTTPS     bool
	ServePort    int
	ListenMode   string
	ServiceName  string
}

// TSNetReadyInfo captures serving details emitted once TSNet is ready.
type TSNetReadyInfo struct {
	ServiceURL           string
	ConfiguredListenMode string
	EffectiveListenMode  string
	ServiceName          string
	ServiceFQDN          string
}

// TSNetServer wraps a tsnet server with additional functionality
type TSNetServer struct {
	server        *tsnet.Server
	logger        *zap.Logger
	config        TSNetConfig
	readyCallback func(TSNetReadyInfo)
	readyMu       sync.RWMutex
}

type funnelClientIPContextKey struct{}

// NewTSNetServer creates a new tsnet server with structured logging
func NewTSNetServer(config TSNetConfig, logger *zap.Logger) *TSNetServer {
	server := &tsnet.Server{
		Hostname: config.Hostname,
		AuthKey:  config.AuthKey,
		Logf:     newTSNetRuntimeLogAdapter(logger, config.Hostname),
		UserLogf: newTSNetRuntimeLogAdapter(logger, config.Hostname),
	}

	logger.Info("Creating TSNet server",
		logging.Component("tsnet_server"),
		logging.TailscaleMode("tsnet"),
		logging.Phase("initialization"),
		logging.NodeName(config.Hostname),
		zap.Bool("has_auth_key", config.AuthKey != ""),
		logging.FunnelEnabled(config.EnableFunnel),
		logging.HTTPSEnabled(config.UseHTTPS),
		logging.ServePort(resolveTSNetServePort(config)),
		zap.String("listen_mode", normalizeTSNetListenMode(config.ListenMode)),
		zap.String("service_name", strings.TrimSpace(config.ServiceName)),
	)

	return &TSNetServer{
		server: server,
		logger: logger,
		config: config,
	}
}

// Listen creates a listener on the tsnet server
func (ts *TSNetServer) Listen(network, addr string) (net.Listener, error) {
	ts.logger.Info("Creating TSNet listener",
		logging.Component("tsnet_server"),
		logging.TailscaleMode("tsnet"),
		logging.Phase("listener_setup"),
		logging.NodeName(ts.config.Hostname),
		zap.String("network", network),
		zap.String("address", addr),
	)

	ln, err := ts.server.Listen(network, addr)
	if err != nil {
		ts.logger.Error("Failed to create TSNet listener",
			logging.Component("tsnet_server"),
			logging.TailscaleMode("tsnet"),
			logging.Phase("listener_setup"),
			zap.String("network", network),
			zap.String("address", addr),
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to listen on Tailscale device: %w", err)
	}

	ts.logger.Info("TSNet listener created successfully",
		logging.Component("tsnet_server"),
		logging.TailscaleMode("tsnet"),
		logging.Phase("listener_ready"),
		zap.String("network", network),
		zap.String("address", addr),
	)

	return ln, nil
}

// Start starts the tsnet server and returns status information
func (ts *TSNetServer) Start(ctx context.Context) (string, error) {
	ts.logger.Info("Starting TSNet server",
		logging.Component("tsnet_server"),
		logging.TailscaleMode("tsnet"),
		logging.Phase("startup"),
		logging.NodeName(ts.config.Hostname),
	)

	// Get the device's Tailscale URL
	status, err := ts.server.Up(ctx)
	if err != nil {
		ts.logger.Error("Failed to start TSNet server",
			logging.Component("tsnet_server"),
			logging.TailscaleMode("tsnet"),
			logging.Phase("startup"),
			logging.NodeName(ts.config.Hostname),
			logging.Error(err),
		)
		return "", err
	}

	port, useTLS := ts.serveSettings()
	dnsName := strings.TrimSuffix(status.Self.DNSName, ".")
	tailscaleURL := buildTSNetServiceURL(dnsName, port, useTLS)

	ts.logger.Info("TSNet server started successfully",
		logging.Component("tsnet_server"),
		logging.TailscaleMode("tsnet"),
		logging.Phase("startup_complete"),
		logging.NodeName(ts.config.Hostname),
		logging.URL(tailscaleURL),
		zap.String("dns_name", dnsName),
	)

	return tailscaleURL, nil
}

// Serve starts serving HTTP on the tsnet server
func (ts *TSNetServer) Serve(ctx context.Context, handler http.Handler) error {
	configuredMode := normalizeTSNetListenMode(ts.config.ListenMode)
	effectiveMode := effectiveTSNetListenMode(configuredMode)
	if err := validateTSNetListenConfig(ts.config, configuredMode); err != nil {
		ts.logger.Error("Invalid TSNet listen configuration",
			logging.Component("tsnet_server"),
			logging.TailscaleMode("tsnet"),
			logging.Phase("listener_setup"),
			zap.String("configured_listen_mode", configuredMode),
			zap.String("service_name", strings.TrimSpace(ts.config.ServiceName)),
			logging.Error(err),
		)
		return err
	}

	port, useTLS := ts.serveSettings()
	addr := fmt.Sprintf(":%d", port)
	if err := validateTSNetServeConfig(ts.config, port, useTLS); err != nil {
		ts.logger.Error("Invalid TSNet serve configuration",
			logging.Component("tsnet_server"),
			logging.TailscaleMode("tsnet"),
			logging.Phase("listener_setup"),
			logging.FunnelEnabled(ts.config.EnableFunnel),
			logging.HTTPSEnabled(useTLS),
			logging.ServePort(port),
			logging.Error(err),
		)
		return err
	}

	ts.logger.Info("Setting up TSNet HTTP server",
		logging.Component("tsnet_server"),
		logging.TailscaleMode("tsnet"),
		logging.Phase("http_setup"),
		logging.NodeName(ts.config.Hostname),
		logging.FunnelEnabled(ts.config.EnableFunnel),
		logging.HTTPSEnabled(useTLS),
		logging.ServePort(port),
		zap.String("configured_listen_mode", configuredMode),
		zap.String("effective_listen_mode", effectiveMode),
		zap.String("service_name", strings.TrimSpace(ts.config.ServiceName)),
	)

	ln, serviceFQDN, serviceName, err := ts.listenForServe(addr, port, useTLS, effectiveMode)
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		ConnContext: func(ctx context.Context, conn net.Conn) context.Context {
			if sourceIP, ok := funnelSourceIPFromConn(conn); ok {
				return context.WithValue(ctx, funnelClientIPContextKey{}, sourceIP.String())
			}
			return ctx
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if sourceIP, ok := funnelClientIPFromContext(r.Context()); ok {
				r.Header.Set("Tailscale-Client-IP", sourceIP)
			}
			handler.ServeHTTP(w, r)
		}),
	}

	// Start the device
	serviceURL, err := ts.Start(ctx)
	if err != nil {
		return err
	}
	if effectiveMode == TSNetListenModeService && serviceFQDN != "" {
		serviceURL = buildTSNetServiceURL(serviceFQDN, port, useTLS)
	}

	ts.emitReady(TSNetReadyInfo{
		ServiceURL:           serviceURL,
		ConfiguredListenMode: configuredMode,
		EffectiveListenMode:  effectiveMode,
		ServiceName:          serviceName,
		ServiceFQDN:          serviceFQDN,
	})

	ts.logger.Info("TSNet HTTP server ready to serve",
		logging.Component("tsnet_server"),
		logging.TailscaleMode("tsnet"),
		logging.Phase("serving"),
		logging.NodeName(ts.config.Hostname),
		logging.Port(port),
		logging.FunnelEnabled(ts.config.EnableFunnel),
		logging.HTTPSEnabled(useTLS),
		logging.Status("serving"),
		zap.String("configured_listen_mode", configuredMode),
		zap.String("effective_listen_mode", effectiveMode),
		zap.String("service_name", serviceName),
		zap.String("service_fqdn", serviceFQDN),
		logging.URL(serviceURL),
	)

	// Serve HTTP requests
	if err := httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
		ts.logger.Error("TSNet HTTP server error",
			logging.Component("tsnet_server"),
			logging.TailscaleMode("tsnet"),
			logging.Phase("serving"),
			logging.Error(err),
		)
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

// SetReadyCallback sets a callback that is invoked once tsnet serving is ready.
func (ts *TSNetServer) SetReadyCallback(callback func(TSNetReadyInfo)) {
	ts.readyMu.Lock()
	defer ts.readyMu.Unlock()
	ts.readyCallback = callback
}

func (ts *TSNetServer) emitReady(info TSNetReadyInfo) {
	if strings.TrimSpace(info.ServiceURL) == "" {
		return
	}

	ts.readyMu.RLock()
	callback := ts.readyCallback
	ts.readyMu.RUnlock()
	if callback != nil {
		callback(info)
	}
}

func (ts *TSNetServer) listenForServe(addr string, port int, useTLS bool, effectiveMode string) (net.Listener, string, string, error) {
	ts.logger.Info("Creating TSNet listener",
		logging.Component("tsnet_server"),
		logging.TailscaleMode("tsnet"),
		logging.Phase("listener_setup"),
		logging.NodeName(ts.config.Hostname),
		zap.String("network", "tcp"),
		zap.String("address", addr),
		logging.FunnelEnabled(ts.config.EnableFunnel),
		logging.HTTPSEnabled(useTLS),
		zap.String("effective_listen_mode", effectiveMode),
		zap.String("service_name", strings.TrimSpace(ts.config.ServiceName)),
	)

	var (
		ln          net.Listener
		err         error
		serviceFQDN string
		serviceName string
	)

	switch effectiveMode {
	case TSNetListenModeService:
		serviceName = strings.TrimSpace(ts.config.ServiceName)
		serviceMode := tsnet.ServiceModeHTTP{
			Port:  uint16(port),
			HTTPS: useTLS,
		}
		serviceListener, listenErr := ts.server.ListenService(serviceName, serviceMode)
		if listenErr != nil {
			err = formatListenServiceError(serviceName, listenErr)
		} else {
			ln = serviceListener
			serviceFQDN = serviceListener.FQDN
		}
	default:
		switch {
		case ts.config.EnableFunnel:
			ln, err = ts.server.ListenFunnel("tcp", addr)
		case useTLS:
			ln, err = ts.server.ListenTLS("tcp", addr)
		default:
			ln, err = ts.server.Listen("tcp", addr)
		}
	}

	if err != nil {
		ts.logger.Error("Failed to create TSNet listener",
			logging.Component("tsnet_server"),
			logging.TailscaleMode("tsnet"),
			logging.Phase("listener_setup"),
			zap.String("network", "tcp"),
			zap.String("address", addr),
			logging.FunnelEnabled(ts.config.EnableFunnel),
			logging.HTTPSEnabled(useTLS),
			zap.String("effective_listen_mode", effectiveMode),
			zap.String("service_name", serviceName),
			logging.Error(err),
		)
		return nil, "", serviceName, err
	}

	ts.logger.Info("TSNet listener created successfully",
		logging.Component("tsnet_server"),
		logging.TailscaleMode("tsnet"),
		logging.Phase("listener_ready"),
		zap.String("network", "tcp"),
		zap.String("address", addr),
		logging.FunnelEnabled(ts.config.EnableFunnel),
		logging.HTTPSEnabled(useTLS),
		zap.String("effective_listen_mode", effectiveMode),
		zap.String("service_name", serviceName),
		zap.String("service_fqdn", serviceFQDN),
	)

	return ln, serviceFQDN, serviceName, nil
}

func (ts *TSNetServer) serveSettings() (port int, useTLS bool) {
	port = resolveTSNetServePort(ts.config)
	useTLS = shouldUseTSNetTLS(ts.config, port)
	return port, useTLS
}

func resolveTSNetServePort(config TSNetConfig) int {
	if config.ServePort > 0 {
		return config.ServePort
	}
	if config.EnableFunnel || config.UseHTTPS {
		return 443
	}
	return 80
}

func shouldUseTSNetTLS(config TSNetConfig, servePort int) bool {
	return config.EnableFunnel || config.UseHTTPS || servePort == 443
}

func validateTSNetServeConfig(config TSNetConfig, servePort int, useTLS bool) error {
	if !config.EnableFunnel {
		return nil
	}
	if !useTLS {
		return fmt.Errorf("tsnet funnel requires HTTPS")
	}
	switch servePort {
	case 443, 8443, 10000:
		return nil
	default:
		return fmt.Errorf("tsnet funnel requires serve port 443, 8443, or 10000 (got %d)", servePort)
	}
}

func validateTSNetListenConfig(config TSNetConfig, configuredMode string) error {
	if configuredMode != TSNetListenModeListener && configuredMode != TSNetListenModeService {
		return fmt.Errorf("invalid tsnet listen mode %q", configuredMode)
	}
	if configuredMode != TSNetListenModeService {
		return nil
	}
	if config.EnableFunnel {
		return fmt.Errorf("tsnet service mode is mutually exclusive with funnel")
	}

	serviceName := strings.TrimSpace(config.ServiceName)
	if serviceName == "" {
		return fmt.Errorf("tsnet service mode requires a non-empty service name")
	}
	if err := tailcfg.ServiceName(serviceName).Validate(); err != nil {
		return fmt.Errorf("invalid tsnet service name %q: %w", serviceName, err)
	}

	servePort := resolveTSNetServePort(config)
	if servePort <= 0 || servePort > 65535 {
		return fmt.Errorf("tsnet service mode requires a valid TCP port, got %d", servePort)
	}
	return nil
}

func normalizeTSNetListenMode(listenMode string) string {
	mode := strings.ToLower(strings.TrimSpace(listenMode))
	if mode == "" {
		return TSNetListenModeListener
	}
	return mode
}

func effectiveTSNetListenMode(configuredMode string) string {
	return configuredMode
}

func formatListenServiceError(serviceName string, err error) error {
	if errors.Is(err, tsnet.ErrUntaggedServiceHost) {
		return fmt.Errorf("failed to create tsnet service listener for %q: %w. service hosts must be tagged nodes and may require admin approval", serviceName, err)
	}
	return fmt.Errorf("failed to create tsnet service listener for %q: %w. verify service host tags and approval prerequisites", serviceName, err)
}

func buildTSNetServiceURL(dnsName string, servePort int, useTLS bool) string {
	scheme := "http"
	if useTLS {
		scheme = "https"
	}

	if (useTLS && servePort == 443) || (!useTLS && servePort == 80) {
		return fmt.Sprintf("%s://%s", scheme, dnsName)
	}
	return fmt.Sprintf("%s://%s:%d", scheme, dnsName, servePort)
}

func funnelSourceIPFromConn(conn net.Conn) (netip.Addr, bool) {
	switch c := conn.(type) {
	case *ipn.FunnelConn:
		addr := c.Src.Addr().Unmap()
		if addr.IsValid() {
			return addr, true
		}
	case *tls.Conn:
		return funnelSourceIPFromConn(c.NetConn())
	}

	return netip.Addr{}, false
}

func funnelClientIPFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(funnelClientIPContextKey{}).(string)
	if !ok {
		return "", false
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}

	return value, true
}

// Close closes the tsnet server
func (ts *TSNetServer) Close() error {
	ts.logger.Info("Closing TSNet server",
		logging.Component("tsnet_server"),
		logging.TailscaleMode("tsnet"),
		logging.Phase("shutdown"),
		logging.NodeName(ts.config.Hostname),
	)

	err := ts.server.Close()
	if err != nil {
		ts.logger.Error("Error closing TSNet server",
			logging.Component("tsnet_server"),
			logging.TailscaleMode("tsnet"),
			logging.Phase("shutdown"),
			logging.Error(err),
		)
	} else {
		ts.logger.Info("TSNet server closed successfully",
			logging.Component("tsnet_server"),
			logging.TailscaleMode("tsnet"),
			logging.Phase("shutdown_complete"),
		)
	}

	return err
}
