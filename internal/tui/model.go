// internal/tui/model.go
package tui

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/jaxxstorm/portal/internal/model"
)

const (
	compactWidthThreshold  = 120
	compactHeightThreshold = 34
	wideWidthThreshold     = 180
	wideHeightThreshold    = 48
)

type layoutProfile string

const (
	layoutCompact  layoutProfile = "compact"
	layoutStandard layoutProfile = "standard"
	layoutWide     layoutProfile = "wide"
)

type layoutSpec struct {
	profile layoutProfile

	endpointWidth  int
	endpointHeight int
	statsWidth     int
	statsHeight    int
	headersWidth   int
	headersHeight  int
	logsWidth      int
	logsHeight     int

	showStats bool
}

// StatsProvider interface for getting statistics
// and endpoint startup state.
type StatsProvider interface {
	GetStats() (ttl, opn int, rt1, rt5, p50, p90 float64)
	GetEndpointState() model.EndpointState
}

// Model represents the TUI application state
type Model struct {
	endpointPane viewport.Model
	statsPane    viewport.Model
	headersPane  viewport.Model
	appLogs      viewport.Model

	width       int
	height      int
	layout      layoutSpec
	appLogLines []string
	lastRequest *model.RequestLog
	ready       bool
	server      StatsProvider
}

// Message types for TUI updates
type LogMsg struct {
	Level   string
	Message string
	Time    time.Time
}

// RequestMsg is the correct message type for request updates
type RequestMsg struct {
	Log model.RequestLog
}

type tickMsg struct{}

// NewModel creates a new TUI model
func NewModel(server StatsProvider) Model {
	return Model{
		endpointPane: viewport.New(0, 0),
		statsPane:    viewport.New(0, 0),
		headersPane:  viewport.New(0, 0),
		appLogs:      viewport.New(0, 0),
		appLogLines:  []string{},
		server:       server,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.applyWindowSize(msg.Width, msg.Height)

	case tickMsg:
		if m.ready {
			m.refreshPaneContent()
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg{}
		})

	case LogMsg:
		m.appendLog(msg)

	case RequestMsg:
		m.lastRequest = &msg.Log
		if m.ready {
			m.updateHeadersPane()
			m.updateStatsPane()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k", "down", "j", "pgup", "pgdown":
			if m.ready {
				m.appLogs, _ = m.appLogs.Update(msg)
			}
			return m, nil
		}
	}

	if m.ready {
		m.appLogs, _ = m.appLogs.Update(msg)
	}

	return m, nil
}

func (m *Model) applyWindowSize(width, height int) {
	m.width = width
	m.height = height
	m.layout = calculateLayout(width, height)

	if !m.ready {
		m.endpointPane = viewport.New(m.layout.endpointWidth, m.layout.endpointHeight)
		m.statsPane = viewport.New(m.layout.statsWidth, m.layout.statsHeight)
		m.headersPane = viewport.New(m.layout.headersWidth, m.layout.headersHeight)
		m.appLogs = viewport.New(m.layout.logsWidth, m.layout.logsHeight)
		m.ready = true
	} else {
		m.endpointPane.Width = m.layout.endpointWidth
		m.endpointPane.Height = m.layout.endpointHeight
		m.statsPane.Width = m.layout.statsWidth
		m.statsPane.Height = m.layout.statsHeight
		m.headersPane.Width = m.layout.headersWidth
		m.headersPane.Height = m.layout.headersHeight
		m.appLogs.Width = m.layout.logsWidth
		m.appLogs.Height = m.layout.logsHeight
	}

	m.refreshPaneContent()
}

