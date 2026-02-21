## Context

tgate startup information is currently emitted by multiple paths (structured logger, TUI messages, and ad-hoc console lines), so operators do not get a single authoritative "ready" statement. This is most visible when `--ui-port` is auto-assigned: startup can succeed but users are left searching logs to discover the effective web UI address.

The implementation must work consistently across two networking backends:
- local Tailscale daemon mode (`tailscaled` + local serve setup)
- tsnet mode (embedded node runtime)

The implementation must also preserve existing safety semantics:
- `tgate <port>` remains tailnet-private by default
- `tgate <port> --funnel` remains explicit public exposure with HTTPS/Funnel prerequisites
- Funnel setup failures (including HTTPS certificate issues) still fail startup with clear errors

## Goals / Non-Goals

**Goals:**
- Produce one deterministic startup summary payload after tgate is actually ready to receive traffic.
- Surface reachable service URL, web UI URL/status, execution mode (local daemon vs tsnet), exposure mode (tailnet vs funnel), and enabled capabilities.
- Render the same semantic payload in all output modes:
  - TUI: visible in startup/status panel
  - non-TUI text logging: concise human-readable readiness line(s)
  - non-TUI JSON logging: structured startup event(s)
- Explicitly report when UI is disabled or unavailable instead of omitting it.
- Keep startup and shutdown logs routed through the same logger path to avoid fragmented output channels.

**Non-Goals:**
- Redesigning the full TUI layout or interaction model.
- Changing serving semantics, access controls, or request proxy behavior.
- Introducing new configuration flags for startup rendering in this iteration.

## Decisions

### Decision: Define a canonical startup summary model
Introduce a single in-memory startup summary struct assembled from resolved configuration and runtime outcomes.

Core fields:
- `mode` (`local-daemon` or `tsnet`)
- `exposure` (`tailnet` or `funnel`)
- `service_url`
- `web_ui_url` plus `web_ui_status` (`enabled`, `disabled`, `unavailable`)
- `capabilities` (funnel enabled, mock enabled, ui enabled, json logging enabled, tui enabled)
- `readiness` (`ready` only emitted after serving path is active)

Rationale:
- Prevents divergent startup messages between TUI, stdout, and JSON logging.
- Gives one source of truth for startup UX and tests.

Alternatives considered:
- Keep per-path bespoke logging: rejected due ongoing drift and inconsistent UX.
- Emit only human-readable text: rejected because headless automation needs structured JSON fields.

### Decision: Split rendering from readiness detection
Keep readiness detection in startup orchestration and pass the canonical summary to output adapters:
- TUI renderer writes a compact status block/panel message.
- Text logger renderer writes explicit startup-complete lines.
- JSON logger renderer emits a single structured startup event with stable keys.

Rationale:
- Decouples business state from presentation and simplifies tests.
- Ensures JSON mode remains machine-consumable while TUI remains readable.

Alternatives considered:
- Build strings inline in startup functions: rejected for poor testability and drift risk.

### Decision: Handle local-daemon and tsnet URL derivation in backend-specific adapters
Local-daemon mode and tsnet mode already derive service details in different setup paths. Keep each path responsible for computing backend-specific values, then normalize into the shared startup summary before rendering.

Rationale:
- Minimizes churn to networking setup code while standardizing output semantics.

Alternatives considered:
- Move all URL derivation into a shared helper only: rejected because mode-specific setup side effects differ and would increase coupling.

### Decision: Keep failure visibility explicit for HTTPS/Funnel prerequisites
When Funnel or HTTPS setup fails (for example certificate checks), startup must continue to fail fast, and error logs must clearly indicate the failed prerequisite. No "ready" summary is emitted in failure paths.

Rationale:
- Avoids false-positive readiness signals.
- Preserves security-sensitive behavior for public exposure.

Alternatives considered:
- Emit partial ready with degraded status: rejected because serving is not actually ready.

## Risks / Trade-offs

- [Risk] Startup summary fields drift from real runtime state over time.
  Mitigation: centralize summary assembly in one helper and cover with mode-specific tests.

- [Risk] TUI and logger adapters still diverge in wording/detail.
  Mitigation: treat startup summary as canonical and keep renderer responsibilities minimal.

- [Risk] JSON consumers depend on unstable keys.
  Mitigation: define fixed key names in spec and avoid ad-hoc field renames.

- [Risk] Funnel startup errors may be hidden by verbosity settings.
  Mitigation: keep prerequisite failures at warning/error levels independent of debug verbosity.

## Migration Plan

1. Add canonical startup summary model and helper constructors for local-daemon and tsnet startup paths.
2. Replace scattered startup-ready prints/logs with renderer calls driven by summary.
3. Update TUI startup/status display to consume canonical summary payload.
4. Add tests for:
- non-TUI text output includes readiness and URLs
- non-TUI JSON output includes structured startup event
- TUI receives startup summary with service/UI capability fields
- startup failure paths do not emit readiness
5. Update docs with examples for `tgate <port>` and `tgate <port> --funnel` showing startup output interpretation.

Rollback strategy:
- Revert startup renderer integration while keeping existing startup serving behavior unchanged.

## Open Questions

- Should we emit an additional periodic status heartbeat after startup, or keep startup summary as a one-time event only?
- For TUI, should the startup summary appear only in logs or also in a dedicated status card with persistent visibility?
