## Why

Funnel startup currently emits several low-level lifecycle messages, but an
operator cannot quickly tell whether an HTTPS certificate was confirmed and the
public Funnel endpoint is actually ready. Clear, final status reporting reduces
uncertainty while keeping public exposure an explicit, security-sensitive choice.

## What Changes

- Report certificate readiness and Funnel activation in a concise, structured
  status summary when `portal <port> --funnel` succeeds.
- Report the failed prerequisite and its remediation context when certificate
  issuance or Funnel activation cannot complete.
- Preserve tailnet-private `portal <port>` behavior and make it explicit that no
  Funnel certificate or public endpoint is being configured in that mode.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `startup-operator-feedback`: Include certificate and Funnel readiness in
  startup reporting and failure feedback.
- `local-service-exposure`: Clarify confirmed public Funnel availability and
  certificate-prerequisite failures.

## Impact

- Affects local Tailscale and tsnet startup paths, Tailscale certificate and
  Funnel lifecycle logging, startup summaries, TUI output, and related tests.
- Does not change the default tailnet-private posture or make Funnel public
  exposure automatic; `portal <port> --funnel` still requires HTTPS and
  Tailscale Funnel prerequisites.
