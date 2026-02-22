## Context

`internal/tui/model.go` currently computes pane sizes with a mostly fixed 70/30 split and minimum widths/heights. This works for mid-size terminals but degrades on very small windows (crowded panes and clipped lines) and underuses very large displays. Endpoint discoverability is also weak because the Web UI URL appears as an unstructured line in the stats pane instead of a dedicated endpoint section.

The change needs to preserve existing startup semantics across local-daemon and tsnet modes while improving operator readability in TUI mode. In particular, tailnet remains the default private exposure mode, and Funnel remains explicit opt-in (`portal <port> --funnel`) with prerequisite-sensitive startup behavior.

## Goals / Non-Goals

**Goals:**
- Add deterministic responsive layout behavior for compact and large terminals.
- Make service URL and Web UI URL presentation high-signal and readable in TUI mode.
- Preserve startup summary semantics and mode/exposure clarity for both `portal <port>` and `portal <port> --funnel`.
- Ensure local-daemon and tsnet runs both satisfy the new layout and URL display requirements.
- Ensure Funnel/HTTPS setup failures do not present misleading endpoint-ready output.

**Non-Goals:**
- Reworking non-TUI logging formats or startup event schema keys.
- Changing core proxy behavior, request capture mechanics, or Funnel security policy.
- Exposing tsnet Web UI where it is currently unsupported.

## Decisions

1. Introduce layout profiles based on terminal size.
Rationale: fixed split logic produces unreadable panes on compact terminals and wasted space on large ones.
Alternatives considered:
- Keep one layout and only tweak constants: rejected because it does not solve both compact and large display problems.
- Add user-configurable manual layouts first: rejected for now to keep first iteration deterministic and testable.

Implementation direction:
- Define compact, standard, and wide profiles from terminal width/height thresholds.
- Compact profile uses stacked panes prioritizing endpoints and logs.
- Standard/wide profiles preserve multi-pane visibility while expanding readable content regions.

2. Add a dedicated endpoint summary section in the TUI model.
Rationale: operators need immediate visibility for service and dashboard entry points.
Alternatives considered:
- Keep endpoint lines in stats pane: rejected due to poor discoverability and URL clipping.
- Emit endpoint info only in logs: rejected because logs are noisy and not stable placement.

Implementation direction:
- Add explicit labeled fields for service URL, Web UI URL/status, mode, and exposure.
- Apply deterministic wrapping/truncation that preserves scheme and host before eliding path/query.
- Preserve readable rendering for both tailnet and Funnel exposure.

3. Track service URL and UI availability explicitly in model-facing state.
Rationale: current `StatsProvider` only exposes `GetWebUIURL`, which limits endpoint presentation and status clarity.
Alternatives considered:
- Parse URLs from free-form log lines: rejected because it is fragile and mode-dependent.
- Keep only startup summary logs as source of truth: rejected because TUI panels need stable structured data.

Implementation direction:
- Extend provider/state to expose service URL and Web UI status/reason in addition to Web UI URL.
- Populate these fields from local-daemon setup callbacks and tsnet readiness callbacks.
- In tsnet mode, surface Web UI as unavailable with an explicit reason instead of empty output.

4. Preserve startup failure semantics while improving endpoint rendering.
Rationale: improved visuals must not imply successful exposure when Funnel prerequisites fail.
Alternatives considered:
- Show last known endpoint even after failure: rejected because it can mislead operators.
- Hide endpoint panel on failure: rejected because operators still need explicit failure context.

Implementation direction:
- On HTTPS/Funnel setup failure, keep endpoint state as not-ready and render clear unavailable/failure status.
- Continue routing explicit startup failure logs into TUI logs without emitting ready-like endpoint cards.

## Risks / Trade-offs

- [Risk] Additional layout branches increase UI complexity and regression risk.
  → Mitigation: table-driven layout tests for compact/standard/wide thresholds and resize transitions.
- [Risk] URL wrapping/truncation can hide important path details for debugging.
  → Mitigation: preserve scheme+host always, with optional shortened tail and full URL available in logs.
- [Risk] State plumbing changes across startup paths could diverge between local-daemon and tsnet.
  → Mitigation: update both setup paths in the same change and add mode-specific tests.

## Migration Plan

1. Introduce structured endpoint state fields and keep existing behavior as fallback.
2. Implement responsive layout profiles and endpoint summary rendering in TUI view/model.
3. Add tests for compact/large terminal rendering and startup mode combinations (`portal <port>`, `portal <port> --funnel`).
4. Validate Funnel failure handling still suppresses ready semantics while showing clear errors.

Rollback: revert TUI layout/profile changes and endpoint state additions; existing startup logging behavior remains intact.

## Open Questions

- Should we expose a user override for layout profile selection in a follow-up (`auto|compact|standard|wide`)?
- Should endpoint summary include copy-friendly plain text mode without styling for limited terminals?
