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

## Commands

Use focused commands for common workflows:

```bash
# Tailnet-private local service (default)
portal serve 8080

# Tailnet-private local service through Portal's isolated Tailscale node
portal serve node 8080

# Explicit public Funnel exposure; HTTPS and Funnel prerequisites apply
portal serve 8080 --funnel

# Tailnet-private mock endpoint
portal mock

# Remove Tailscale Serve configuration
portal cleanup
```

`portal 8080`, `portal --mock`, and `portal --cleanup-serve` remain compatible
with the command-oriented forms above.

`portal --help` lists commands rather than serving flags. Only `--version` is a
global flag.

`portal serve --help` and `portal mock --help` show the common runtime flags.
`portal serve node --help` and `portal mock node --help` additionally show the
isolated node's `--device-name` and `--auth-key` settings. Advanced UI, path,
port, and credential settings remain accepted for compatibility but are
intentionally kept in `~/.portal/config.yml` or `PORTAL_*` environment variables
to keep everyday command help short.

## Configuration Template

Generate an editable, tailnet-private starter file with:

```bash
portal config init
```

Portal creates `~/.portal/config.yml` with user-only file permissions and never
includes credentials in the template. Existing files are protected; use
`portal config init --force` only when intentionally replacing it. Keep OAuth,
authorization, and other secret values in a secret manager or `PORTAL_*`
environment variables instead of the configuration file.

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
auth-mode: none
```

Serve-port default behavior:
- listener mode defaults to `80` (or `443` when `use-https=true`)
- service mode defaults to the target port argument (for example `portal 8080 --listen-mode service` defaults to `serve-port=8080`), unless `--serve-port` is explicitly set

## Key Mode Flags

| Purpose | CLI | Env | Default |
|---|---|---|---|
| Isolated Tailscale node | `portal serve node` / `portal mock node` | n/a | local daemon |
| Legacy forced tsnet backend | `--force-tsnet` | `PORTAL_FORCE_TSNET` | `false` |
| Isolated-node auth key | `--auth-key` | `PORTAL_AUTH_KEY` | empty |
| Device name | `--device-name` | `PORTAL_DEVICE_NAME` | `portal` |
| Mock backend mode | `--mock` | `PORTAL_MOCK` | `false` |
| Listen mode | `--listen-mode` | `PORTAL_LISTEN_MODE` | `listener` |
| Service name | `--service-name` | `PORTAL_SERVICE_NAME` | `svc:portal` |
| Public exposure | `--funnel` | `PORTAL_FUNNEL` | `false` |
| Endpoint authorization mode | `--auth-mode` | `PORTAL_AUTH_MODE` | `none` |
| HTTP Basic username | `--auth-username` | `PORTAL_AUTH_USERNAME` | empty |
| HTTP Basic password | `--auth-password` | `PORTAL_AUTH_PASSWORD` | empty |
| Tailscale OAuth App client ID | `--tailscale-oauth-client-id` | `PORTAL_TAILSCALE_OAUTH_CLIENT_ID` | empty |
| Tailscale OAuth App client secret | `--tailscale-oauth-client-secret` | `PORTAL_TAILSCALE_OAUTH_CLIENT_SECRET` | empty |
| Tailscale login session secret | `--tailscale-session-secret` | `PORTAL_TAILSCALE_SESSION_SECRET` | empty |

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
- `PORTAL_AUTH_MODE=basic`
- `PORTAL_AUTH_USERNAME=alice`
- `PORTAL_AUTH_PASSWORD=change-me`
- `PORTAL_AUTH_MODE=tailscale`
- `PORTAL_TAILSCALE_OAUTH_CLIENT_ID=...`
- `PORTAL_TAILSCALE_OAUTH_CLIENT_SECRET=...`
- `PORTAL_TAILSCALE_SESSION_SECRET=...`

## CLI Examples

Tailnet/private listener mode (default):

```bash
portal 8080
```

Tailnet/private service mode:

```bash
portal 8080 --listen-mode service --service-name svc:portal
```

Tailnet/private service mode through Portal's isolated node:

```bash
portal serve node 8080 --listen-mode service --service-name svc:portal
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

Tailnet/private endpoint with HTTP Basic authorization:

```bash
portal 8080 --auth-mode basic --auth-username alice --auth-password change-me
```

Public Funnel endpoint with HTTP Basic authorization:

```bash
portal 8080 --funnel --auth-mode basic --auth-username alice --auth-password change-me
```

Public Funnel endpoint with Tailscale login:

