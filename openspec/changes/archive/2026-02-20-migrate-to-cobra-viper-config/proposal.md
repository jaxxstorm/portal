## Why

tgate currently relies on CLI-only configuration with `alecthomas/kong`, which limits 12-factor deployment patterns and makes containerized and sidecar usage harder to operate consistently. Moving to Cobra/Viper enables file-based and environment-based configuration while preserving the current CLI interface for existing users.

## What Changes

- Replace `alecthomas/kong` command/flag parsing with Cobra while keeping existing command shape and flags stable (for example, `tgate 8080` and `tgate 8080 --funnel` remain valid).
- Introduce configuration loading via Viper from `~/.tgate/config.yml`.
- Add environment variable overrides using the `TGATE_` prefix for all supported configuration values (for example, `TGATE_FUNNEL=true`).
- Define precedence and merge behavior across CLI args, env vars, and config file to support local development, CI, Docker, and sidecar deployment models.
- Keep tailnet-private default behavior and explicit Funnel opt-in behavior unchanged while improving how these modes are configured.

## Capabilities

### New Capabilities
- `runtime-configuration`: Unified configuration system for CLI arguments, config file (`~/.tgate/config.yml`), and `TGATE_*` environment variables with deterministic precedence.

### Modified Capabilities
- `local-service-exposure`: Exposure modes remain the same, but requirements are expanded to guarantee compatibility of existing CLI usage while allowing equivalent file/env configuration.

## Impact

- Affected code: `internal/config`, CLI entrypoint wiring, startup configuration validation, and related tests.
- Affected dependencies: remove primary reliance on `alecthomas/kong`; add/standardize Cobra and Viper usage.
- Affected behavior surface: startup configuration source resolution and validation errors; no intended breaking change to existing command syntax.
- Operational impact: improved support for Docker images, sidecar-style deployments, and environment-driven automation.