func calculateLayout(width, height int) layoutSpec {
	if width < 60 {
		width = 60
	}
	if height < 20 {
		height = 20
	}

	profile := selectLayoutProfile(width, height)
	fullWidth := maxInt(width-4, 40)
	availableHeight := contentHeightBudget(height, profile)

	spec := layoutSpec{
		profile:       profile,
		endpointWidth: fullWidth,
		logsWidth:     fullWidth,
	}

	switch profile {
	case layoutCompact:
		spec.endpointHeight = 6
		spec.logsHeight = maxInt(8, (availableHeight*45)/100)
		if spec.logsHeight > availableHeight-6 {
			spec.logsHeight = availableHeight - 6
		}

		remaining := availableHeight - spec.endpointHeight - spec.logsHeight
		if remaining >= 12 {
			spec.headersHeight = (remaining * 2) / 3
			spec.statsHeight = remaining - spec.headersHeight
			spec.showStats = spec.statsHeight >= 4
		} else if remaining >= 6 {
			spec.headersHeight = remaining
			spec.showStats = false
		} else {
			spec.headersHeight = 0
			spec.statsHeight = 0
			spec.showStats = false
			spec.logsHeight += maxInt(remaining, 0)
		}

		spec.headersWidth = fullWidth
		spec.statsWidth = fullWidth

	case layoutWide:
		spec.endpointWidth = maxInt((fullWidth*72)/100, 88)
		spec.endpointHeight = 4
		middleHeight := maxInt(12, (availableHeight*50)/100)
		spec.logsHeight = availableHeight - spec.endpointHeight - middleHeight
		if spec.logsHeight < 10 {
			shift := 10 - spec.logsHeight
			middleHeight -= shift
			spec.logsHeight = 10
			if middleHeight < 10 {
				middleHeight = 10
			}
		}

		innerWidth := maxInt(fullWidth-2, 72)
		statsWidth := (innerWidth * 30) / 100
		if statsWidth < 34 {
			statsWidth = 34
		}
		headersWidth := innerWidth - statsWidth
		if headersWidth < 44 {
			headersWidth = 44
			statsWidth = innerWidth - headersWidth
		}

		spec.statsWidth = statsWidth
		spec.headersWidth = headersWidth
		spec.statsHeight = middleHeight
		spec.headersHeight = middleHeight
		spec.showStats = true

	default:
		spec.endpointWidth = maxInt((fullWidth*72)/100, 88)
		spec.endpointHeight = 4
		middleHeight := maxInt(10, (availableHeight*50)/100)
		spec.logsHeight = availableHeight - spec.endpointHeight - middleHeight
		if spec.logsHeight < 8 {
			shift := 8 - spec.logsHeight
			middleHeight -= shift
			spec.logsHeight = 8
			if middleHeight < 8 {
				middleHeight = 8
			}
		}

		innerWidth := maxInt(fullWidth-2, 64)
		statsWidth := (innerWidth * 38) / 100
		if statsWidth < 30 {
			statsWidth = 30
		}
		headersWidth := innerWidth - statsWidth
		if headersWidth < 38 {
			headersWidth = 38
			statsWidth = innerWidth - headersWidth
		}

		spec.statsWidth = statsWidth
		spec.headersWidth = headersWidth
		spec.statsHeight = middleHeight
		spec.headersHeight = middleHeight
		spec.showStats = true
	}

	if spec.endpointHeight < 5 {
		spec.endpointHeight = 5
	}
	if spec.logsHeight < 6 {
		spec.logsHeight = 6
	}

	return spec
}

func contentHeightBudget(totalHeight int, profile layoutProfile) int {
	// Layout uses titled, bordered sections and a footer.
	// We budget only inner viewport heights here.
	switch profile {
	case layoutCompact:
		// endpoint + request + stats + logs rows + footer
		// Include one extra safety row to avoid top clipping in terminals that
		// soft-wrap border-adjacent content.
		if totalHeight-14 < 12 {
			return 12
		}
		return totalHeight - 14
	default:
		// endpoint + middle + logs rows + footer
		// Include one extra safety row to avoid top clipping at startup.
		if totalHeight-11 < 12 {
			return 12
		}
		return totalHeight - 11
	}
}

func selectLayoutProfile(width, height int) layoutProfile {
	if width < compactWidthThreshold || height < compactHeightThreshold {
		return layoutCompact
	}
	if width >= wideWidthThreshold && height >= wideHeightThreshold {
		return layoutWide
	}
	return layoutStandard
}

func (m *Model) refreshPaneContent() {
	m.updateEndpointPane()
	m.updateStatsPane()
	m.updateHeadersPane()
	m.appLogs.SetContent(m.renderLogsContent())
}

