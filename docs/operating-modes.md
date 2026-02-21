# Operating Modes

tgate supports four common operating modes. In all modes, the security baseline
is private tailnet access unless Funnel is explicitly enabled.

## Proxy Mode (Default)

Proxy mode forwards traffic to a local service running on a port on your machine.

```bash
tgate 8080
```

Behavior:
- Private by default (tailnet-only access)
- Proxies requests to `localhost:8080`
- Keeps request visibility in TUI/console logs

Startup confirmation:
- tgate emits `Startup ready` after serving is active
- Startup output includes `exposure=tailnet`, `service_url`, and Web UI status/location

## Mock Mode

Mock mode runs without a backing local service and is useful for webhooks.

```bash
tgate --mock
```

Behavior:
- Creates a mock endpoint for testing
- Auto-enables Funnel and HTTPS behavior required for public webhook testing
- Logs inbound requests for inspection

## Local Tailscale Mode

If the local Tailscale daemon is available, tgate uses it by default.

```bash
tgate 8080
```

Behavior:
- Reuses your existing Tailscale node/session
- Keeps service tailnet-private unless Funnel is enabled

## TSNet Mode

TSNet mode runs as a separate Tailscale node managed by the application.

```bash
tgate 8080 --force-tsnet
```

Or with auth key:

```bash
tgate 8080 --auth-key tskey-auth-xxxxx
```

TSNet Web UI note:
- In current tsnet mode, the Web UI endpoint is not exposed
- Startup logs report this as `web_ui_status=unavailable` with `web_ui_reason=tsnet_ui_not_exposed`

## Funnel (Public Access) Is Opt-In

Enable public internet access with:

```bash
tgate 8080 --funnel
```

Funnel notes:
- Public access is opt-in only
- HTTPS prerequisites must be met in your Tailscale admin settings
- Keep Funnel disabled when services should remain tailnet-private
- Optional `funnel-allowlist`/`TGATE_FUNNEL_ALLOWLIST` can restrict Funnel requests by source IP
- With Funnel allowlist on root path (`set-path: /`), tgate enables PROXY protocol v2 for Funnel TCP forwarding and uses the PROXY source IP for allowlist checks
- Source IP resolution depends on available trusted request metadata; if allowlist is enabled and source IP cannot be resolved, the request is denied (`403`)

Startup confirmation:
- tgate emits `Startup ready` only after Funnel serving is active
- Startup output includes `exposure=funnel` and the public `service_url`
- If Funnel prerequisites fail (for example HTTPS certificates), startup fails and no `Startup ready` event is emitted

## Startup Output Examples

Tailnet/private startup:

```bash
tgate 8080 --no-tui
```

Expected startup-ready fields:
- `mode=local_daemon` or `mode=tsnet`
- `exposure=tailnet`
- `service_url=...`
- `web_ui_status=enabled|disabled|unavailable`
- If `mode=tsnet`, expect `web_ui_reason=tsnet_ui_not_exposed`

Funnel/public startup:

```bash
tgate 8080 --funnel --no-tui
```

Expected startup-ready fields:
- `exposure=funnel`
- `service_url=...` (public URL)
- capability flags such as `capability_funnel=true`
- If `mode=tsnet`, expect `web_ui_status=unavailable` and `web_ui_reason=tsnet_ui_not_exposed`
