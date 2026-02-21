## Context

`tgate` currently uses tsnet listener APIs (`Listen`, `ListenTLS`, and
`ListenFunnel`) for all tsnet exposure paths. Tailscale now supports
`tsnet.Server.ListenService`, which enables service-style routing with stable
service FQDNs, but this introduces new prerequisites and mode interactions:

- Service hosts must run as tagged nodes.
- Service advertisement may require admin approval.
- Service mode is not the same as Funnel public exposure.

The change needs to add `ListenService` support without weakening current
security defaults (`tailnet` by default, Funnel only when explicitly enabled)
and without regressing startup UX/logging clarity.

## Goals / Non-Goals

**Goals:**
- Add an explicit tsnet service-listening mode that uses
  `tsnet.Server.ListenService`.
- Keep existing listener-based mode as the default and preserve behavior for
  existing users.
- Preserve funnel opt-in semantics and define deterministic behavior when
  service mode is selected alongside Funnel.
- Add clear startup logging that reports selected mode and effective mode.
- Ensure runtime configuration supports service mode across CLI, env, and config
  file with validation.

**Non-Goals:**
- Replacing local-daemon (`tailscaled`) serving with Tailscale Services.
- Changing Funnel security prerequisites or allowlist semantics.
- Automating ACL tag assignment or service auto-approval in Tailscale admin.

## Decisions

### Decision: Introduce explicit listen-mode configuration

Add tsnet configuration keys for mode selection and service naming:
- `listen-mode`: `listener` (default) or `service`
- `service-name`: service identifier (for example `svc:tgate`)
- legacy aliases: `tsnet-listen-mode` and `tsnet-service-name`

Rationale:
- Keeps behavior additive and backwards-compatible.
- Gives operators explicit control over when service mode is used.

Alternatives considered:
- Auto-selecting service mode whenever tsnet is active: rejected because
  service prerequisites are stricter and would surprise existing deployments.

### Decision: Enforce strict mutual exclusivity between service mode and Funnel

When `funnel=true` and `listen-mode=service`, tgate will fail startup
configuration validation and return a clear error. No fallback behavior is
applied.

Rationale:
- Avoids hidden behavior changes and ambiguous operator intent.
- Keeps mode semantics deterministic across local-daemon and tsnet backends.
- Prevents accidental public-exposure assumptions when service mode is chosen.

Alternatives considered:
- Automatic fallback to listener mode: rejected because it hides operator intent
  and creates difficult-to-debug runtime behavior.

### Decision: Validate service configuration early, fail on invalid service identity

Startup configuration validation will reject invalid service names before
runtime setup. Runtime failures from tsnet (for example untagged host or
service advertisement failures) remain startup errors with actionable messages.

Rationale:
- Catches deterministic mistakes early.
- Keeps runtime failure causes visible when external prerequisites are missing.

Alternatives considered:
- Silent fallback for invalid service names: rejected because it hides real
  configuration errors.

### Decision: Extend startup summary with effective tsnet serving details

Startup-ready output will include fields describing configured and effective
serving mode in tsnet:
- configured mode (`listener` or `service`)
- effective mode (after validation and backend resolution)
- service name/FQDN when service mode is active

Rationale:
- Prevents operator ambiguity about what was actually activated.
- Aligns TUI/non-TUI output with existing startup-summary direction.

Alternatives considered:
- Logging mode-selection only in verbose output: rejected because this is
  operationally important even at normal log levels.

## Risks / Trade-offs

- [Risk] Service mode requires tags and may fail unexpectedly in unmanaged
  tailnets.
  Mitigation: emit explicit startup error text describing tagged-node and
  approval prerequisites.

- [Risk] Strict validation may surprise users expecting implicit fallback.
  Mitigation: return explicit config errors and document valid mode
  combinations with examples.

- [Risk] New config keys increase complexity.
  Mitigation: keep defaults stable (`listener` mode) and document examples for
  all config sources.

## Migration Plan

1. Add config keys, validation, and precedence wiring for tsnet listen mode and
   service name.
2. Refactor tsnet server setup to support both listener and service paths.
3. Add incompatibility handling for Funnel+service selection with deterministic
   configuration validation failure.
4. Extend startup-summary payload and renderers for effective mode reporting.
5. Add tests for:
   - service mode success in tailnet/private tsnet mode
   - Funnel+service conflict validation behavior
   - invalid service-name validation failure
   - startup summary fields in text and JSON logging
6. Update docs with prerequisites, mode selection, examples, and incompatibility
   semantics.

Rollback strategy:
- Keep old listener-only path behind default behavior; reverting service-mode
  branch and config keys restores current behavior.

## Open Questions

- Should a future release add an explicit override mode for service+public
  exposure, and if so, what security constraints are required?
- Do we want additional service-mode controls (`service-https`,
  `service-proxy-protocol`) in this change, or keep minimal mode+name only?
