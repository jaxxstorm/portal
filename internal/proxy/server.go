// internal/proxy/server.go
package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"

	"github.com/jaxxstorm/portal/internal/logging"
	"github.com/jaxxstorm/portal/internal/model"
	"github.com/jaxxstorm/portal/internal/stats"
)

// LoggingResponseWriter wraps http.ResponseWriter to capture response information
type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	size          int64
	headers       map[string]string
	bodyPreview   []byte
	bodyTruncated bool
}

const maxResponseBodyPreviewBytes = 256 * 1024

// WriteHeader captures the status code
func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.statusCode == 0 {
		lrw.statusCode = 200
	}

	remaining := maxResponseBodyPreviewBytes - len(lrw.bodyPreview)
	if remaining > 0 {
		if len(b) > remaining {
			lrw.bodyPreview = append(lrw.bodyPreview, b[:remaining]...)
			lrw.bodyTruncated = true
		} else {
			lrw.bodyPreview = append(lrw.bodyPreview, b...)
		}
	} else if len(b) > 0 {
		lrw.bodyTruncated = true
	}

	size, err := lrw.ResponseWriter.Write(b)
	lrw.size += int64(size)
	return size, err
}

// Header returns the response headers
func (lrw *LoggingResponseWriter) Header() http.Header {
	return lrw.ResponseWriter.Header()
}

// captureHeaders captures response headers for logging
func (lrw *LoggingResponseWriter) captureHeaders() {
	lrw.headers = make(map[string]string)
	for k, v := range lrw.ResponseWriter.Header() {
		lrw.headers[k] = strings.Join(v, ", ")
	}
}

// Server handles HTTP requests with logging and optional proxying
type Server struct {
	logger          *zap.Logger
	sugarLogger     *zap.SugaredLogger
	proxy           *httputil.ReverseProxy
	targetURL       *url.URL
	requestLog      []model.RequestLog
	logMutex        sync.RWMutex
	program         *tea.Program
	useTUI          bool
	mode            model.ServerMode
	stats           *stats.Tracker
	requestID       int64
	webUIURL        string                   // Store the web UI URL for display
	maxLogsCap      int                      // Maximum number of logs to keep
	listeners       []func(model.RequestLog) // Event listeners for new requests
	funnelEnabled   bool
	funnelAllowlist []netip.Prefix
	preferRemoteIP  bool
}

// Config holds configuration for the proxy server
type Config struct {
	TargetPort      int
	UseTUI          bool
	Mode            model.ServerMode
	Logger          *zap.Logger
	MaxLogs         int // Maximum number of logs to keep (default: 1000)
	FunnelEnabled   bool
	FunnelAllowlist []netip.Prefix
	PreferRemoteIP  bool
}

// NewServer creates a new proxy server
func NewServer(config Config) *Server {
	var targetURL *url.URL
	var proxy *httputil.ReverseProxy

	maxLogs := config.MaxLogs
	if maxLogs <= 0 {
		maxLogs = 1000 // Default
	}

	if config.Mode == model.ModeProxy {
		targetURL = &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%d", config.TargetPort),
		}

		proxy = httputil.NewSingleHostReverseProxy(targetURL)

		// Customize the director to preserve original headers
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Header.Set("X-Forwarded-Proto", "https")
			req.Header.Set("X-Forwarded-Host", req.Host)
		}
	}

	return &Server{
		logger:          config.Logger,
		sugarLogger:     config.Logger.Sugar(),
		proxy:           proxy,
		targetURL:       targetURL,
		requestLog:      make([]model.RequestLog, 0),
		useTUI:          config.UseTUI,
		mode:            config.Mode,
		stats:           stats.NewTracker(),
		requestID:       0,
		maxLogsCap:      maxLogs,
		listeners:       make([]func(model.RequestLog), 0),
		funnelEnabled:   config.FunnelEnabled,
		funnelAllowlist: config.FunnelAllowlist,
		preferRemoteIP:  config.PreferRemoteIP,
	}
}

