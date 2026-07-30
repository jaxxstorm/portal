## Why

Funnel exposes Portal to users outside the host tailnet, but its current endpoint authorization options cannot authenticate a user through their Tailscale account. A Tailscale sign-in flow lets a developer share a Funnel endpoint with authenticated Tailscale users without distributing static credentials.

## What Changes

- Add a Tailscale-backed endpoint authorization provider that redirects unauthenticated browser requests to Tailscale sign-in and accepts authenticated users after a verified callback.
- Support authenticated external Tailscale users, rather than restricting sign-in to the Portal host's tailnet.
- Add configuration and validation for the Tailscale OAuth application credentials and authorization mode.
- Preserve tailnet-private behavior for `portal <port>` and require explicit `portal <port> --funnel` opt-in for public exposure; Funnel HTTPS and OAuth callback prerequisites remain required.

## Capabilities

### New Capabilities
- `tailscale-login`: Authenticate Portal endpoint users with a Tailscale account and maintain a verified browser session.

### Modified Capabilities
- `endpoint-authorization`: Add Tailscale sign-in as an endpoint authorization provider.
- `runtime-configuration`: Configure and validate Tailscale authorization settings from supported configuration sources.

## Impact

- Affects Portal's HTTP request authorization middleware, CLI/configuration surface, and tsnet/Funnel serving path.
- Adds Tailscale OAuth application credentials, callback handling, signed session state, and provider-specific integration tests.
- Funnel users must configure an OAuth application with Portal's HTTPS callback URL; Funnel remains publicly reachable, but requests are gated by Tailscale authentication when this provider is enabled.
