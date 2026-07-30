## ADDED Requirements

### Requirement: Tailscale Endpoint Authorization Provider
Portal MUST support `tailscale` as an endpoint authorization provider. The
provider MUST permit requests only with a valid Portal session established by
the Tailscale login flow. For browser requests without a valid session, the
provider MUST redirect to Portal's internal login route rather than returning an
HTTP Basic challenge, and Portal MUST NOT dispatch those requests to a mock or
proxied backend.

#### Scenario: Authenticated Tailscale request is proxied
- **WHEN** `auth-mode=tailscale` is configured and a request to `portal <port>` contains a valid Portal session
- **THEN** Portal proxies the request to the local service or handles it in mock mode

#### Scenario: Unauthenticated Tailscale Funnel request is not proxied
- **WHEN** `auth-mode=tailscale` is configured and a request to `portal <port> --funnel` has no valid Portal session
- **THEN** Portal redirects the browser to its internal Tailscale login route
- **AND** Portal does not proxy or mock-handle the request

### Requirement: Provider-Specific Unauthorized Responses
Portal MUST allow an endpoint authorization provider to produce its own
unauthorized response before request dispatch. Portal MUST preserve the existing
HTTP Basic `401` response and `WWW-Authenticate: Basic realm="portal"`
challenge when `auth-mode=basic` is configured.

#### Scenario: Basic authorization behavior remains unchanged
- **WHEN** `auth-mode=basic` is configured and a request lacks valid credentials
- **THEN** Portal returns HTTP 401 with `WWW-Authenticate: Basic realm="portal"`
- **AND** Portal does not dispatch the request to the configured backend
