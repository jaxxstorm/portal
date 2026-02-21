## Why

`tgate` currently serves tsnet traffic through port listeners, which limits flexibility for service-oriented exposure and creates mode-specific behavior gaps. `tsnet.Server.ListenService` now provides a native path for service-mode listeners, so we should add first-class support while preserving secure defaults.

## What Changes

- Add optional tsnet service mode that uses `tsnet.Server.ListenService` for tsnet-backed serving.
- Keep tailnet-private as the default behavior for `tgate <port>` and keep Funnel public exposure opt-in for `tgate <port> --funnel`.
- Introduce runtime configuration controls for tsnet service mode selection and service name/value configuration.
- Define validation behavior for invalid or incompatible mode combinations (including service+funnel conflict).
- Add startup observability that clearly indicates whether tsnet is running in listener mode or service mode.
- Update documentation with examples and prerequisites for tsnet service mode in both tailnet and Funnel contexts.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `local-service-exposure`: Extend tsnet exposure behavior to support service-mode listening via `ListenService`.
- `runtime-configuration`: Add and validate configuration keys/env mappings for tsnet service mode and service identifier settings.
- `documentation-system`: Add explicit documentation for tsnet service mode usage, defaults, and constraints.

## Impact

- Affected code: tsnet server setup/listen path, startup mode reporting, config parsing/validation, and mode selection logic.
- Affected docs: operating modes and configuration guides for service mode syntax, examples, and caveats.
- Security impact: no default exposure broadening; `tgate <port>` remains tailnet-private and `tgate <port> --funnel` remains explicit opt-in with existing prerequisite checks.
- Compatibility impact: existing tsnet listener behavior remains supported; service mode is additive and configurable.
