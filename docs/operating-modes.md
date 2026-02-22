# Operating Modes

This page explains runtime behavior in user-facing terms.
For normative resolution rules, see [Mode Resolution Spec](mode-resolution-spec.md).

## Mental Model

portal behavior is the combination of three dimensions:
- backend: `local-daemon` or `tsnet`
- backend mode: `proxy` or `mock`
- listen mode: `listener` or `service`
- exposure: `tailnet` or `funnel`

Defaults:
- backend: local daemon when available, otherwise tsnet
- backend mode: proxy
- listen mode: listener
- exposure: tailnet

## Backend Selection

Backend is selected in this order:
1. If `--force-tsnet` is set, use `tsnet`.
2. If `--auth-key` is set, use `tsnet`.
3. Otherwise, try local tailscaled:
   - available -> `local-daemon`
   - unavailable -> `tsnet`

## Compatibility Matrix

| Backend | Listen | Exposure | Status |
|---|---|---|---|
| local-daemon | listener | tailnet | supported |
| local-daemon | listener | funnel | supported |
| local-daemon | service | tailnet | supported |
| local-daemon | service | funnel | invalid |
| tsnet | listener | tailnet | supported |
| tsnet | listener | funnel | supported |
| tsnet | service | tailnet | supported |
| tsnet | service | funnel | invalid |

Constraint:
- `service + funnel` is invalid and fails startup config validation.

## Common Runs

Tailnet/private listener mode (default):

```bash
portal 8080
```

Tailnet/private listener mode on tsnet:

```bash
portal 8080 --force-tsnet
```

Tailnet/private service mode on local daemon:

```bash
portal 8080 --listen-mode service --service-name svc:portal
```

Tailnet/private service mode on tsnet:

```bash
portal 8080 --force-tsnet --listen-mode service --service-name svc:portal
```

Mock backend on tailnet/private exposure (default):

```bash
portal --mock
```

Mock backend with explicit public Funnel exposure:

```bash
portal --mock --funnel
```

Public funnel mode:

```bash
portal 8080 --funnel
```

Invalid combination:

```bash
portal 8080 --listen-mode service --funnel
```

## Service Mode Notes

- `--listen-mode service` does not mean tsnet backend only.
- Service mode can run on either backend.
- Canonical config names are backend-agnostic (`listen-mode`, `service-name`).
- Legacy `tsnet-*` names remain as compatibility aliases.
- Local-daemon backend configures service-scoped serve behavior.
- tsnet backend uses `tsnet.Server.ListenService`.
- Service mode requires valid `svc:<dns-label>` naming.
- Service mode requires tag-based host identity; startup fails early if the node has no `tag:*` identity.
- Service advertisement still may require admin approval in your tailnet after identity validation.
- Service mode defaults `serve-port` to the target port unless explicitly overridden.

## Startup Output

Startup-ready output includes:
- `mode`: `local_daemon` or `tsnet`
- `backend_mode`: `proxy` or `mock`
- `exposure`: `tailnet` or `funnel`
- `service_url`
- `web_ui_status`

When mode is `tsnet`, startup-ready output also includes:
- `tsnet_listen_mode_configured`
- `tsnet_listen_mode_effective`
- `tsnet_service_name` (when configured)
- `tsnet_service_fqdn` (when resolved)

TSNet Web UI note:
- In current tsnet mode, the Web UI endpoint is not exposed.
- Startup output reports `web_ui_status=unavailable` with `web_ui_reason=tsnet_ui_not_exposed`.
