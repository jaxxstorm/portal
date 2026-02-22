package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/jaxxstorm/portal/internal/model"
)

type stubStatsProvider struct {
	state         model.EndpointState
	ttl, opn      int
	rt1, rt5, p50 float64
	p90           float64
}

func (s *stubStatsProvider) GetStats() (ttl, opn int, rt1, rt5, p50, p90 float64) {
	return s.ttl, s.opn, s.rt1, s.rt5, s.p50, s.p90
}

func (s *stubStatsProvider) GetEndpointState() model.EndpointState {
	return s.state
}

func resizeModel(t *testing.T, m *Model, width, height int) {
	t.Helper()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	next, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected updated model type %T", m)
	}
	*m = next
}

func updateModel(t *testing.T, m *Model, msg tea.Msg) {
	t.Helper()
	updated, _ := m.Update(msg)
	next, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected updated model type %T", m)
	}
	*m = next
}

func normalizePaneText(in string) string {
	lines := strings.Split(in, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n")
}

func TestLayoutProfilesAndResizePreserveState(t *testing.T) {
	provider := &stubStatsProvider{
		state: model.EndpointState{
			Readiness:   model.EndpointReadinessStarting,
			Mode:        "local_daemon",
			Exposure:    "tailnet",
			WebUIStatus: "unavailable",
			WebUIReason: "ui_setup_unavailable",
		},
	}

	m := NewModel(provider)

	resizeModel(t, &m, 100, 30)
	if got, want := m.layout.profile, layoutCompact; got != want {
		t.Fatalf("unexpected layout profile for compact size: got %q want %q", got, want)
	}

	request := model.RequestLog{
		Method:     "GET",
		URL:        "/health",
		RemoteAddr: "100.1.2.3:1234",
		Timestamp:  time.Now(),
		Response: model.ResponseLog{
			StatusCode: 200,
		},
		Duration: 5 * time.Millisecond,
	}
	updateModel(t, &m, RequestMsg{Log: request})
	updateModel(t, &m, LogMsg{Level: "INFO", Message: "startup complete", Time: time.Now()})

	resizeModel(t, &m, 140, 40)
	if got, want := m.layout.profile, layoutStandard; got != want {
		t.Fatalf("unexpected layout profile for standard size: got %q want %q", got, want)
	}
	if m.layout.endpointWidth >= m.layout.logsWidth {
		t.Fatalf("expected endpoint pane to be narrower than full content width")
	}
	if m.layout.headersWidth <= m.layout.statsWidth {
		t.Fatalf("expected request details pane to be wider than stats pane")
	}
	if m.lastRequest == nil {
		t.Fatalf("expected last request to be preserved after resize")
	}
	if content := m.appLogs.View(); !strings.Contains(content, "startup complete") {
		t.Fatalf("expected log history to be preserved after resize, got %q", content)
	}

	resizeModel(t, &m, 200, 56)
	if got, want := m.layout.profile, layoutWide; got != want {
		t.Fatalf("unexpected layout profile for wide size: got %q want %q", got, want)
	}
}

func TestEndpointSummaryTailnetRendering(t *testing.T) {
	provider := &stubStatsProvider{
		state: model.EndpointState{
			Readiness:   model.EndpointReadinessReady,
			Mode:        "local_daemon",
			Exposure:    "tailnet",
			ServiceURL:  "http://portal.tail4cf751.ts.net/",
			WebUIStatus: "enabled",
			WebUIURL:    "http://lbr-macbook-pro.tail4cf751.ts.net:8147/ui/",
		},
	}

	m := NewModel(provider)
	resizeModel(t, &m, 150, 44)

	content := m.endpointPane.View()
	content = normalizePaneText(content)
	for _, required := range []string{
		"Mode: local_daemon  Exposure: tailnet-private",
		"Service: http://portal.tail4cf751.ts.net/",
		"Web UI: enabled  http://lbr-macbook-pro.tail4cf751.ts.net:8147/ui/",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("expected endpoint summary to contain %q, got %q", required, content)
		}
	}
}

func TestFunnelLongURLRenderingPreservesHost(t *testing.T) {
	longURL := "https://portal-super-long-service-name.tail4cf751.ts.net/path/with/many/segments/that/need/to/be/truncated?token=abcdef1234567890&debug=true"
	lines := formatURLForDisplay(longURL, 42, 3)
	if len(lines) != 1 {
		t.Fatalf("expected single-line URL rendering, got %d (%q)", len(lines), lines)
	}
	joined := lines[0]
	if !strings.HasPrefix(joined, "https://portal-super-long-service-name") {
		t.Fatalf("expected formatted URL to preserve URL prefix, got %q", lines)
	}
	if !strings.HasSuffix(lines[0], "...") {
		t.Fatalf("expected long URL rendering to end with ellipsis, got %q", lines)
	}

	provider := &stubStatsProvider{
		state: model.EndpointState{
			Readiness:   model.EndpointReadinessReady,
			Mode:        "local_daemon",
			Exposure:    "funnel",
			ServiceURL:  longURL,
			WebUIStatus: "enabled",
			WebUIURL:    "http://portal.tail4cf751.ts.net:8094/ui/",
		},
	}

	m := NewModel(provider)
	resizeModel(t, &m, 96, 32)
	content := m.endpointPane.View()
	if !strings.Contains(content, "Exposure: funnel-public") {
		t.Fatalf("expected funnel exposure label in endpoint summary, got %q", content)
	}
}

