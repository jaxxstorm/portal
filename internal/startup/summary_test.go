package startup

import (
	"testing"

	"github.com/jaxxstorm/portal/internal/config"
)

func TestBuildReadySummaryTailnetIncludesWebUI(t *testing.T) {
	cfg := &config.Config{
		Funnel: false,
		Mock:   false,
		NoUI:   false,
		NoTUI:  true,
		JSON:   false,
	}

	summary := BuildReadySummary(cfg, true, "https://node.ts.net", "http://localhost:1234", "https://node.ts.net:9999", TSNetDetails{})
	if !summary.IsReady() {
		t.Fatalf("expected summary to be ready")
	}
	if got, want := summary.Mode, ModeLocalDaemon; got != want {
		t.Fatalf("unexpected mode: got %q want %q", got, want)
	}
	if got, want := summary.BackendMode, BackendModeProxy; got != want {
		t.Fatalf("unexpected backend mode: got %q want %q", got, want)
	}
	if got, want := summary.Exposure, ExposureTailnet; got != want {
		t.Fatalf("unexpected exposure: got %q want %q", got, want)
	}
	if got, want := summary.WebUIStatus, WebUIStatusEnabled; got != want {
		t.Fatalf("unexpected web UI status: got %q want %q", got, want)
	}
	if got := summary.WebUIReason; got != "" {
		t.Fatalf("expected no web UI reason for enabled UI, got %q", got)
	}
}

func TestBuildReadySummaryFunnelMode(t *testing.T) {
	cfg := &config.Config{
		Funnel:          true,
		Mock:            true,
		NoUI:            false,
		NoTUI:           false,
		JSON:            true,
		TSNetListenMode: config.TSNetListenModeListener,
	}

	summary := BuildReadySummary(cfg, false, "https://node.ts.net", "", "", TSNetDetails{
		ConfiguredListenMode: config.TSNetListenModeListener,
		EffectiveListenMode:  config.TSNetListenModeListener,
	})
	if got, want := summary.Mode, ModeTSNet; got != want {
		t.Fatalf("unexpected mode: got %q want %q", got, want)
	}
	if got, want := summary.BackendMode, BackendModeMock; got != want {
		t.Fatalf("unexpected backend mode: got %q want %q", got, want)
	}
	if got, want := summary.Exposure, ExposureFunnel; got != want {
		t.Fatalf("unexpected exposure: got %q want %q", got, want)
	}
	if got, want := summary.WebUIStatus, WebUIStatusUnavailable; got != want {
		t.Fatalf("unexpected web UI status: got %q want %q", got, want)
	}
	if got, want := summary.WebUIReason, WebUIReasonTSNetNotSupported; got != want {
		t.Fatalf("unexpected web UI reason: got %q want %q", got, want)
	}
	if !summary.Capabilities.Funnel {
		t.Fatalf("expected funnel capability to be true")
	}
	if !summary.Capabilities.JSONLogging {
		t.Fatalf("expected json logging capability to be true")
	}
	if got, want := summary.ConfiguredListenMode, config.TSNetListenModeListener; got != want {
		t.Fatalf("unexpected configured tsnet listen mode: got %q want %q", got, want)
	}
	if got, want := summary.EffectiveListenMode, config.TSNetListenModeListener; got != want {
		t.Fatalf("unexpected effective tsnet listen mode: got %q want %q", got, want)
	}
	if got := summary.ServiceName; got != "" {
		t.Fatalf("expected empty tsnet service name, got %q", got)
	}
}

func TestBuildReadySummaryUIDisabled(t *testing.T) {
	cfg := &config.Config{
		NoUI: true,
	}

	summary := BuildReadySummary(cfg, true, "https://node.ts.net", "", "", TSNetDetails{})
	if got, want := summary.WebUIStatus, WebUIStatusDisabled; got != want {
		t.Fatalf("unexpected web UI status: got %q want %q", got, want)
	}
	if got, want := summary.WebUIReason, WebUIReasonDisabledByConfig; got != want {
		t.Fatalf("unexpected web UI reason: got %q want %q", got, want)
	}
}

