package startup

import (
	"net/url"
	"strings"

	"go.uber.org/zap"

	"github.com/jaxxstorm/portal/internal/config"
	"github.com/jaxxstorm/portal/internal/logging"
	"github.com/jaxxstorm/portal/internal/model"
)

const (
	ReadinessReady = "ready"

	ModeLocalDaemon = "local_daemon"
	ModeTSNet       = "tsnet"

	ExposureTailnet = "tailnet"
	ExposureFunnel  = "funnel"

	BackendModeProxy = "proxy"
	BackendModeMock  = "mock"

	WebUIStatusEnabled     = "enabled"
	WebUIStatusDisabled    = "disabled"
	WebUIStatusUnavailable = "unavailable"

	WebUIReasonDisabledByConfig  = "disabled_by_configuration"
	WebUIReasonTSNetNotSupported = "tsnet_ui_not_exposed"
	WebUIReasonSetupUnavailable  = "ui_setup_unavailable"
)

type Capabilities struct {
	Funnel      bool
	Mock        bool
	UI          bool
	TUI         bool
	JSONLogging bool
	HTTPS       bool
}

type TSNetDetails struct {
	ConfiguredListenMode string
	EffectiveListenMode  string
	ServiceName          string
	ServiceFQDN          string
}

type Summary struct {
	Readiness   string
	Mode        string
	BackendMode string
	Exposure    string
	ServiceURL  string
	LocalURL    string
	WebUIStatus string
	WebUIURL    string
	WebUIReason string
	TSNetDetails
	Capabilities
}

func BuildReadySummary(cfg *config.Config, useLocalDaemon bool, serviceURL, localURL, webUIURL string, tsnetDetails TSNetDetails) Summary {
	mode := ModeTSNet
	if useLocalDaemon {
		mode = ModeLocalDaemon
	}

	exposure := ExposureTailnet
	if cfg.Funnel {
		exposure = ExposureFunnel
	}

	summary := Summary{
		Readiness:   ReadinessReady,
		Mode:        mode,
		BackendMode: resolveBackendMode(cfg),
		Exposure:    exposure,
		ServiceURL:  strings.TrimSpace(serviceURL),
		LocalURL:    strings.TrimSpace(localURL),
		WebUIStatus: ResolveWebUIStatus(cfg.NoUI, webUIURL),
		WebUIURL:    strings.TrimSpace(webUIURL),
		WebUIReason: ResolveWebUIReason(cfg.NoUI, useLocalDaemon, webUIURL),
		Capabilities: Capabilities{
			Funnel:      cfg.Funnel,
			Mock:        cfg.Mock,
			UI:          !cfg.NoUI,
			TUI:         !cfg.NoTUI,
			JSONLogging: cfg.JSON,
			HTTPS:       cfg.UseHTTPS,
		},
	}

	if mode == ModeTSNet {
		configured := strings.TrimSpace(tsnetDetails.ConfiguredListenMode)
		if configured == "" {
			configured = cfg.TSNetListenMode
		}
		if configured == "" {
			configured = config.TSNetListenModeListener
		}

		effective := strings.TrimSpace(tsnetDetails.EffectiveListenMode)
		if effective == "" {
			effective = cfg.EffectiveTSNetListenMode()
		}
		if effective == "" {
			effective = config.TSNetListenModeListener
		}

		serviceName := strings.TrimSpace(tsnetDetails.ServiceName)
		if serviceName == "" && configured == config.TSNetListenModeService {
			serviceName = strings.TrimSpace(cfg.TSNetServiceName)
		}

		serviceFQDN := strings.TrimSpace(tsnetDetails.ServiceFQDN)
		if serviceFQDN == "" && effective == config.TSNetListenModeService && summary.ServiceURL != "" {
			if parsedURL, err := url.Parse(summary.ServiceURL); err == nil {
				serviceFQDN = strings.TrimSpace(parsedURL.Hostname())
			}
		}

		summary.TSNetDetails = TSNetDetails{
			ConfiguredListenMode: configured,
			EffectiveListenMode:  effective,
			ServiceName:          serviceName,
			ServiceFQDN:          serviceFQDN,
		}
	}

	return summary
}

func ResolveWebUIStatus(uiDisabled bool, webUIURL string) string {
	if uiDisabled {
		return WebUIStatusDisabled
	}
	if strings.TrimSpace(webUIURL) == "" {
		return WebUIStatusUnavailable
	}
	return WebUIStatusEnabled
}

func ResolveWebUIReason(uiDisabled bool, useLocalDaemon bool, webUIURL string) string {
	if uiDisabled {
		return WebUIReasonDisabledByConfig
	}
	if strings.TrimSpace(webUIURL) != "" {
		return ""
	}
	if !useLocalDaemon {
		return WebUIReasonTSNetNotSupported
	}
	return WebUIReasonSetupUnavailable
}

func (s Summary) IsReady() bool {
	return s.Readiness == ReadinessReady && strings.TrimSpace(s.ServiceURL) != ""
}

func (s Summary) EndpointState() model.EndpointState {
	state := model.EndpointState{
		Readiness:   strings.TrimSpace(s.Readiness),
		Mode:        strings.TrimSpace(s.Mode),
		Exposure:    strings.TrimSpace(s.Exposure),
		ServiceURL:  strings.TrimSpace(s.ServiceURL),
		WebUIStatus: strings.TrimSpace(s.WebUIStatus),
		WebUIURL:    strings.TrimSpace(s.WebUIURL),
		WebUIReason: strings.TrimSpace(s.WebUIReason),
	}
	if state.Readiness == "" {
		state.Readiness = model.EndpointReadinessStarting
	}
	return state
}

func (s Summary) Fields() []zap.Field {
	fields := []zap.Field{
		logging.Component("startup"),
		zap.String("readiness", s.Readiness),
		zap.String("mode", s.Mode),
		zap.String("backend_mode", s.BackendMode),
		zap.String("exposure", s.Exposure),
		zap.String("service_url", s.ServiceURL),
		zap.String("web_ui_status", s.WebUIStatus),
		zap.Bool("capability_funnel", s.Funnel),
		zap.Bool("capability_mock", s.Mock),
		zap.Bool("capability_ui", s.UI),
		zap.Bool("capability_tui", s.TUI),
		zap.Bool("capability_json_logging", s.JSONLogging),
		zap.Bool("capability_https", s.HTTPS),
	}
	if s.LocalURL != "" {
		fields = append(fields, zap.String("local_url", s.LocalURL))
	}
	if s.WebUIURL != "" {
		fields = append(fields, zap.String("web_ui_url", s.WebUIURL))
	}
	if s.WebUIReason != "" {
		fields = append(fields, zap.String("web_ui_reason", s.WebUIReason))
	}
	if s.Mode == ModeTSNet {
		fields = append(fields,
			zap.String("tsnet_listen_mode_configured", s.ConfiguredListenMode),
			zap.String("tsnet_listen_mode_effective", s.EffectiveListenMode),
		)
		if s.ServiceName != "" {
			fields = append(fields, zap.String("tsnet_service_name", s.ServiceName))
		}
		if s.ServiceFQDN != "" {
			fields = append(fields, zap.String("tsnet_service_fqdn", s.ServiceFQDN))
		}
	}

	return fields
}

func resolveBackendMode(cfg *config.Config) string {
	if cfg != nil && cfg.Mock {
		return BackendModeMock
	}
	return BackendModeProxy
}
