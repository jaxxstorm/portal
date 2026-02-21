# IP Whitelisting

Use Funnel IP whitelisting to restrict public (`--funnel`) traffic by client IP.

This control applies only to Funnel mode. Tailnet-only mode is unchanged.

## Quick Start

Config file (`~/.tgate/config.yml`):

```yaml
port: 8080
funnel: true
funnel-allowlist:
  - 203.0.113.10
  - 198.51.100.0/24
```

Environment variables:

```bash
TGATE_PORT=8080
TGATE_FUNNEL=true
TGATE_FUNNEL_ALLOWLIST=203.0.113.10,198.51.100.0/24
```

Entries must be valid IPs or CIDRs. Invalid entries fail startup.

## Precedence

Standard tgate precedence applies:

1. CLI
2. Environment
3. Config file
4. Defaults

For allowlist, `TGATE_FUNNEL_ALLOWLIST` overrides `funnel-allowlist` in config.

## Source IP Resolution

When allowlist is active, tgate resolves source IP differently based on runtime mode:

- Funnel + allowlist + `set-path: /` + local Tailscale daemon:
  tgate configures TLS-terminated TCP forwarding with PROXY protocol v2 and uses
  connection source IP (`RemoteAddr`) for allowlist checks.
- Any other case (for example non-root `set-path`, tsnet mode, or fallback):
  tgate uses trusted HTTP metadata in this order:
  `Tailscale-Client-IP` -> `Forwarded` -> `X-Forwarded-For` ->
  `X-Real-IP` -> socket `RemoteAddr`.

## Enforcement Behavior

When allowlist is configured:

- Resolved source IP matches allowlist -> request is proxied.
- Resolved source IP does not match -> HTTP `403`.
- Source IP cannot be resolved -> HTTP `403` (fail closed).

## Operational Notes

- If allowlist is enabled with non-root `set-path` (for example `/api`), tgate
  cannot use the Funnel TCP+PROXY path and falls back to HTTP metadata.
- In local-daemon mode, if PROXY protocol is expected but not present, requests
  are denied in allowlist mode.
- Structured logs include allow/deny outcome, source signal, and deny reason.

## See Also

- [Configuration](configuration.md)
- [Operating Modes](operating-modes.md)
- [Troubleshooting](troubleshooting.md)