func (m *Model) appendLog(msg LogMsg) {
	timestamp := msg.Time.Format("15:04:05")
	levelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	switch msg.Level {
	case "ERROR", "FATAL":
		levelStyle = levelStyle.Foreground(lipgloss.Color("196"))
	case "WARN":
		levelStyle = levelStyle.Foreground(lipgloss.Color("208"))
	case "INFO":
		levelStyle = levelStyle.Foreground(lipgloss.Color("34"))
	case "DEBUG":
		levelStyle = levelStyle.Foreground(lipgloss.Color("75"))
	}

	logLine := fmt.Sprintf("%s %s %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(timestamp),
		levelStyle.Render(fmt.Sprintf("%-5s", msg.Level)),
		msg.Message)

	m.appLogLines = append(m.appLogLines, logLine)
	if len(m.appLogLines) > 1000 {
		m.appLogLines = m.appLogLines[1:]
	}

	if m.ready {
		m.appLogs.SetContent(m.renderLogsContent())
		m.appLogs.GotoBottom()
	}
}

func (m *Model) renderLogsContent() string {
	if len(m.appLogLines) == 0 {
		return ""
	}

	maxWidth := m.appLogs.Width
	if maxWidth <= 0 {
		return strings.Join(m.appLogLines, "\n")
	}

	// Keep one column of headroom to avoid terminal hard-wrap at exact pane width.
	displayWidth := maxInt(maxWidth-1, 8)

	lines := make([]string, len(m.appLogLines))
	for i, line := range m.appLogLines {
		lines[i] = ansi.Truncate(line, displayWidth, "...")
	}
	return strings.Join(lines, "\n")
}

func (m *Model) updateEndpointPane() {
	if m.server == nil {
		return
	}

	state := m.server.GetEndpointState()
	contentWidth := maxInt(m.endpointPane.Width-2, 24)
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Mode: %s  Exposure: %s\n",
		fallbackString(strings.TrimSpace(state.Mode), "unknown"),
		exposureLabel(strings.TrimSpace(state.Exposure)),
	))

	b.WriteString("Service: ")
	for _, line := range formatURLForDisplay(state.ServiceURL, contentWidth-9, 1) {
		b.WriteString(line)
	}
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("Web UI: %s", fallbackString(strings.TrimSpace(state.WebUIStatus), "unavailable")))
	if strings.TrimSpace(state.WebUIURL) != "" {
		b.WriteString("  ")
		for _, line := range formatURLForDisplay(state.WebUIURL, contentWidth-10, 1) {
			b.WriteString(line)
		}
	}
	b.WriteString("\n")
	if strings.TrimSpace(state.WebUIReason) != "" {
		b.WriteString(fmt.Sprintf("Web UI Reason: %s\n", state.WebUIReason))
	}

	m.endpointPane.SetContent(b.String())
}

// updateStatsPane updates the statistics pane content
func (m *Model) updateStatsPane() {
	if m.server == nil {
		return
	}

	ttl, opn, rt1, rt5, p50, p90 := m.server.GetStats()

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Connection Statistics"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("%-12s %5s %5s %6s %6s %6s %6s\n",
		"Connections", "ttl", "opn", "rt1", "rt5", "p50", "p90"))
	b.WriteString(strings.Repeat("-", 55) + "\n")
	b.WriteString(fmt.Sprintf("%-12s %5d %5d %6.1f %6.1f %6.1f %6.1f\n\n",
		"", ttl, opn, rt1, rt5, p50, p90))

	b.WriteString("Legend:\n")
	b.WriteString("  ttl: Total requests\n")
	b.WriteString("  opn: Open connections\n")
	b.WriteString("  rt1: Avg response time 1m (ms)\n")
	b.WriteString("  rt5: Avg response time 5m (ms)\n")
	b.WriteString("  p50: 50th percentile (ms)\n")
	b.WriteString("  p90: 90th percentile (ms)\n")

	m.statsPane.SetContent(b.String())
}

