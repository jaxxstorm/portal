# Configuration

portal supports 12-factor configuration using CLI flags, environment variables,
and `~/.portal/config.yml`.

For normative mode-combination rules, see
[Mode Resolution Spec](mode-resolution-spec.md).

## Configuration Sources

Precedence order:
1. CLI flags/args
2. Environment variables (`PORTAL_*`)
3. Config file (`~/.portal/config.yml`)
4. Built-in defaults

## Config File

Default path:

`~/.portal/config.yml`

Example:

```yaml
port: 8080
funnel: false
funnel-allowlist: []
verbose: true
device-name: portal
listen-mode: listener
service-name: svc:portal
set-path: /
serve-port: 80
no-tui: false
```

Serve-port default behavior:
- listener mode defaults to `80` (or `443` when `use-https=true`)
- service mode defaults to the target port argument (for example `portal 8080 --listen-mode service` defaults to `serve-port=8080`), unless `--serve-port` is explicitly set

## Key Mode Flags

| Purpose | CLI | Env | Default |
|---|---|---|---|
| Force tsnet backend | `--force-tsnet` | `PORTAL_FORCE_TSNET` | `false` |
| Auth-key tsnet backend | `--auth-key` | `PORTAL_AUTH_KEY` | empty |
| Device name | `--device-name` | `PORTAL_DEVICE_NAME` | `portal` |
| Mock backend mode | `--mock` | `PORTAL_MOCK` | `false` |
| Listen mode | `--listen-mode` | `PORTAL_LISTEN_MODE` | `listener` |
| Service name | `--service-name` | `PORTAL_SERVICE_NAME` | `svc:portal` |
| Public exposure | `--funnel` | `PORTAL_FUNNEL` | `false` |

Hard rule:
- `--listen-mode service` cannot be combined with `--funnel`.
- `--listen-mode service` requires a tag-based host identity. Startup fails if the node has no `tag:*` identity.

Naming note:
- Canonical naming is backend-agnostic: `device-name`, `listen-mode`, and `service-name`.
- Legacy aliases (`tailscale-name`, `tsnet-listen-mode`, `tsnet-service-name`) are still accepted for compatibility.

## Environment Variables

Examples:
- `PORTAL_PORT=8080`
- `PORTAL_FUNNEL=true`
- `PORTAL_FUNNEL_ALLOWLIST=203.0.113.10,198.51.100.0/24`
- `PORTAL_SERVE_PORT=443`
- `PORTAL_DEVICE_NAME=my-node`
- `PORTAL_LISTEN_MODE=service`
- `PORTAL_SERVICE_NAME=svc:my-service`
- `PORTAL_NO_TUI=true`

## CLI Examples

Tailnet/private listener mode (default):

```bash
portal 8080
```

Tailnet/private service mode:

```bash
portal 8080 --listen-mode service --service-name svc:portal
```

Tailnet/private service mode on forced tsnet:

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

Public Funnel mode:

```bash
portal 8080 --funnel
```

Invalid combination:

```bash
portal 8080 --listen-mode service --funnel
```

## Startup Output And Web UI Location

portal emits a definitive startup-ready log event only after serving is active.
This includes service reachability, exposure mode, and Web UI status.

When `--ui-port` is omitted, portal prefers local port `4040` for the UI (or the
next available nearby port) and reports the effective tailnet URL in startup
output (`web_ui_url`) when UI is available.

Non-TUI JSON logging example:

```bash
portal 8080 --no-tui --json
```

Look for keys:
- `message: "Startup ready"`
- `readiness`
- `mode`
- `backend_mode`
- `exposure`
- `service_url`
- `web_ui_status`
- `web_ui_url` (when available)
- `tsnet_listen_mode_configured` / `tsnet_listen_mode_effective` (when `mode=tsnet`)

## Funnel Allowlist

Use `funnel-allowlist` in config or `PORTAL_FUNNEL_ALLOWLIST` in env to restrict
Funnel requests by source IP.

Supported entry formats:
- single IP (for example `203.0.113.10`)
- CIDR block (for example `198.51.100.0/24`)
- comma-separated env list (for example `203.0.113.10,198.51.100.0/24`)

When Funnel allowlist is active and `set-path` is `/`, portal configures Funnel
with TLS-terminated TCP forwarding + PROXY protocol v2 and uses PROXY source IP
for allowlist checks.

If `set-path` is non-root, portal falls back to trusted HTTP metadata
(`Tailscale-Client-IP`, `Forwarded`, `X-Forwarded-For`, `X-Real-IP`) plus
socket remote address fallback.

Invalid allowlist entries fail startup with a configuration error.

For end-to-end setup details, see [IP Whitelisting](ip-whitelisting.md).
