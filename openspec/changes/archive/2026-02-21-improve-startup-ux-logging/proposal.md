## Why

tgate startup output is currently fragmented across logs, TUI panels, and stdout, which makes it unclear when startup is complete and where the web UI is reachable (especially with auto-assigned `--ui-port`). We need a single, definitive startup UX across TUI, plain logs, and JSON logs so operators can immediately see endpoints, mode, and enabled capabilities.

## What Changes

- Add a unified startup status summary that is emitted once the service is ready, with consistent core fields across output modes.
- Define mode-specific presentation rules for startup status in TUI, non-TUI standard logging, and non-TUI JSON logging.
- Require explicit startup reporting of service reachability (`tailnet` vs `funnel`), web UI reachability, and enabled capabilities.
- Require explicit status when UI is disabled, unavailable, or auto-assigned to a dynamic local port.
- Preserve existing exposure behavior for `tgate <port>` (tailnet private) and `tgate <port> --funnel` (public via Funnel), while improving clarity of startup observability.
- Update documentation to show concrete startup output expectations and operator interpretation guidance.

## Capabilities

### New Capabilities
- `startup-operator-feedback`: Defines deterministic startup-ready reporting across TUI and non-TUI logging formats, including reachable service/UI endpoints and enabled capabilities.

### Modified Capabilities
- `documentation-system`: Extend docs coverage so startup UX and logging mode behavior are explicitly documented for tailnet and Funnel usage.

## Impact

- Affected code: startup orchestration in `main.go`, local-daemon setup in `internal/server/setup.go`, TUI startup/status presentation, and structured logging emission points.
- Affected docs: `docs/configuration.md`, operating mode docs, and troubleshooting guidance for startup visibility.
- Behavior impact: users running `tgate <port>` and `tgate <port> --funnel` get clearer startup-complete confirmation and endpoint visibility without changing serving semantics.
- Operational impact: easier verification of web UI location and active capabilities in both interactive and headless execution.