func TestEndpointSummaryUnavailableAndFailureStates(t *testing.T) {
	t.Run("tsnet unavailable ui", func(t *testing.T) {
		provider := &stubStatsProvider{
			state: model.EndpointState{
				Readiness:   model.EndpointReadinessReady,
				Mode:        "tsnet",
				Exposure:    "tailnet",
				ServiceURL:  "https://portal.tail4cf751.ts.net",
				WebUIStatus: "unavailable",
				WebUIReason: "tsnet_ui_not_exposed",
			},
		}

		m := NewModel(provider)
		resizeModel(t, &m, 140, 42)
		content := normalizePaneText(m.endpointPane.View())
		for _, required := range []string{"Mode: tsnet  Exposure: tailnet-private", "Web UI: unavailable", "Web UI Reason: tsnet_ui_not_exposed"} {
			if !strings.Contains(content, required) {
				t.Fatalf("expected endpoint summary to contain %q, got %q", required, content)
			}
		}
	})

	t.Run("startup failure state", func(t *testing.T) {
		provider := &stubStatsProvider{
			state: model.EndpointState{
				Readiness:   model.EndpointReadinessFailed,
				Mode:        "local_daemon",
				Exposure:    "funnel",
				ServiceURL:  "",
				WebUIStatus: "unavailable",
				WebUIReason: "funnel requires HTTPS on port 443",
			},
		}

		m := NewModel(provider)
		resizeModel(t, &m, 140, 42)
		content := normalizePaneText(m.endpointPane.View())
		for _, required := range []string{"Mode: local_daemon  Exposure: funnel-public", "Service: unavailable", "Web UI Reason: funnel requires HTTPS on port 443"} {
			if !strings.Contains(content, required) {
				t.Fatalf("expected endpoint summary to contain %q, got %q", required, content)
			}
		}
		if strings.Contains(content, "http://") || strings.Contains(content, "https://") {
			t.Fatalf("expected failure state to avoid ready-like service URL, got %q", content)
		}
	})
}

func TestScrollKeysOnlyMoveLogsPane(t *testing.T) {
	provider := &stubStatsProvider{
		state: model.EndpointState{
			Readiness:   model.EndpointReadinessReady,
			Mode:        "local_daemon",
			Exposure:    "tailnet",
			ServiceURL:  "http://portal.tail4cf751.ts.net/",
			WebUIStatus: "enabled",
			WebUIURL:    "http://node.tail4cf751.ts.net:8141/ui/",
		},
	}

	m := NewModel(provider)
	resizeModel(t, &m, 140, 42)

	updateModel(t, &m, LogMsg{Level: "INFO", Message: "line 1", Time: time.Now()})
	updateModel(t, &m, LogMsg{Level: "INFO", Message: "line 2", Time: time.Now()})
	updateModel(t, &m, LogMsg{Level: "INFO", Message: "line 3", Time: time.Now()})

	before := normalizePaneText(m.endpointPane.View())
	if !strings.Contains(before, "Mode: local_daemon  Exposure: tailnet-private") {
		t.Fatalf("expected endpoint pane top content before scroll, got %q", before)
	}

	updateModel(t, &m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	after := normalizePaneText(m.endpointPane.View())
	if !strings.Contains(after, "Mode: local_daemon  Exposure: tailnet-private") {
		t.Fatalf("expected endpoint pane to remain fixed while scrolling logs, got %q", after)
	}
}

func TestViewFitsWithinWindowHeight(t *testing.T) {
	provider := &stubStatsProvider{
		state: model.EndpointState{
			Readiness:   model.EndpointReadinessReady,
			Mode:        "local_daemon",
			Exposure:    "tailnet",
			ServiceURL:  "http://portal.tail4cf751.ts.net/",
			WebUIStatus: "enabled",
			WebUIURL:    "http://node.tail4cf751.ts.net:8141/ui/",
		},
	}

	cases := []struct {
		name   string
		width  int
		height int
	}{
		{name: "compact", width: 100, height: 30},
		{name: "standard", width: 140, height: 42},
		{name: "wide", width: 200, height: 56},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewModel(provider)
			resizeModel(t, &m, tc.width, tc.height)
			updateModel(t, &m, LogMsg{
				Level:   "INFO",
				Message: strings.Repeat("very-long-startup-log-segment-", 40),
				Time:    time.Now(),
			})
			view := m.View()
			lines := strings.Count(view, "\n") + 1
			if lines > tc.height {
				t.Fatalf("view overflow: got %d lines for height %d", lines, tc.height)
			}
			for _, line := range strings.Split(view, "\n") {
				if w := ansi.StringWidth(line); w > tc.width {
					t.Fatalf("line overflow: got width %d for terminal width %d", w, tc.width)
				}
			}
		})
	}
}
