package ui

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jaxxstorm/portal/internal/model"
)

type stubLogProvider struct {
	cleared bool
}

func (s *stubLogProvider) GetRequestLogs() []model.RequestLog {
	return nil
}

func (s *stubLogProvider) GetStats() (ttl, opn int, rt1, rt5, p50, p90 float64) {
	return 0, 0, 0, 0, 0, 0
}

func (s *stubLogProvider) ClearRequestLogs() {
	s.cleared = true
}

func testServerWithUIFiles(t *testing.T, provider LogProvider) *Server {
	t.Helper()

	root := t.TempDir()
	uiDir := filepath.Join(root, "ui")
	if err := os.MkdirAll(uiDir, 0o755); err != nil {
		t.Fatalf("mkdir ui dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(uiDir, "index.html"), []byte("<html>ok</html>"), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}
	if err := os.WriteFile(filepath.Join(uiDir, "app.js"), []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}

	return NewServer(provider, os.DirFS(root))
}

func TestHandleAPIOptionsIncludesDeleteAndSameOrigin(t *testing.T) {
	srv := testServerWithUIFiles(t, nil)

	req := httptest.NewRequest(http.MethodOptions, "/api/requests", nil)
	req.Host = "portal.example.ts.net:8443"
	req.Header.Set("Origin", "http://portal.example.ts.net:8443")
	rr := httptest.NewRecorder()

	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, DELETE, OPTIONS" {
		t.Fatalf("unexpected allow methods: %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://portal.example.ts.net:8443" {
		t.Fatalf("unexpected allow origin: %q", got)
	}
}

func TestHandleAPIDoesNotAllowCrossOrigin(t *testing.T) {
	srv := testServerWithUIFiles(t, nil)

	req := httptest.NewRequest(http.MethodOptions, "/api/requests", nil)
	req.Host = "portal.example.ts.net"
	req.Header.Set("Origin", "http://other.example.ts.net")
	rr := httptest.NewRecorder()

	srv.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no allow-origin header, got %q", got)
	}
}

func TestHandleAPIDeleteRequestsClearsLogs(t *testing.T) {
	provider := &stubLogProvider{}
	srv := testServerWithUIFiles(t, provider)

	req := httptest.NewRequest(http.MethodDelete, "/api/requests", nil)
	rr := httptest.NewRecorder()

	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
	if !provider.cleared {
		t.Fatalf("expected provider logs to be cleared")
	}
}

func TestHandleStaticServesUIPrefixedAsset(t *testing.T) {
	srv := testServerWithUIFiles(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/ui/app.js", nil)
	rr := httptest.NewRecorder()

	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "application/javascript") {
		t.Fatalf("unexpected content type: %q", got)
	}
	body, err := io.ReadAll(rr.Result().Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(body) != "console.log('ok')" {
		t.Fatalf("unexpected body: %q", string(body))
	}
}

func TestHandleStaticRedirectsUIWithoutTrailingSlash(t *testing.T) {
	srv := testServerWithUIFiles(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/ui", nil)
	rr := httptest.NewRecorder()

	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusMovedPermanently {
		t.Fatalf("expected status %d, got %d", http.StatusMovedPermanently, rr.Code)
	}
	if got := rr.Header().Get("Location"); got != "/ui/" {
		t.Fatalf("unexpected redirect location: %q", got)
	}
}
