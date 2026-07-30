## 1. Configuration

- [x] 1.1 Add the `tailscale` authorization mode and OAuth client ID, OAuth client secret, and session signing secret to Portal configuration, flags, environment bindings, and config-file support.
- [x] 1.2 Validate provider-specific credentials and incompatible authorization settings while preserving `none` and HTTP Basic defaults.
- [x] 1.3 Add configuration tests for source precedence, missing Tailscale credentials, invalid mode combinations, and unchanged Basic validation.

## 2. Tailscale Login

- [x] 2.1 Implement Tailscale OAuth App OpenID Connect setup and internal login, callback, and logout handlers.
- [x] 2.2 Implement cryptographically random, expiring OAuth state and signed, expiring Portal session cookies with secure cookie attributes and constant-time validation.
- [x] 2.3 Validate post-login return paths and reject external, protocol-relative, or out-of-mount redirects.
- [x] 2.4 Add handler tests for successful callback/session creation, invalid or replayed state, tampered or expired sessions, logout, and return-path validation.

## 3. Endpoint Authorization

- [x] 3.1 Evolve the authorization provider boundary so a provider can write its own unauthorized redirect or response before backend dispatch.
- [x] 3.2 Add the Tailscale session authorization provider and route Portal-owned OAuth endpoints around proxy/mock dispatch.
- [x] 3.3 Preserve HTTP Basic's `401` and `WWW-Authenticate` behavior and remove Portal authorization credentials before proxying authorized requests.
- [x] 3.4 Add proxy tests proving unauthenticated Tailscale requests never reach either backend mode and authenticated sessions are dispatched.

## 4. Exposure And Documentation

- [x] 4.1 Wire Tailscale authorization into both local Tailscale Serve and tsnet startup paths without changing tailnet-private or explicit Funnel exposure semantics.
- [x] 4.2 Surface the exact OAuth callback URL and Funnel HTTPS prerequisite in startup errors or operator guidance without reporting readiness after setup failure.
- [x] 4.3 Document `portal <port>` and `portal <port> --funnel` examples, required OAuth application configuration, external-account behavior, and the browser-only limitation.
- [x] 4.4 Run the Go test suite and add coverage for tailnet-only and Funnel-public login paths, including existing certificate/Funnel startup failure behavior.