```bash
portal 8080 --funnel --auth-mode tailscale \
  --tailscale-oauth-client-id "$PORTAL_TAILSCALE_OAUTH_CLIENT_ID" \
  --tailscale-oauth-client-secret "$PORTAL_TAILSCALE_OAUTH_CLIENT_SECRET" \
  --tailscale-session-secret "$PORTAL_TAILSCALE_SESSION_SECRET"
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
- `certificate_status`
- `funnel_status`
- `web_ui_status`
- `web_ui_url` (when available)
- `tsnet_listen_mode_configured` / `tsnet_listen_mode_effective` (when `mode=tsnet`)

## Funnel Readiness

`portal <port>` remains tailnet-private. Its startup-ready summary reports:

```text
exposure=tailnet certificate_status=not_required funnel_status=disabled
```

`portal <port> --funnel` explicitly enables public exposure. Once Tailscale has
confirmed HTTPS certificate availability and activated Funnel, its startup-ready
summary reports the public `service_url` with:

```text
exposure=funnel certificate_status=verified funnel_status=ready
```

`certificate_status=verified` and `funnel_status=ready` confirm that Portal's
Tailscale serving path is configured and active. They are not an external
synthetic reachability probe, so firewall, DNS, or client-network failures can
still prevent a particular internet client from reaching the endpoint.

If certificate validation or Funnel activation fails, Portal emits a structured
failure log with the failed status and does not emit `Startup ready`. Verify that
HTTPS certificates and Funnel are enabled for the tailnet, then retry.

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

## Endpoint Authorization

Endpoint authorization is disabled by default. `auth-mode=basic` requires
non-empty `auth-username` and `auth-password` values before Portal proxies a
request or produces a mock response. Unauthenticated requests receive `401
Unauthorized` with a Basic challenge.

Authorization applies to both tailnet-private endpoints and explicitly public
Funnel endpoints. It does not make a tailnet endpoint public, and it does not
enable Funnel. In Funnel mode, Portal performs the Funnel allowlist check before
the authorization check; HTTPS remains a Funnel prerequisite and protects Basic
credentials in transit.

Do not commit Basic passwords to source control. Prefer `PORTAL_AUTH_PASSWORD`
or a configuration file with access restricted to the Portal operator. Portal
redacts request authorization headers from its request log and does not forward
them to the configured backend.

### Tailscale Login

`auth-mode=tailscale` protects a Portal endpoint with a Tailscale OAuth App. An
unauthenticated browser is redirected to Tailscale, then Portal exchanges the
authorization code and creates a signed session after verifying the returned ID
token.

This flow is browser-only. The OAuth App and each person who signs in must belong
to the same tailnet. Use HTTP Basic authorization for webhooks and other
non-browser clients.

#### Before You Start

You need:

1. The exact hostname Portal will use, such as
   `portal.example-tailnet.ts.net`. It is normally the Tailscale device name
   followed by the tailnet DNS name; Portal's startup-ready `service_url`
   confirms the value.
2. A Tailscale management OAuth client owned by a tailnet Owner or Admin, with
   permission to create OAuth Apps.
3. A service that can run with HTTPS. Funnel is optional, but is required when
   the endpoint must be reachable from the public internet.

The hostname cannot change after you create the OAuth App. Portal always
registers this callback URL and does not accept a custom callback path:

```text
https://<hostname>/_portal/auth/tailscale/callback
```

#### 1. Create The OAuth App

Pass only the hostname, without `https://` or a path:

```bash
portal oauth-app create \
  --oauth-client-id "$PORTAL_MANAGEMENT_OAUTH_CLIENT_ID" \
  --oauth-client-secret "$PORTAL_MANAGEMENT_OAUTH_CLIENT_SECRET" \
  --hostname "portal.example-tailnet.ts.net"
```

The command exchanges the management credentials for a short-lived API token,
creates the OAuth App, and prints its client ID, client secret, session secret,
and the Portal authorization flags. The generated OAuth App secret is shown only
once. Store it and the session secret in a secret manager immediately; Portal
does not write either value to disk.

#### 2. Configure Portal

Set the printed values as environment variables. Do not commit them to a config
file or source control:

```bash
export PORTAL_AUTH_MODE=tailscale
export PORTAL_TAILSCALE_OAUTH_CLIENT_ID="<OAuth App client ID>"
export PORTAL_TAILSCALE_OAUTH_CLIENT_SECRET="<OAuth App client secret>"
export PORTAL_TAILSCALE_SESSION_SECRET="<Portal session secret>"
```

Start the endpoint that uses the hostname registered above. This example makes
the service public through Funnel:

```bash
portal serve 8080 --funnel
```

For a tailnet-only endpoint, omit `--funnel`. For Portal's isolated Tailscale
node, use `portal serve node 8080 --funnel` and use the hostname formed from
that node's `--device-name` when creating the OAuth App.

`auth-mode=tailscale` automatically enables HTTPS for secure session cookies.
Funnel remains an explicit choice. When Funnel is enabled, Portal applies any
configured Funnel allowlist before redirecting a browser to Tailscale login.

#### 3. Verify Browser Login

1. Open Portal's `service_url` in a browser.
2. Portal redirects you to Tailscale login when no valid Portal session exists.
3. Sign in with an account in the OAuth App's tailnet and approve the request.
4. Tailscale redirects the browser to Portal's fixed callback URL.
5. Portal verifies the ID token, sets a secure session cookie, and serves the
   requested endpoint.

If Tailscale reports a redirect URI mismatch, recreate the OAuth App with the
exact hostname in Portal's `service_url`. Do not append the callback path to
`--hostname`; Portal adds it automatically. If the login callback cannot be
reached, verify HTTPS certificate and Funnel readiness before retrying.
