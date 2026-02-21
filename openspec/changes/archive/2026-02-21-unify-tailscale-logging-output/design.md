## Context

tgate runs in two Tailscale integration paths today: local daemon mode and embedded `tsnet` mode. The application already has a primary structured logger that feeds both non-interactive console output and the TUI log pane, but parts of Tailscale/tsnet initialization still write directly to alternate stdout/stderr streams. This fragments operator visibility, especially during startup and connectivity failures where logs are most important.

The change must preserve existing product behavior for private tailnet exposure and opt-in Funnel exposure while making log delivery consistent. It also must keep enough diagnostic detail for troubleshooting HTTPS certificate and Funnel setup failures.

## Goals / Non-Goals

**Goals:**
- Route all Tailscale and tsnet lifecycle output through the same primary logger channel used by tgate runtime logs.
- Keep logging behavior consistent across local daemon mode, tsnet mode, TUI mode, and non-TUI mode.
- Preserve actionable diagnostics for certificate acquisition and Funnel setup failures in the unified stream.
- Standardize structured fields so startup/setup/failure events are searchable and correlated.

**Non-Goals:**
- Changing tailnet vs Funnel exposure semantics.
- Introducing a new logging backend or external log transport.
- Redesigning request logging format outside Tailscale/tsnet lifecycle events.

## Decisions

### Decision: Add a Tailscale log adapter into the primary logger pipeline

Tailscale-facing logging hooks will be wired to an adapter that writes into the existing application logger instead of writing directly to process stdout/stderr. This gives one authoritative output channel while preserving current logging sink behavior in TUI and standard CLI execution.

Alternative considered: leaving direct Tailscale stream output enabled and documenting two channels. Rejected because it does not address fragmented UX and breaks operational expectations.

### Decision: Standardize context fields for Tailscale lifecycle events

Unified events will include consistent structured fields (for example `component`, `mode`, `phase`, and relevant node/listener identifiers). This keeps cross-mode observability coherent and avoids mixing ad hoc message formats.

Alternative considered: pass through raw upstream messages only. Rejected because it reduces filterability and correlation with existing tgate logs.

### Decision: Funnel and certificate failures must surface through the same logger channel

Failures in HTTPS certificate prerequisites, listener creation, or Funnel enablement will be emitted as structured error entries on the primary logger before startup exits or degrades. This ensures operators can diagnose failures without inspecting alternate streams.

Alternative considered: only return wrapped errors to caller without explicit lifecycle log entries. Rejected because startup failures are harder to triage without nearby structured context.

### Decision: Prevent duplicate emission across adapters and default outputs

When attaching the unified adapter, direct fallback writes from Tailscale/tsnet setup paths should be removed or gated so each lifecycle event appears once in the main channel.

Alternative considered: allow duplicate logs for safety. Rejected because duplicates reduce signal quality and degrade TUI readability.

## Risks / Trade-offs

- [Risk] Losing low-level diagnostics when mapping raw output into structured logs. -> Mitigation: preserve original message text and include mode/component metadata without truncation.
- [Risk] Duplicate logs if both adapter and legacy writers remain active in some paths. -> Mitigation: audit startup paths for direct writes and add focused tests for single-emission behavior.
- [Risk] Level-mapping mismatches (info/warn/error) from upstream messages. -> Mitigation: define deterministic mapping rules and verify with startup/failure fixtures.
- [Risk] Behavioral drift between tailnet and Funnel startup paths. -> Mitigation: require scenario coverage for both modes in specs and implementation tests.

## Migration Plan

1. Introduce the shared adapter and inject it into both local daemon and tsnet setup flows.
2. Remove or gate legacy direct log writes in startup and lifecycle paths.
3. Validate behavior in `tgate <port>`, `tgate <port> --funnel`, and `tgate --mock` with TUI and non-TUI execution.
4. Update operator documentation to state that Tailscale/tsnet lifecycle logs are available in the main application log channel.
5. Rollback strategy: revert adapter wiring and re-enable prior direct stream behavior if critical diagnostics are missing.

## Open Questions

- Should we expose a dedicated debug toggle to include additional raw Tailscale diagnostic verbosity in the unified channel?
