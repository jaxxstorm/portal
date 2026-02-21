# Configuration

tgate supports 12-factor configuration using CLI flags, environment variables,
and `~/.tgate/config.yml`.

For normative mode-combination rules, see
[Mode Resolution Spec](mode-resolution-spec.md).

## Configuration Sources

Precedence order:
1. CLI flags/args
2. Environment variables (`TGATE_*`)
3. Config file (`~/.tgate/config.yml`)
4. Built-in defaults

## Config File

Default path:

`~/.tgate/config.yml`

Example:

```yaml
port: 8080
funnel: false
funnel-allowlist: []
verbose: true
device-name: tgate
listen-mode: listener
service-name: svc:tgate
set-path: /
serve-port: 80
no-tui: false
```

## Key Mode Flags

| Purpose | CLI | Env | Default |
|---|---|---|---|
| Force tsnet backend | `--force-tsnet` | `TGATE_FORCE_TSNET` | `false` |
| Auth-key tsnet backend | `--auth-key` | `TGATE_AUTH_KEY` | empty |
| Device name | `--device-name` | `TGATE_DEVICE_NAME` | `tgate` |
| Listen mode | `--listen-mode` | `TGATE_LISTEN_MODE` | `listener` |
| Service name | `--service-name` | `TGATE_SERVICE_NAME` | `svc:tgate` |
| Public exposure | `--funnel` | `TGATE_FUNNEL` | `false` |

Hard rule:
- `--listen-mode service` cannot be combined with `--funnel`.

Naming note:
- Canonical naming is backend-agnostic: `device-name`, `listen-mode`, and `service-name`.
- Legacy aliases (`tailscale-name`, `tsnet-listen-mode`, `tsnet-service-name`) are still accepted for compatibility.

## Environment Variables

Examples:
- `TGATE_PORT=8080`
- `TGATE_FUNNEL=true`
- `TGATE_FUNNEL_ALLOWLIST=203.0.113.10,198.51.100.0/24`
- `TGATE_SERVE_PORT=443`
- `TGATE_DEVICE_NAME=my-node`
- `TGATE_LISTEN_MODE=service`
- `TGATE_SERVICE_NAME=svc:my-service`
- `TGATE_NO_TUI=true`

## CLI Examples

Tailnet/private listener mode (default):

```bash
tgate 8080
```

Tailnet/private service mode:

```bash
tgate 8080 --listen-mode service --service-name svc:tgate
```

Tailnet/private service mode on forced tsnet:

```bash
tgate 8080 --force-tsnet --listen-mode service --service-name svc:tgate
```

Public Funnel mode:

```bash
tgate 8080 --funnel
```

Invalid combination:

```bash
tgate 8080 --listen-mode service --funnel
```

## Startup Output And Web UI Location

tgate emits a definitive startup-ready log event only after serving is active.
This includes service reachability, exposure mode, and Web UI status.

When `--ui-port` is omitted, tgate prefers local port `4040` for the UI (or the
next available nearby port) and reports the effective tailnet URL in startup
output (`web_ui_url`) when UI is available.

Non-TUI JSON logging example:

```bash
tgate 8080 --no-tui --json
```

Look for keys:
- `message: "Startup ready"`
- `readiness`
- `mode`
- `exposure`
- `service_url`
- `web_ui_status`
- `web_ui_url` (when available)
- `tsnet_listen_mode_configured` / `tsnet_listen_mode_effective` (when `mode=tsnet`)

## Funnel Allowlist

Use `funnel-allowlist` in config or `TGATE_FUNNEL_ALLOWLIST` in env to restrict
Funnel requests by source IP.

Supported entry formats:
- single IP (for example `203.0.113.10`)
- CIDR block (for example `198.51.100.0/24`)
- comma-separated env list (for example `203.0.113.10,198.51.100.0/24`)

When Funnel allowlist is active and `set-path` is `/`, tgate configures Funnel
with TLS-terminated TCP forwarding + PROXY protocol v2 and uses PROXY source IP
for allowlist checks.

If `set-path` is non-root, tgate falls back to trusted HTTP metadata
(`Tailscale-Client-IP`, `Forwarded`, `X-Forwarded-For`, `X-Real-IP`) plus
socket remote address fallback.

Invalid allowlist entries fail startup with a configuration error.

For end-to-end setup details, see [IP Whitelisting](ip-whitelisting.md).