func TestSummaryIsNotReadyWithoutServiceURL(t *testing.T) {
	summary := Summary{
		Readiness: ReadinessReady,
	}
	if summary.IsReady() {
		t.Fatalf("expected summary with empty service URL to be not ready")
	}
}

func TestSummaryFieldsExposeStableKeys(t *testing.T) {
	cfg := &config.Config{
		Funnel: false,
		NoUI:   false,
		NoTUI:  false,
	}
	summary := BuildReadySummary(cfg, true, "https://node.ts.net", "http://localhost:8080", "https://node.ts.net:9999", TSNetDetails{})

	fields := summary.Fields()
	keys := make(map[string]bool, len(fields))
	for _, field := range fields {
		keys[field.Key] = true
	}

	required := []string{
		"readiness",
		"mode",
		"backend_mode",
		"exposure",
		"service_url",
		"web_ui_status",
		"capability_funnel",
		"capability_ui",
		"capability_tui",
		"capability_json_logging",
		"web_ui_url",
	}

	for _, key := range required {
		if !keys[key] {
			t.Fatalf("expected field key %q to be present", key)
		}
	}
}

func TestBuildReadySummaryMockTailnetMode(t *testing.T) {
	cfg := &config.Config{
		Mock:   true,
		Funnel: false,
	}

	summary := BuildReadySummary(cfg, true, "https://node.ts.net", "", "", TSNetDetails{})
	if got, want := summary.BackendMode, BackendModeMock; got != want {
		t.Fatalf("unexpected backend mode: got %q want %q", got, want)
	}
	if got, want := summary.Exposure, ExposureTailnet; got != want {
		t.Fatalf("unexpected exposure: got %q want %q", got, want)
	}
}

func TestBuildReadySummaryLocalModeUnavailableReason(t *testing.T) {
	cfg := &config.Config{
		NoUI: false,
	}

	summary := BuildReadySummary(cfg, true, "https://node.ts.net", "", "", TSNetDetails{})
	if got, want := summary.WebUIReason, WebUIReasonSetupUnavailable; got != want {
		t.Fatalf("unexpected web UI reason: got %q want %q", got, want)
	}
}

func TestBuildReadySummaryTSNetServiceFQDNFromURL(t *testing.T) {
	cfg := &config.Config{
		NoUI:             true,
		TSNetListenMode:  config.TSNetListenModeService,
		TSNetServiceName: "svc:my-service",
	}

	summary := BuildReadySummary(cfg, false, "https://my-service.tail4cf751.ts.net", "", "", TSNetDetails{
		ConfiguredListenMode: config.TSNetListenModeService,
		EffectiveListenMode:  config.TSNetListenModeService,
		ServiceName:          "svc:my-service",
	})
	if got, want := summary.ServiceFQDN, "my-service.tail4cf751.ts.net"; got != want {
		t.Fatalf("unexpected service fqdn: got %q want %q", got, want)
	}
}

func TestSummaryFieldsIncludeTSNetModeFields(t *testing.T) {
	cfg := &config.Config{
		NoUI:             true,
		TSNetListenMode:  config.TSNetListenModeService,
		TSNetServiceName: "svc:portal",
	}

	summary := BuildReadySummary(cfg, false, "https://portal.tail4cf751.ts.net", "", "", TSNetDetails{
		ConfiguredListenMode: config.TSNetListenModeService,
		EffectiveListenMode:  config.TSNetListenModeService,
		ServiceName:          "svc:portal",
		ServiceFQDN:          "portal.tail4cf751.ts.net",
	})
	fields := summary.Fields()
	keys := make(map[string]bool, len(fields))
	for _, field := range fields {
		keys[field.Key] = true
	}

	for _, key := range []string{"tsnet_listen_mode_configured", "tsnet_listen_mode_effective", "tsnet_service_name", "tsnet_service_fqdn"} {
		if !keys[key] {
			t.Fatalf("expected field key %q to be present", key)
		}
	}
}
