package tailscale

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestTSNetRuntimeLogsUsePrimaryLoggerWithStructuredContext(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	server := NewTSNetServer(TSNetConfig{Hostname: "tgate"}, logger)
	observed.TakeAll() // Drop construction log.

	server.server.Logf("AuthLoop: state is Running; done")

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("expected exactly one runtime log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Level != zapcore.InfoLevel {
		t.Fatalf("expected info level, got %s", entry.Level)
	}
	if entry.Message != "AuthLoop: state is Running; done" {
		t.Fatalf("unexpected message: %q", entry.Message)
	}

	fields := entry.ContextMap()
	assertStringField(t, fields, "component", "tsnet_runtime")
	assertStringField(t, fields, "tailscale_mode", "tsnet")
	assertStringField(t, fields, "node_name", "tgate")
	assertStringField(t, fields, "phase", "auth")
}

func TestTSNetRuntimeLogLevelAndPhaseMapping(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	server := NewTSNetServer(TSNetConfig{Hostname: "tgate"}, logger)
	observed.TakeAll() // Drop construction log.

	server.server.Logf("Authkey is set; but state is NoState. Ignoring authkey.")
	server.server.Logf("Funnel setup failed: certificate check failed")

	entries := observed.All()
	if len(entries) != 2 {
		t.Fatalf("expected two runtime log entries, got %d", len(entries))
	}

	if entries[0].Level != zapcore.WarnLevel {
		t.Fatalf("expected warn level for authkey warning, got %s", entries[0].Level)
	}
	assertStringField(t, entries[0].ContextMap(), "phase", "auth")

	if entries[1].Level != zapcore.ErrorLevel {
		t.Fatalf("expected error level for funnel failure, got %s", entries[1].Level)
	}
	assertStringField(t, entries[1].ContextMap(), "phase", "funnel")
}

func TestTSNetRuntimeLogAdapterSkipsEmptyMessages(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	server := NewTSNetServer(TSNetConfig{Hostname: "tgate"}, logger)
	observed.TakeAll() // Drop construction log.

	server.server.Logf("   ")

	if got := len(observed.All()); got != 0 {
		t.Fatalf("expected no entries for empty messages, got %d", got)
	}
}

func TestTSNetRuntimeInfoLogsSuppressedWithoutVerbose(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	server := NewTSNetServer(TSNetConfig{Hostname: "tgate"}, logger)
	observed.TakeAll() // Drop construction log.

	server.server.Logf("wgengine: Reconfig: configuring userspace WireGuard config")

	if got := len(observed.All()); got != 0 {
		t.Fatalf("expected no runtime info logs without verbose mode, got %d", got)
	}
}

func TestTSNetRuntimeWarnAndErrorLogsSuppressedWithoutVerbose(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	server := NewTSNetServer(TSNetConfig{Hostname: "tgate"}, logger)
	observed.TakeAll() // Drop construction log.

	server.server.Logf("Authkey is set; but state is NoState. Ignoring authkey.")
	server.server.Logf("health(warnable=warming-up): error: Tailscale is starting. Please wait.")

	if got := len(observed.All()); got != 0 {
		t.Fatalf("expected no runtime warn/error logs without verbose mode, got %d", got)
	}
}

func TestTSNetUserLogsUseAdapterAndDoNotLeak(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	server := NewTSNetServer(TSNetConfig{Hostname: "tgate"}, logger)
	observed.TakeAll() // Drop construction log.

	if server.server.UserLogf == nil {
		t.Fatalf("expected tsnet user logger hook to be configured")
	}

	server.server.UserLogf("AuthLoop: state is Running; done")

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("expected one user log entry, got %d", len(entries))
	}
	assertStringField(t, entries[0].ContextMap(), "component", "tsnet_runtime")
	assertStringField(t, entries[0].ContextMap(), "phase", "auth")
}

func assertStringField(t *testing.T, fields map[string]interface{}, key, want string) {
	t.Helper()

	got, ok := fields[key]
	if !ok {
		t.Fatalf("missing %q field in %+v", key, fields)
	}

	gotString, ok := got.(string)
	if !ok {
		t.Fatalf("field %q is not a string: %#v", key, got)
	}

	if gotString != want {
		t.Fatalf("unexpected %q value: got %q want %q", key, gotString, want)
	}
}
