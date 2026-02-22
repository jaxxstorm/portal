package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/jaxxstorm/portal/internal/config"
	"github.com/jaxxstorm/portal/internal/logging"
	"github.com/jaxxstorm/portal/internal/startup"
)

func TestLogStartupSummaryEmitsStructuredJSON(t *testing.T) {
	cfg := &config.Config{
		Funnel: true,
		NoUI:   false,
		NoTUI:  true,
		JSON:   true,
		Mock:   false,
	}
	summary := startup.BuildReadySummary(cfg, false, "https://node.ts.net", "", "", startup.TSNetDetails{})

	var buf bytes.Buffer
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(logging.JSONEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	))

	logStartupSummary(logger, summary)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 || lines[0] == "" {
		t.Fatalf("expected exactly one json log line, got %q", buf.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &payload); err != nil {
		t.Fatalf("failed to decode json log: %v", err)
	}

	if got, want := payload["message"], logging.MsgStartupReady; got != want {
		t.Fatalf("unexpected message: got %v want %q", got, want)
	}
	for _, key := range []string{"readiness", "mode", "backend_mode", "exposure", "service_url", "web_ui_status", "web_ui_reason", "capability_funnel", "tsnet_listen_mode_configured", "tsnet_listen_mode_effective"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in startup payload", key)
		}
	}
}

func TestLogStartupSummarySkipsNonReadyState(t *testing.T) {
	summary := startup.Summary{
		Readiness: startup.ReadinessReady,
	}

	var buf bytes.Buffer
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(logging.JSONEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	))

	logStartupSummary(logger, summary)
	if got := strings.TrimSpace(buf.String()); got != "" {
		t.Fatalf("expected no startup summary log for non-ready summary, got %q", got)
	}
}

func TestLogStartupSummarySkipsWhenFunnelStartupNotReady(t *testing.T) {
	cfg := &config.Config{
		Funnel: true,
		Mock:   true,
	}
	summary := startup.BuildReadySummary(cfg, true, "", "", "", startup.TSNetDetails{})

	var buf bytes.Buffer
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(logging.JSONEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	))

	logStartupSummary(logger, summary)
	if got := strings.TrimSpace(buf.String()); got != "" {
		t.Fatalf("expected no startup summary log when funnel startup is not ready, got %q", got)
	}
}
