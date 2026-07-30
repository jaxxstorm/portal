## Purpose

Define opt-in authorization for Portal's exposed endpoints, initially through
HTTP Basic credentials and with a provider-neutral extension point for future
authorization providers.

## Requirements

### Requirement: Opt-In Endpoint Authorization
Portal MUST leave endpoint authorization disabled unless an authorization mode
is explicitly configured. When enabled, Portal MUST authorize requests before
they are dispatched to a mock backend or proxied local service.

#### Scenario: Default tailnet endpoint remains authorization-free
- **WHEN** a developer runs `portal <port>` without authorization configuration
- **THEN** Portal serves tailnet-private requests without an authorization challenge

#### Scenario: Default Funnel endpoint remains authorization-free
- **WHEN** a developer runs `portal <port> --funnel` without authorization configuration and Funnel HTTPS prerequisites are satisfied
- **THEN** Portal serves the explicitly public endpoint without an authorization challenge

### Requirement: HTTP Basic Authorization Provider
Portal MUST support an HTTP Basic authorization provider using a configured
username and password. Portal MUST return HTTP 401 with a Basic authentication
challenge when credentials are missing or invalid and MUST NOT dispatch that
request to the configured backend.

#### Scenario: Authorized tailnet request is proxied
- **WHEN** `auth-mode=basic` is configured and a request to `portal <port>` supplies the configured HTTP Basic username and password
- **THEN** Portal proxies the request to the local service or handles it in mock mode

#### Scenario: Unauthorized Funnel request is rejected
- **WHEN** `auth-mode=basic` is configured and a request to `portal <port> --funnel` reaches Portal without valid HTTP Basic credentials
- **THEN** Portal returns HTTP 401 with `WWW-Authenticate: Basic realm="portal"`
- **AND** Portal does not proxy or mock-handle the request

### Requirement: Extensible Authorization Provider Boundary
Portal MUST route request authorization through a provider-neutral interface so
additional authorization providers can be introduced without changing proxy or
mock request dispatch behavior.

#### Scenario: Disabled authorization uses the same dispatch path
- **WHEN** no authorization provider is configured
- **THEN** Portal dispatches permitted requests through its existing mock or reverse-proxy behavior

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
