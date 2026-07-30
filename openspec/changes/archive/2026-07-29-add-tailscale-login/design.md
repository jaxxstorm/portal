## Context

Portal currently sends all incoming traffic through a proxy `Authorizer` before
dispatching to its local backend or mock handler. The only configured provider
is HTTP Basic, whose boolean interface always produces the same `401` response.
Portal is tailnet-private by default, but `portal <port> --funnel` creates an
HTTPS public endpoint; that endpoint needs a browser authentication flow for
Tailscale users outside the host tailnet.

## Goals / Non-Goals

**Goals:**
- Add `auth-mode=tailscale` using the Tailscale OAuth authorization-code flow.
- Allow any account that successfully completes that flow; do not require host
  tailnet membership or make a tailnet API lookup.
- Keep a short-lived, signed, secure browser session and block unauthenticated
  traffic before it reaches a backend.
- Work with both local Tailscale Serve and tsnet, including Funnel.

**Non-Goals:**
- Authorize by tailnet, user, group, device posture, or ACL policy.
- Support API clients, webhooks, or other non-browser login mechanisms.
- Change tailnet-private defaults, Funnel activation, HTTPS certificate setup,
  Funnel allowlist enforcement, or the existing Basic provider.
- Persist user profiles or OAuth access tokens.

## Decisions

### Use Tailscale OAuth App OpenID Connect

Portal will register internal login and callback routes, redirect the browser to
Tailscale's OAuth App authorization endpoint with the `openid` scope, exchange
the returned authorization code with the OAuth App client secret, and validate
the returned signed ID token against Tailscale's published JWKS. Portal obtains
the authorization and token endpoints from Tailscale's OIDC discovery document.
A valid token establishes authentication; Portal does not retain it because the
feature needs proof of Tailscale login, not Tailscale API access or a user
profile.

This lets an external Tailscale account authenticate without requiring it to
belong to the Portal host's tailnet. The OAuth application configuration remains
the operator's control point for which Tailscale accounts can authorize.

An OAuth flow was selected over `WhoIs` because `WhoIs` identifies tailnet
traffic only and cannot authenticate a public Funnel request. Basic credentials
remain available for non-browser clients but require sharing a secret.

### Model authorization as a handler decision

Replace the proxy's boolean-only authorization result with a provider response
that can either permit dispatch or write its own rejection/redirect response.
The Basic provider will retain its existing `401` challenge. The Tailscale
provider will redirect a browser to the login route and preserve only a
validated, same-origin return path. Login and callback routes bypass backend
dispatch but remain handled by Portal's authorization middleware, not the
proxied service.

### Protect OAuth and session state with signed cookies

Portal will generate cryptographically random OAuth state, bind it to an
HTTP-only, Secure, SameSite=Lax cookie, and compare it in constant time at the
callback. It will reject missing, expired, or mismatched state and consume it
after use. The callback will mint a short-lived HMAC-authenticated session
cookie; the signing secret is supplied through configuration rather than derived
from OAuth credentials. Cookies use the exposed mount path where practical and
are never forwarded to the backend. Logout clears the Portal session.

Signed, stateless sessions avoid a new datastore and permit local Serve and
tsnet processes to operate identically. Server-side sessions would make
revocation easier but add persistence and cleanup complexity not needed for
this initial login gate.

### Make configuration explicit and mutually validated

Add `tailscale` to `auth-mode` plus OAuth client ID, OAuth client secret, and
session signing secret settings. Each setting is available as a CLI flag,
`PORTAL_*` environment variable, and `~/.portal/config.yml`, with current
precedence. `tailscale` requires all three non-empty values; `none` and `basic`
reject Tailscale-specific values, and `basic` and `tailscale` credentials cannot
be combined. Documentation and startup errors will include the required HTTPS
Funnel callback URL and OAuth application setup.

### Preserve exposure and startup boundaries

`portal <port>` remains a tailnet-private endpoint. With `auth-mode=tailscale`,
tailnet users must also authenticate in a browser before dispatch. `portal
<port> --funnel` remains the sole public-exposure opt-in; it requires Funnel's
existing HTTPS certificate prerequisites and a publicly reachable, exact OAuth
callback URL. Funnel allowlist validation continues before authentication.

Both local Tailscale and tsnet mode run the same proxy handler, so the OAuth
routes and provider behavior are mode-independent. Existing certificate/Funnel
setup errors remain startup failures and must prevent a ready indication.

## Risks / Trade-offs

- [Incorrect callback URL or unavailable Funnel certificate] -> Validate the
  configured public base URL, document the exact callback URL, and retain clear
  HTTPS/Funnel startup errors.
- [OAuth state or session replay] -> Use CSPRNG values, short expirations,
  signed cookies, constant-time verification, one-time state consumption, and
  HTTPS-only cookie delivery.
- [Open redirect after sign-in] -> Accept only relative paths rooted at the
  configured Portal mount path and reject scheme, host, and protocol-relative
  values.
- [OAuth code interception or token substitution] -> Use a short-lived,
  one-time state value and nonce, exchange the code with the OAuth App secret,
  validate the returned ID token signature, issuer, audience, expiration, and
  nonce, then discard both tokens immediately.
- [Unauthenticated API/webhook client receives an HTML redirect] -> Scope this
  provider to browser access; operators needing machine-to-machine access use
  Basic authorization or leave endpoint authorization disabled.

## Migration Plan

1. Add configuration parsing and validation with `none` as the default, so
   existing invocations retain their behavior.
2. Add OAuth/session handlers and the richer authorizer interface, preserving
   Basic's exact challenge response.
3. Wire the provider into local Tailscale and tsnet proxy setup, then add unit
   and handler tests for callbacks, sessions, return paths, and rejected backend
   dispatch.
4. Publish configuration examples and the Funnel OAuth callback prerequisite.

Rollback is setting `auth-mode=none` or reverting the release; no persisted
state requires migration. Existing cookies become invalid after the signing
secret changes or the service is rolled back.
