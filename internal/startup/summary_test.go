package startup

import (
	"testing"

	"github.com/jaxxstorm/tgate/internal/config"
)

func TestBuildReadySummaryTailnetIncludesWebUI(t *testing.T) {
	cfg := &config.Config{
		Funnel: false,
		Mock:   false,
		NoUI:   false,
		NoTUI:  true,
		JSON:   false,
	}

	summary := BuildReadySummary(cfg, true, "https://node.ts.net", "http://localhost:1234", "https://node.ts.net:9999")
	if !summary.IsReady() {
		t.Fatalf("expected summary to be ready")
	}
	if got, want := summary.Mode, ModeLocalDaemon; got != want {
		t.Fatalf("unexpected mode: got %q want %q", got, want)
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
		Funnel: true,
		Mock:   true,
		NoUI:   false,
		NoTUI:  false,
		JSON:   true,
	}

	summary := BuildReadySummary(cfg, false, "https://node.ts.net", "", "")
	if got, want := summary.Mode, ModeTSNet; got != want {
		t.Fatalf("unexpected mode: got %q want %q", got, want)
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
}

func TestBuildReadySummaryUIDisabled(t *testing.T) {
	cfg := &config.Config{
		NoUI: true,
	}

	summary := BuildReadySummary(cfg, true, "https://node.ts.net", "", "")
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
	summary := BuildReadySummary(cfg, true, "https://node.ts.net", "http://localhost:8080", "https://node.ts.net:9999")

	fields := summary.Fields()
	keys := make(map[string]bool, len(fields))
	for _, field := range fields {
		keys[field.Key] = true
	}

	required := []string{
		"readiness",
		"mode",
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

func TestBuildReadySummaryLocalModeUnavailableReason(t *testing.T) {
	cfg := &config.Config{
		NoUI: false,
	}

	summary := BuildReadySummary(cfg, true, "https://node.ts.net", "", "")
	if got, want := summary.WebUIReason, WebUIReasonSetupUnavailable; got != want {
		t.Fatalf("unexpected web UI reason: got %q want %q", got, want)
	}
}
