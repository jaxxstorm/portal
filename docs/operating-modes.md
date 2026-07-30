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
- backend: local daemon; Portal exits if tailscaled is unavailable
- backend mode: proxy
- listen mode: listener
- exposure: tailnet

## Backend Selection

Backend is selected in this order:
1. `portal serve node` or `portal mock node` uses Portal's isolated `tsnet` node.
2. The legacy `--force-tsnet` and `--auth-key` inputs also select `tsnet`.
3. Otherwise, Portal requires the local tailscaled daemon. If it is unavailable,
   Portal exits and tells the operator to start tailscaled or select `node`.

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
portal serve 8080
```

Tailnet/private listener mode through Portal's isolated node:

```bash
portal serve node 8080
```

Tailnet/private service mode on local daemon:

```bash
portal serve 8080 --listen-mode service --service-name svc:portal
```

Tailnet/private service mode through Portal's isolated node:

```bash
portal serve node 8080 --listen-mode service --service-name svc:portal
```

Mock backend on tailnet/private exposure (default):

```bash
portal mock
```

Mock backend with explicit public Funnel exposure:

```bash
portal mock --funnel
```

Mock backend through Portal's isolated node:

```bash
portal mock node
```

Public funnel mode:

```bash
portal serve 8080 --funnel
```

Invalid combination:

```bash
portal serve 8080 --listen-mode service --funnel
```

`portal 8080`, `portal --mock`, and `portal --cleanup-serve` remain supported
for existing scripts. `--force-tsnet` and `--auth-key` are also accepted for
existing scripts, but new commands should use `node`. Run `portal config init`
to create a commented default configuration file before using `portal serve`
without a port argument.

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