// updateHeadersPane updates the headers pane content
func (m *Model) updateHeadersPane() {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Latest Request"))
	b.WriteString("\n\n")

	if m.lastRequest == nil {
		b.WriteString("No requests yet...")
		m.headersPane.SetContent(b.String())
		return
	}

	statusColor := lipgloss.Color("34")
	if m.lastRequest.Response.StatusCode >= 400 {
		statusColor = lipgloss.Color("196")
	} else if m.lastRequest.Response.StatusCode >= 300 {
		statusColor = lipgloss.Color("208")
	}

	lineWidth := maxInt(m.headersPane.Width-4, 32)
	headerValueLimit := maxInt(lineWidth-18, 24)

	b.WriteString(fmt.Sprintf("%s %s\n",
		lipgloss.NewStyle().Bold(true).Render(m.lastRequest.Method),
		truncateString(m.lastRequest.URL, lineWidth)))

	b.WriteString(fmt.Sprintf("Status: %s  Duration: %s\n",
		lipgloss.NewStyle().Foreground(statusColor).Render(fmt.Sprintf("%d", m.lastRequest.Response.StatusCode)),
		m.lastRequest.Duration.Round(time.Millisecond).String()))

	b.WriteString(fmt.Sprintf("From: %s\n", truncateString(m.lastRequest.RemoteAddr, lineWidth)))
	b.WriteString(fmt.Sprintf("Time: %s\n\n", m.lastRequest.Timestamp.Format("15:04:05")))

	if len(m.lastRequest.Headers) > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Request Headers:"))
		b.WriteString("\n")

		priorityHeaders := []string{"User-Agent", "Content-Type", "Authorization", "Accept", "Host", "Accept-Encoding"}
		shown := make(map[string]bool)
		for _, key := range priorityHeaders {
			if value, exists := m.lastRequest.Headers[key]; exists {
				b.WriteString(fmt.Sprintf("  %s: %s\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render(key),
					truncateString(value, headerValueLimit)))
				shown[key] = true
			}
		}

		var otherHeaders []string
		for k := range m.lastRequest.Headers {
			if !shown[k] {
				otherHeaders = append(otherHeaders, k)
			}
		}
		sort.Strings(otherHeaders)

		currentLines := 7 + len(shown)
		availableLines := m.headersPane.Height - currentLines - 3
		if availableLines < 0 {
			availableLines = 0
		}

		for i, k := range otherHeaders {
			if i >= availableLines {
				remaining := len(otherHeaders) - i
				if remaining > 0 {
					b.WriteString(fmt.Sprintf("  %s\n",
						lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
							fmt.Sprintf("... and %d more headers", remaining))))
				}
				break
			}
			b.WriteString(fmt.Sprintf("  %s: %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render(k),
				truncateString(m.lastRequest.Headers[k], headerValueLimit)))
		}
		b.WriteString("\n")
	}

	if m.lastRequest.Body != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Request Body:"))
		b.WriteString("\n")

		currentLines := strings.Count(b.String(), "\n")
		availableLines := m.headersPane.Height - currentLines - 2
		maxBodyChars := maxInt(availableLines*lineWidth, 160)

		if len(m.lastRequest.Body) > maxBodyChars {
			b.WriteString(fmt.Sprintf("[%d bytes - showing first %d chars]\n", len(m.lastRequest.Body), maxBodyChars))
			bodyPreview := m.lastRequest.Body[:maxBodyChars]
			if lastNewline := strings.LastIndex(bodyPreview, "\n"); lastNewline > maxBodyChars-100 {
				bodyPreview = bodyPreview[:lastNewline]
			} else if lastSpace := strings.LastIndex(bodyPreview, " "); lastSpace > maxBodyChars-50 {
				bodyPreview = bodyPreview[:lastSpace]
			}
			b.WriteString(bodyPreview)
			b.WriteString("\n...")
		} else {
			b.WriteString(m.lastRequest.Body)
		}
		b.WriteString("\n")
	}

	m.headersPane.SetContent(b.String())
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if maxLen <= 3 {
		if len(s) <= maxLen {
			return s
		}
		return s[:maxLen]
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// View renders the TUI
func (m Model) View() string {
	if !m.ready {
		return "Initializing TUI..."
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	endpointSection := lipgloss.JoinVertical(lipgloss.Top,
		titleStyle.Render("Endpoint Summary"),
		panelStyle.Width(m.layout.endpointWidth).Height(m.layout.endpointHeight).Render(m.endpointPane.View()),
	)

	requestSection := ""
	if m.layout.headersHeight > 0 {
		requestSection = lipgloss.JoinVertical(lipgloss.Top,
			titleStyle.Render("Request Details"),
			panelStyle.Width(m.layout.headersWidth).Height(m.layout.headersHeight).Render(m.headersPane.View()),
		)
	}

	statsSection := ""
	if m.layout.showStats {
		statsSection = lipgloss.JoinVertical(lipgloss.Top,
			titleStyle.Render("Statistics"),
			panelStyle.Width(m.layout.statsWidth).Height(m.layout.statsHeight).Render(m.statsPane.View()),
		)
	}

	logsSection := lipgloss.JoinVertical(lipgloss.Top,
		titleStyle.Render("Application Logs"),
		panelStyle.Width(m.layout.logsWidth).Height(m.layout.logsHeight).Render(m.appLogs.View()),
	)

	mainSections := []string{endpointSection}

	switch m.layout.profile {
	case layoutCompact:
		if requestSection != "" {
			mainSections = append(mainSections, requestSection)
		}
		if statsSection != "" {
			mainSections = append(mainSections, statsSection)
		}
		mainSections = append(mainSections, logsSection)
	default:
		middle := ""
		if statsSection != "" && requestSection != "" {
			middle = lipgloss.JoinHorizontal(lipgloss.Top, statsSection, requestSection)
		} else if requestSection != "" {
			middle = requestSection
		} else if statsSection != "" {
			middle = statsSection
		}

		if middle != "" {
			mainSections = append(mainSections, middle)
		}
		mainSections = append(mainSections, logsSection)
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Press 'q' or Ctrl+C to quit | Up/Down or j/k to scroll logs | PgUp/PgDn for faster scrolling")

	mainView := lipgloss.JoinVertical(lipgloss.Top, mainSections...)
	final := lipgloss.JoinVertical(lipgloss.Top, mainView, footer)
	return sanitizeViewToWindow(strings.TrimRight(final, "\n"), m.width, m.height)
}

func exposureLabel(exposure string) string {
	switch exposure {
	case "tailnet":
		return "tailnet-private"
	case "funnel":
		return "funnel-public"
	default:
		return fallbackString(exposure, "unknown")
	}
}

func formatURLForDisplay(raw string, maxWidth, maxLines int) []string {
	if maxWidth < 16 {
		maxWidth = 16
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{"unavailable"}
	}

	if maxLines <= 0 {
		maxLines = 1
	}

	if ansi.StringWidth(trimmed) <= maxWidth {
		return []string{trimmed}
	}

	// Preserve start of URL and enforce single-line deterministic truncation.
	parsed, err := url.Parse(trimmed)
	if err == nil && parsed != nil && parsed.Scheme != "" && parsed.Host != "" {
		prefix := parsed.Scheme + "://" + parsed.Host
		if ansi.StringWidth(prefix) > maxWidth-3 {
			return []string{ansi.Truncate(trimmed, maxWidth, "...")}
		}
		pathPrefix := parsed.EscapedPath()
		if pathPrefix == "" {
			pathPrefix = "/"
		}
		candidate := prefix + pathPrefix
		if parsed.RawQuery != "" {
			candidate += "?" + parsed.RawQuery
		}
		if parsed.Fragment != "" {
			candidate += "#" + parsed.Fragment
		}
		return []string{ansi.Truncate(candidate, maxWidth, "...")}
	}

	return []string{ansi.Truncate(trimmed, maxWidth, "...")}
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sanitizeViewToWindow(view string, width, height int) string {
	if view == "" {
		return view
	}

	lines := strings.Split(view, "\n")

	if width > 0 {
		maxWidth := maxInt(width, 1)
		for i, line := range lines {
			lines[i] = ansi.Truncate(line, maxWidth, "")
		}
	}

	if height > 0 && len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

// LogWriter implements io.Writer to capture zap logs for the TUI
type LogWriter struct {
	program *tea.Program
}

// NewLogWriter creates a new log writer for TUI integration
func NewLogWriter(program *tea.Program) *LogWriter {
	return &LogWriter{program: program}
}

// Write implements io.Writer
func (w *LogWriter) Write(p []byte) (n int, err error) {
	line := string(p)
	parts := strings.SplitN(line, "\t", 4)
	if len(parts) >= 3 {
		level := strings.TrimSpace(parts[1])
		message := strings.TrimSpace(parts[2])
		if len(parts) > 3 {
			message += " " + strings.TrimSpace(parts[3])
		}

		w.program.Send(LogMsg{
			Level:   level,
			Message: message,
			Time:    time.Now(),
		})
	}
	return len(p), nil
}

// CreateRequestMsg creates a request message for the TUI
func CreateRequestMsg(log model.RequestLog) tea.Msg {
	return RequestMsg{Log: log}
}