// SetProgram sets the TUI program for sending messages
func (s *Server) SetProgram(p *tea.Program) {
	s.program = p
}

// SetWebUIURL stores the web UI URL for display
func (s *Server) SetWebUIURL(url string) {
	s.webUIURL = url
}

// GetWebUIURL returns the stored web UI URL
func (s *Server) GetWebUIURL() string {
	return s.webUIURL
}

// AddListener adds a listener function that will be called for each new request
func (s *Server) AddListener(listener func(model.RequestLog)) {
	if listener != nil {
		s.listeners = append(s.listeners, listener)
	}
}

// ReplaceLogger replaces the current logger with a new one
func (s *Server) ReplaceLogger(logger *zap.Logger) {
	s.logger = logger
	s.sugarLogger = logger.Sugar()
}

// nextRequestID generates a unique request ID
func (s *Server) nextRequestID() string {
	id := atomic.AddInt64(&s.requestID, 1)
	return fmt.Sprintf("req_%d_%d", time.Now().Unix(), id)
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := s.nextRequestID()

	// Track connection stats
	s.stats.IncrementOpen()
	defer s.stats.DecrementOpen()

	// Create logging response writer
	lrw := &LoggingResponseWriter{
		ResponseWriter: w,
		statusCode:     0,
		size:           0,
		headers:        make(map[string]string),
		bodyPreview:    make([]byte, 0),
		bodyTruncated:  false,
	}

	// Read request body for logging (if not too large)
	var bodyBytes []byte
	var bodyString string
	if r.Body != nil && r.ContentLength < 10*1024*1024 { // Limit to 10MB
		bodyBytes, _ = io.ReadAll(r.Body)
		bodyString = string(bodyBytes)
		r.Body = io.NopCloser(strings.NewReader(bodyString))
	}

	// Capture request headers
	reqHeaders := make(map[string]string)
	for k, v := range r.Header {
		reqHeaders[k] = strings.Join(v, ", ")
	}

	// Log application-level events using the same pattern as other components
	s.logger.Info("Request received",
		logging.Component("proxy_server"),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
	)

	if s.enforceFunnelAllowlist(lrw, r) {
		// Handle request based on mode
		switch s.mode {
		case model.ModeMock:
			s.handleMockRequest(lrw, r, bodyString)
		case model.ModeProxy:
			s.proxy.ServeHTTP(lrw, r)
		}
	}
	// Capture response headers after serving
	lrw.captureHeaders()

	duration := time.Since(start)

	// Add to stats
	s.stats.AddRequest(duration)

	// Create request log entry
	logEntry := model.RequestLog{
		ID:          requestID,
		Timestamp:   start,
		Method:      r.Method,
		URL:         r.URL.String(),
		RemoteAddr:  r.RemoteAddr,
		Headers:     reqHeaders,
		Body:        bodyString,
		UserAgent:   r.UserAgent(),
		ContentType: r.Header.Get("Content-Type"),
		Size:        r.ContentLength,
		StatusCode:  lrw.statusCode, // Convenience field for UI
		Response: model.ResponseLog{
			StatusCode:    lrw.statusCode,
			Headers:       lrw.headers,
			Body:          formatResponseBodyPreview(lrw.headers, lrw.bodyPreview),
			BodyTruncated: lrw.bodyTruncated,
			Size:          lrw.size,
		},
		Duration: duration,
	}

	// Store log entry and notify listeners
	s.captureRequest(logEntry)

	// Log application-level response events with proper structured format
	s.logger.Info("Request completed",
		zap.Int("status_code", lrw.statusCode),
		zap.Duration("duration", duration),
		zap.Int64("response_size", lrw.size),
	)

}

func formatResponseBodyPreview(headers map[string]string, preview []byte) string {
	if len(preview) == 0 {
		return ""
	}

	contentType := strings.ToLower(strings.TrimSpace(headers["Content-Type"]))
	textLike := strings.HasPrefix(contentType, "text/") ||
		strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "application/xml") ||
		strings.HasPrefix(contentType, "application/javascript") ||
		strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "+json") ||
		strings.Contains(contentType, "+xml")

	if textLike || utf8.Valid(preview) {
		return string(bytes.ToValidUTF8(preview, []byte("\uFFFD")))
	}

	return "[binary response body omitted]"
}

