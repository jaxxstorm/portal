## Context

Portal already checks local-daemon HTTPS capability and certificate availability
before enabling Funnel, and emits lifecycle messages during that work. Its final
startup summary reports `exposure=funnel` and the service URL, but does not
provide a stable, concise status for certificate verification or Funnel
activation. tsnet performs equivalent serving setup through its own lifecycle,
so status reporting must use a common startup-summary contract without claiming
certificate details that a mode cannot observe directly.

## Goals / Non-Goals

**Goals:**
- Add stable certificate and Funnel readiness fields to startup output in TUI,
  text logs, and JSON logs.
- Distinguish tailnet-private startup from public Funnel startup, and identify
  the public URL only after serving is ready.
- Make certificate and Funnel prerequisite failures actionable in the primary
  logger for both local-daemon and tsnet modes.

**Non-Goals:**
- Change Funnel enablement, HTTPS certificate issuance, ACLs, or public-exposure
  defaults.
- Add an external internet probe or continuously monitor certificate renewal.
- Guarantee that a URL is reachable from every network after Tailscale reports
  the Funnel listener ready.

## Decisions

### Extend the startup summary with explicit status fields

The startup summary will carry `certificate_status` and `funnel_status` in
addition to its existing exposure and URL fields. Funnel-ready runs report a
verified/ready status and public URL. Tailnet-only runs report that Funnel is
disabled and certificate readiness is not required, preventing users of
`portal <port>` from mistaking HTTPS capability logs for public exposure.

This extends the single authoritative readiness event rather than relying on
operators to infer state from a sequence of lifecycle messages. A separate
status endpoint was rejected because it would duplicate state and offer less
value during startup failures.

### Report only verified setup milestones

For local-daemon Funnel setup, report certificate verified only after the
existing certificate-domain check succeeds and Funnel ready only after the serve
configuration is successfully applied. For tsnet, report Funnel ready only once
the tsnet serving path reports readiness; certificate status reflects the
successful HTTPS/Funnel setup rather than an unverified external reachability
claim.

Failures emit the failed phase and the existing underlying error through the
primary logger, and no ready summary is emitted. This preserves the current
security boundary: a configured URL is not described as public until Funnel is
active.

### Keep output mode semantics aligned

Populate the fields in the shared startup summary model so JSON logs retain
stable keys, console logs retain readable labels, and the TUI can render the
same state. Existing service URL and capability fields remain unchanged for
automation compatibility.

## Risks / Trade-offs

- [A ready status is mistaken for universal internet health] -> Name the field
  setup/readiness status and document that it confirms Tailscale configuration,
  not an external synthetic probe.
- [Local daemon and tsnet expose different certificate internals] -> Normalize
  the common verified/ready outcome and retain mode-specific failure details in
  lifecycle logs.
- [Additional fields disrupt log consumers] -> Add fields without renaming or
  removing existing startup summary keys.

## Migration Plan

1. Add status values to the shared startup-summary model and output renderers.
2. Thread confirmed local-daemon and tsnet Funnel setup milestones into that
   model, retaining existing failure paths.
3. Add unit and startup-path tests for tailnet, Funnel-ready, certificate
   failure, and Funnel failure cases.
4. Document the meaning and limits of the reported status values.

Rollback consists of reverting the additive status fields; serving and exposure
behavior remains unchanged.

## Open Questions

None.
