## Context

tgate currently supports private tailnet exposure by default and optional public exposure with `--funnel`, but it does not enforce request-origin restrictions for Funnel traffic. The project already uses Cobra/Viper and supports config resolution from `~/.tgate/config.yml`, `TGATE_*` environment variables, and CLI overrides. This change introduces a security-sensitive control (Funnel IP allowlist) where client source IP may be ambiguous depending on what Funnel forwards, so the design must prioritize fail-safe behavior and clear operator visibility.

## Goals / Non-Goals

**Goals:**
- Add Funnel request-origin filtering from configuration (`~/.tgate/config.yml`) and `TGATE_*` env mappings.
- Keep CLI invocation behavior unchanged for `tgate <port>` and `tgate <port> --funnel`.
- Enforce allowlist checks only for Funnel-public requests.
- Define deterministic source-IP resolution with explicit trusted-signal rules and deny-by-default on ambiguity.
- Provide logs/diagnostics that explain allow/deny decisions.

**Non-Goals:**
- Replacing IP filtering with user authentication or API-key authentication.
- Applying allowlist enforcement to tailnet-private mode by default.
- Implementing geographic blocking, rate limiting, or full WAF behavior.

## Decisions

### Decision: Add explicit Funnel allowlist configuration keys
- Add a list-based configuration key for Funnel request filtering (IP and CIDR entries).
- Add a `TGATE_*` environment mapping for the same setting using comma-separated values.
- Parse and validate entries at startup; invalid entries fail startup with a clear error.

Rationale:
- Keeps behavior consistent with existing config model.
- Failing fast avoids running a publicly exposed endpoint with a malformed policy.

Alternatives considered:
- CLI-only allowlist flag: rejected because user requested config-file-driven behavior and future sidecar/container use.
- Separate file path for allowlist: rejected due higher complexity and weaker precedence semantics.

### Decision: Apply allowlist middleware only when Funnel is enabled and allowlist is configured
- Request filtering runs in the inbound proxy path only for Funnel mode.
- Tailnet mode remains unchanged.
- Empty allowlist means no additional Funnel restriction beyond existing behavior.

Rationale:
- Preserves backward compatibility and avoids surprising tailnet behavior changes.

Alternatives considered:
- Always-on allowlist for all modes: rejected because it alters private tailnet workflows unexpectedly.

### Decision: Use trusted source-signal resolution with fail-closed fallback
- Resolve client identity from a trusted source-signal chain:
  1. trusted forwarding headers (when present and considered trustworthy for Funnel ingress),
  2. socket remote address,
  3. unresolved state.
- If no trustworthy client IP can be resolved while allowlist enforcement is enabled, deny request (403) and log reason.

Rationale:
- Funnel source-IP preservation can vary; explicit unresolved handling avoids silent bypass.

Alternatives considered:
- Fail-open on unresolved client IP: rejected because it weakens security exactly when identity is uncertain.
- RemoteAddr only: rejected because Funnel often introduces proxy hops that hide original client IP.

### Decision: Add observability for policy outcomes
- Emit structured logs for allow/deny decisions including mode, resolved source signal type, and matching allowlist entry (if any).
- Provide startup summary indicating whether Funnel allowlist enforcement is active.

Rationale:
- Necessary for debugging real-world Funnel header behavior and policy mistakes.

Alternatives considered:
- Silent enforcement only: rejected because operational debugging becomes difficult.

## Risks / Trade-offs

- [Risk] Forwarded header semantics differ from expectations in some Funnel paths.  
  Mitigation: trust-policy abstraction + fail-closed behavior + explicit logging of source signal used.

- [Risk] Overly strict policy can block legitimate traffic when source IP is unavailable.  
  Mitigation: clear documentation, startup notice, and troubleshooting guidance for required ingress/header behavior.

- [Risk] Misconfigured CIDR entries cause startup failure.  
  Mitigation: validation errors pinpoint offending entry.

## Migration Plan

1. Add new config fields and env bindings without changing existing defaults.
2. Implement allowlist parsing and request middleware behind Funnel-mode checks.
3. Add tests for config parsing, precedence, and request allow/deny behavior.
4. Document expected behavior and limitations for Funnel source-IP handling.
5. Release with clear note that unresolved source identity in Funnel + allowlist mode is denied by design.

Rollback strategy:
- Disable the feature by removing allowlist configuration entries; existing Funnel behavior remains available.

## Open Questions

- Which exact headers does Tailscale Funnel provide consistently for end-client IP across all deployment paths used by tgate users?
- Do we need an explicit operator-configurable trust-header list in v1, or is a fixed trusted set sufficient?
