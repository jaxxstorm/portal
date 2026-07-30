## Purpose

Define browser-based Tailscale OAuth App authentication for Portal endpoints,
including support for authenticated users outside the Portal host's tailnet.

## Requirements

### Requirement: Tailscale OAuth Browser Login
When Tailscale endpoint authorization is enabled, Portal MUST authenticate a
browser through Tailscale OAuth App OpenID Connect before permitting the request
to reach a proxy or mock backend. Portal MUST provide login and callback routes
that are handled internally and MUST NOT proxy either route.

#### Scenario: Unauthenticated tailnet browser starts sign-in
- **WHEN** a browser requests `portal <port>` with Tailscale authorization enabled and no valid Portal session
- **THEN** Portal redirects the browser to the Tailscale authorization endpoint
- **AND** Portal does not dispatch the request to the configured backend

#### Scenario: Unauthenticated Funnel browser starts sign-in
- **WHEN** a browser requests `portal <port> --funnel` with Tailscale authorization enabled, Funnel HTTPS prerequisites satisfied, and no valid Portal session
- **THEN** Portal redirects the browser to the Tailscale authorization endpoint
- **AND** Portal does not dispatch the request to the configured backend

#### Scenario: Successful callback creates a session
- **WHEN** Tailscale redirects to Portal's callback route with an authorization code and matching unexpired state
- **THEN** Portal exchanges the code using the OAuth App credentials, verifies the returned ID token's signature, issuer, audience, expiration, and nonce, creates a secure authenticated Portal session, and redirects to the validated original Portal path

### Requirement: External Tailscale Accounts Are Accepted
Portal MUST treat a successful Tailscale OAuth callback as authentication without
requiring the account to be a member of the Portal host's tailnet. Portal MUST
NOT use tailnet membership, device identity, or a `WhoIs` result as a condition
for this provider to authorize a session.

#### Scenario: External Tailscale user accesses Funnel endpoint
- **WHEN** a Tailscale account outside the Portal host's tailnet successfully completes the configured OAuth flow for `portal <port> --funnel`
- **THEN** Portal creates an authenticated session and permits that user's subsequent valid requests

### Requirement: OAuth And Session State Is Protected
Portal MUST generate OAuth state with cryptographically secure randomness, bind
it to an HTTP-only, Secure, SameSite cookie with a bounded lifetime, and reject
missing, expired, reused, or mismatched callback state. Portal MUST issue an
HTTP-only, Secure, signed session cookie with a bounded lifetime, MUST validate
its signature and expiration before authorization, and MUST clear the session on
logout.

#### Scenario: Invalid callback state is rejected
- **WHEN** the OAuth callback has no state, a state different from the stored state, or an expired state
- **THEN** Portal rejects the callback without creating a session
- **AND** Portal does not dispatch the callback to the configured backend

#### Scenario: Tampered session is rejected
- **WHEN** a request presents a Portal session cookie with an invalid signature or expired timestamp
- **THEN** Portal treats the request as unauthenticated and does not dispatch it to the configured backend

### Requirement: Post-Login Redirect Is Same-Origin
Portal MUST preserve a requested return path only when it is a relative path
within Portal's configured mount path. Portal MUST reject absolute URLs,
protocol-relative URLs, and paths outside that mount path as post-login return
targets.

#### Scenario: External return URL is rejected
- **WHEN** an unauthenticated request supplies `https://attacker.example` as a return target
- **THEN** Portal redirects after successful sign-in to the configured Portal root rather than the supplied URL
