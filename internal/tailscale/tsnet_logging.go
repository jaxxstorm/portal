package tailscale

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/jaxxstorm/portal/internal/logging"
)

func newTSNetRuntimeLogAdapter(logger *zap.Logger, nodeName string) func(format string, args ...any) {
	verboseEnabled := logger.Core().Enabled(zapcore.DebugLevel)
	runtimeLogger := logger.WithOptions(zap.AddStacktrace(zapcore.PanicLevel))

	baseFields := []zap.Field{
		logging.Component("tsnet_runtime"),
		logging.TailscaleMode("tsnet"),
	}
	if nodeName != "" {
		baseFields = append(baseFields, logging.NodeName(nodeName))
	}

	return func(format string, args ...any) {
		message := strings.TrimSpace(fmt.Sprintf(format, args...))
		if message == "" {
			return
		}

		// Underlying tsnet runtime logs are debug-only noise for normal runs.
		// Surface them only when verbose mode enables debug-level logging.
		if !verboseEnabled {
			return
		}

		level := classifyTSNetRuntimeLogLevel(message)
		fields := append(baseFields, logging.Phase(classifyTSNetRuntimePhase(message)))

		switch level {
		case zapcore.ErrorLevel:
			runtimeLogger.Error(message, fields...)
		case zapcore.WarnLevel:
			runtimeLogger.Warn(message, fields...)
		case zapcore.DebugLevel:
			runtimeLogger.Debug(message, fields...)
		default:
			runtimeLogger.Info(message, fields...)
		}
	}
}

func classifyTSNetRuntimeLogLevel(message string) zapcore.Level {
	lower := strings.ToLower(message)

	switch {
	case containsAny(lower, "panic", "fatal", "error", "failed", "failure", "unable", "cannot", "denied", "timed out", "timeout"):
		return zapcore.ErrorLevel
	case containsAny(lower, "warning", "warn:", "[warn]", "ignoring", "retrying"):
		return zapcore.WarnLevel
	case containsAny(lower, "[v1]", "[v2]", "[v3]", "debug"):
		return zapcore.DebugLevel
	default:
		return zapcore.InfoLevel
	}
}

func classifyTSNetRuntimePhase(message string) string {
	lower := strings.ToLower(message)

	switch {
	case containsAny(lower, "funnel", "public"):
		return "funnel"
	case containsAny(lower, "cert", "https", "tls"):
		return "certificate"
	case containsAny(lower, "auth", "login", "oauth"):
		return "auth"
	case containsAny(lower, "listen", "listener", "bind"):
		return "listener"
	case containsAny(lower, "serve", "http"):
		return "serve"
	case containsAny(lower, "state", "control", "tailnet", "up", "down"):
		return "state"
	default:
		return "runtime"
	}
}

func containsAny(value string, tokens ...string) bool {
	for _, token := range tokens {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}
