# Mode Resolution Spec

This page is the canonical behavior spec for how portal resolves:
- backend mode (`local-daemon` vs `tsnet`)
- listen mode (`listener` vs `service`)
- exposure mode (`tailnet` vs `funnel`)

## Inputs

Primary configuration inputs:
- `--force-tsnet` / `PORTAL_FORCE_TSNET`
- `--auth-key` / `PORTAL_AUTH_KEY`
- `--device-name` / `PORTAL_DEVICE_NAME`
- `--listen-mode` / `PORTAL_LISTEN_MODE`
- `--service-name` / `PORTAL_SERVICE_NAME`
- `--funnel` / `PORTAL_FUNNEL`

Naming note:
- Canonical names are `device-name`, `listen-mode`, and `service-name`.
- Legacy aliases `tailscale-name`, `tsnet-listen-mode`, and `tsnet-service-name` are accepted.
- If canonical and legacy names are both set with different values, startup fails with a configuration conflict error.

Supporting inputs:
- `--serve-port` / `PORTAL_SERVE_PORT`
- `--use-https` / `PORTAL_USE_HTTPS`
- `--set-path` / `PORTAL_SET_PATH`

## Resolution Order

1. Parse config with standard precedence: CLI > env > config file > defaults.
2. Validate static constraints (mode conflicts and value formats).
3. Resolve backend:
   - If `force-tsnet=true` or `auth-key` is set: backend is `tsnet`.
   - Otherwise, if local tailscaled is available: backend is `local-daemon`.
   - Otherwise: backend is `tsnet`.
4. Resolve listen mode:
   - `listener` (default) or `service`.
5. Resolve exposure mode:
   - `tailnet` (default) or `funnel`.

## Compatibility Matrix

| Backend | Listen | Exposure | Result |
|---|---|---|---|
| local-daemon | listener | tailnet | supported |
| local-daemon | listener | funnel | supported |
| local-daemon | service | tailnet | supported |
| local-daemon | service | funnel | invalid |
| tsnet | listener | tailnet | supported |
| tsnet | listener | funnel | supported |
| tsnet | service | tailnet | supported |
| tsnet | service | funnel | invalid |

## Hard Constraints

- `listen-mode=service` MUST NOT be combined with `funnel=true`.
- `service-name` MUST be a valid `svc:<dns-label>` value when service mode is selected.
- Invalid combinations fail startup configuration validation (no silent fallback).

## Backend-Specific Behavior

Listener mode:
- `local-daemon`: uses Tailscale Serve node listener behavior.
- `tsnet`: uses `Listen`/`ListenTLS`/`ListenFunnel` based on TLS/funnel settings.

Service mode:
- `local-daemon`: configures service-scoped serve handlers and advertises the service.
- `tsnet`: uses `tsnet.Server.ListenService`.

## Startup Observability

When backend is `tsnet`, startup-ready logs include:
- `tsnet_listen_mode_configured`
- `tsnet_listen_mode_effective`
- `tsnet_service_name` (when set)
- `tsnet_service_fqdn` (when resolved)

For mode conflicts (for example `service + funnel`), startup fails early with a configuration error and no ready event is emitted.

## Examples

Tailnet/private listener mode (default):

```bash
portal 8080
```

Tailnet/private service mode on local daemon:

```bash
portal 8080 --listen-mode service --service-name svc:portal
```

Tailnet/private service mode on tsnet:

```bash
portal 8080 --force-tsnet --listen-mode service --service-name svc:portal
```

Invalid mode combination:

```bash
portal 8080 --listen-mode service --funnel
```