func (s *Server) enforceFunnelAllowlist(w http.ResponseWriter, r *http.Request) bool {
	if !s.funnelEnabled || len(s.funnelAllowlist) == 0 {
		return true
	}

	sourceIP, sourceSignal, resolved := resolveSourceIP(r, s.preferRemoteIP)
	if !resolved {
		s.logger.Warn("Funnel request denied",
			logging.Component("funnel_allowlist"),
			logging.FunnelEnabled(true),
			zap.String("source_signal", sourceSignal),
			zap.String("deny_reason", "source_ip_unresolved"),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}

	matchedEntry, allowed := allowlistedEntry(sourceIP, s.funnelAllowlist)
	if !allowed {
		s.logger.Warn("Funnel request denied",
			logging.Component("funnel_allowlist"),
			logging.FunnelEnabled(true),
			zap.String("source_signal", sourceSignal),
			zap.String("source_ip", sourceIP.String()),
			zap.String("deny_reason", "source_ip_not_allowlisted"),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}

	s.logger.Info("Funnel request allowed",
		logging.Component("funnel_allowlist"),
		logging.FunnelEnabled(true),
		zap.String("source_signal", sourceSignal),
		zap.String("source_ip", sourceIP.String()),
		zap.String("matched_allowlist_entry", matchedEntry.String()),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	return true
}

// captureRequest stores the log entry and notifies listeners
func (s *Server) captureRequest(logEntry model.RequestLog) {
	// Store log entry
	s.logMutex.Lock()
	s.requestLog = append(s.requestLog, logEntry)
	// Keep only last maxLogsCap requests
	if len(s.requestLog) > s.maxLogsCap {
		s.requestLog = s.requestLog[1:]
	}
	s.logMutex.Unlock()

	// Notify listeners - this is the primary way to send to TUI now
	for _, listener := range s.listeners {
		listener(logEntry)
	}
}

// handleMockRequest handles mock responses for testing
func (s *Server) handleMockRequest(w http.ResponseWriter, r *http.Request, body string) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-portal-mode", "mock")
	w.Header().Set("X-portal-timestamp", time.Now().UTC().Format(time.RFC3339))

	// Create a simple response
	response := map[string]interface{}{
		"status":    "received",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"method":    r.Method,
		"path":      r.URL.Path,
		"headers":   len(r.Header),
		"body_size": len(body),
	}

	// Add query parameters if present
	if len(r.URL.RawQuery) > 0 {
		response["query"] = r.URL.RawQuery
	}

	// Add content type if present
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		response["content_type"] = contentType
	}

	// Return 200 OK with JSON response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetRequestLogs returns a copy of the request logs (implements model.LogProvider)
func (s *Server) GetRequestLogs() []model.RequestLog {
	s.logMutex.RLock()
	defer s.logMutex.RUnlock()

	// Return a copy
	logs := make([]model.RequestLog, len(s.requestLog))
	copy(logs, s.requestLog)
	return logs
}

// GetStats returns current statistics (implements model.StatsProvider)
func (s *Server) GetStats() (ttl, opn int, rt1, rt5, p50, p90 float64) {
	return s.stats.GetStats()
}

// ClearRequestLogs clears captured request history and resets runtime stats.
func (s *Server) ClearRequestLogs() {
	s.logMutex.Lock()
	for i := range s.requestLog {
		s.requestLog[i] = model.RequestLog{}
	}
	s.requestLog = nil
	s.logMutex.Unlock()
	s.stats.Reset()
}

// SendTUIMessage sends a message to the TUI if available (implements model.TUIMessageSender)
func (s *Server) SendTUIMessage(msg interface{}) {
	if s.program != nil {
		s.program.Send(msg)
	}
}
