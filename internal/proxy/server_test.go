package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"testing"

	"go.uber.org/zap"

	"github.com/jaxxstorm/tgate/internal/model"
)

func TestServeHTTPTailnetModeIgnoresFunnelAllowlist(t *testing.T) {
	server := NewServer(Config{
		Mode:            model.ModeMock,
		UseTUI:          true,
		Logger:          zap.NewNop(),
		FunnelEnabled:   false,
		FunnelAllowlist: mustPrefixes(t, "203.0.113.0/24"),
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.99:12345"
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected tailnet request to bypass allowlist, got status %d", rr.Code)
	}
}

func TestServeHTTPFunnelModeAllowsAllowlistedForwardedSource(t *testing.T) {
	server := NewServer(Config{
		Mode:            model.ModeMock,
		UseTUI:          true,
		Logger:          zap.NewNop(),
		FunnelEnabled:   true,
		FunnelAllowlist: mustPrefixes(t, "203.0.113.0/24"),
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected allowlisted request, got status %d", rr.Code)
	}
}

func TestServeHTTPFunnelModeAllowsAllowlistedRemoteAddrFallback(t *testing.T) {
	server := NewServer(Config{
		Mode:            model.ModeMock,
		UseTUI:          true,
		Logger:          zap.NewNop(),
		FunnelEnabled:   true,
		FunnelAllowlist: mustPrefixes(t, "198.51.100.12/32"),
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.12:12345"
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected remote-addr allowlisted request, got status %d", rr.Code)
	}
}

func TestServeHTTPFunnelModeDeniesNonAllowlistedSource(t *testing.T) {
	server := NewServer(Config{
		Mode:            model.ModeMock,
		UseTUI:          true,
		Logger:          zap.NewNop(),
		FunnelEnabled:   true,
		FunnelAllowlist: mustPrefixes(t, "203.0.113.0/24"),
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:12345"
	req.Header.Set("X-Forwarded-For", "192.0.2.9")
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected non-allowlisted request to be denied, got status %d", rr.Code)
	}
}

func TestServeHTTPFunnelModeDeniesUnresolvedSource(t *testing.T) {
	server := NewServer(Config{
		Mode:            model.ModeMock,
		UseTUI:          true,
		Logger:          zap.NewNop(),
		FunnelEnabled:   true,
		FunnelAllowlist: mustPrefixes(t, "203.0.113.0/24"),
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "not-a-valid-source"
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected unresolved request to be denied, got status %d", rr.Code)
	}
}

func TestResolveSourceIPPrioritizesTailscaleClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.10:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.25")
	req.Header.Set("Tailscale-Client-IP", "192.0.2.4")

	addr, signal, ok := resolveSourceIP(req, false)

	if !ok {
		t.Fatalf("expected source IP resolution to succeed")
	}
	if got, want := signal, sourceSignalTailscaleClientIP; got != want {
		t.Fatalf("expected source signal %q, got %q", want, got)
	}
	if got, want := addr.String(), "192.0.2.4"; got != want {
		t.Fatalf("expected source IP %q, got %q", want, got)
	}
}

func TestResolveSourceIPPrefersRemoteAddrWhenConfigured(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.10:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.25")
	req.Header.Set("Tailscale-Client-IP", "192.0.2.4")

	addr, signal, ok := resolveSourceIP(req, true)

	if !ok {
		t.Fatalf("expected source IP resolution to succeed")
	}
	if got, want := signal, sourceSignalRemoteAddr; got != want {
		t.Fatalf("expected source signal %q, got %q", want, got)
	}
	if got, want := addr.String(), "198.51.100.10"; got != want {
		t.Fatalf("expected source IP %q, got %q", want, got)
	}
}

func TestServeHTTPFunnelModePrefersRemoteAddrWhenConfigured(t *testing.T) {
	server := NewServer(Config{
		Mode:            model.ModeMock,
		UseTUI:          true,
		Logger:          zap.NewNop(),
		FunnelEnabled:   true,
		FunnelAllowlist: mustPrefixes(t, "198.51.100.0/24"),
		PreferRemoteIP:  true,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.10:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected remote allowlisted request, got status %d", rr.Code)
	}
}

func TestServeHTTPNonTUIModeDoesNotWriteLegacyConsoleOutput(t *testing.T) {
	server := NewServer(Config{
		Mode:          model.ModeMock,
		UseTUI:        false,
		Logger:        zap.NewNop(),
		FunnelEnabled: false,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.12:12345"
	rr := httptest.NewRecorder()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	os.Stdout = writer
	server.ServeHTTP(rr, req)

	if closeErr := writer.Close(); closeErr != nil {
		t.Fatalf("failed to close stdout writer: %v", closeErr)
	}
	os.Stdout = oldStdout
	defer reader.Close()

	var captured bytes.Buffer
	if _, copyErr := io.Copy(&captured, reader); copyErr != nil {
		t.Fatalf("failed to read captured stdout: %v", copyErr)
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if got := captured.String(); got != "" {
		t.Fatalf("expected no legacy stdout request logs, got %q", got)
	}
}

func mustPrefixes(t *testing.T, values ...string) []netip.Prefix {
	t.Helper()

	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			t.Fatalf("failed to parse prefix %q: %v", value, err)
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}
