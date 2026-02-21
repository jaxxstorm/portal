# Configuration

tgate supports 12-factor configuration using CLI arguments, environment
variables, and a config file.

## Configuration Sources

tgate resolves configuration in this precedence order:

1. CLI arguments and flags
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
tailscale-name: tgate
set-path: /
serve-port: 80
no-tui: false
```

## Environment Variables

All runtime settings can be configured with the `TGATE_` prefix.

Examples:
- `TGATE_PORT=8080`
- `TGATE_FUNNEL=true`
- `TGATE_FUNNEL_ALLOWLIST=203.0.113.10,198.51.100.0/24`
- `TGATE_SERVE_PORT=443`
- `TGATE_TAILSCALE_NAME=my-node`
- `TGATE_NO_TUI=true`

## CLI Examples

Tailnet-only default:

```bash
tgate 8080
```

Public Funnel:

```bash
tgate 8080 --funnel
```

Public Funnel with allowlist:

```bash
TGATE_FUNNEL=true TGATE_FUNNEL_ALLOWLIST=203.0.113.10,198.51.100.0/24 tgate 8080
```

Mock mode:

```bash
tgate --mock
```

## Startup Output And Web UI Location

tgate emits a definitive startup-ready log entry once traffic serving is active.
This includes service reachability, exposure mode, and Web UI status.

When `--ui-port` is omitted, tgate auto-assigns a local UI port and reports the
effective tailnet URL in the same startup-ready output (`web_ui_url`).

Non-TUI text logging example:

```bash
tgate 8080 --no-tui
```

Look for:
- `Startup ready`
- `service_url=...`
- `web_ui_status=enabled|disabled|unavailable`
- `web_ui_url=...` (when UI is available)

Non-TUI JSON logging example:

```bash
tgate 8080 --no-tui --json
```

Look for a structured event with keys:
- `message: "Startup ready"`
- `readiness`
- `mode`
- `exposure`
- `service_url`
- `web_ui_status`
- `web_ui_url` (when available)

## Precedence Example

If you have:
- `~/.tgate/config.yml` with `serve-port: 80`
- `TGATE_SERVE_PORT=443`
- CLI flag `--serve-port 8443`

Then tgate uses `8443` (CLI wins).

If you also configure Funnel allowlist in multiple places:
- `~/.tgate/config.yml` with `funnel-allowlist: [203.0.113.10]`
- `TGATE_FUNNEL_ALLOWLIST=198.51.100.0/24`

Then tgate uses the environment value (`198.51.100.0/24`).

## Funnel Allowlist

Use `funnel-allowlist` in config or `TGATE_FUNNEL_ALLOWLIST` in env to restrict
Funnel requests by source IP.

Supported entry formats:
- Single IP (for example `203.0.113.10`)
- CIDR block (for example `198.51.100.0/24`)
- Comma-separated values in env (for example `203.0.113.10,198.51.100.0/24`)

When Funnel allowlist is active and `set-path` is `/`, tgate configures Funnel
with TLS-terminated TCP forwarding + PROXY protocol v2 and uses the connection
source IP from PROXY metadata for allowlist checks.

If `set-path` is non-root, tgate falls back to trusted HTTP metadata
(`Tailscale-Client-IP`, `Forwarded`, `X-Forwarded-For`, `X-Real-IP`) plus
socket remote address fallback.

Invalid allowlist entries fail startup with a configuration error.

For full setup and behavior details, see [IP Whitelisting](ip-whitelisting.md).
