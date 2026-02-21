## Why

tgate currently exposes Funnel endpoints to the public internet without any built-in request origin filtering. Teams need a configurable allowlist to reduce unintended access, but Funnel may not always preserve the original client IP in the same way as direct ingress, so the behavior must be explicit and safe.

## What Changes

- Add configurable IP allowlist support for Funnel mode, sourced from `~/.tgate/config.yml` and `TGATE_*` environment variables.
- Enforce request filtering before proxying traffic to local services when Funnel is enabled.
- Define trusted client IP extraction behavior (socket remote address and trusted forwarding headers) with a fail-safe deny outcome when no trustworthy client IP can be determined.
- Add runtime diagnostics/logging to explain why requests are allowed or denied.
- Keep tailnet-private mode behavior unchanged for `tgate <port>`.
- Document behavior and limitations for `tgate <port> --funnel`, including prerequisite and security implications.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `local-service-exposure`: Add Funnel allowlist enforcement semantics and handling when source IP cannot be reliably determined.
- `runtime-configuration`: Add configuration keys and `TGATE_*` mappings for Funnel allowlist behavior.

## Impact

- Affected code: HTTP request handling and middleware in Funnel mode, configuration loading/validation, and startup configuration reporting.
- Affected docs: Funnel operating mode and configuration reference.
- External dependencies: None required beyond existing net/IP parsing support.
- Risk areas: Misinterpreting forwarded headers could allow bypasses; design must define trusted header policy and deny-by-default behavior for ambiguous client identity.
