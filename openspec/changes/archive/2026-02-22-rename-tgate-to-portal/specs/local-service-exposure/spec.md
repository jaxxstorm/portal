## MODIFIED Requirements

### Requirement: Tailnet Proxy For Local Developer Services
portal MUST proxy a developer's local HTTP service to their Tailscale tailnet
so authenticated tailnet users can access it securely, regardless of whether
the service configuration is provided by CLI, `~/.portal/config.yml`, or
`PORTAL_*` environment variables. Funnel request allowlist controls MUST NOT
affect tailnet-private request handling.

#### Scenario: Tailnet proxy is available from CLI configuration
- **WHEN** a developer runs `portal <port>` and the local service is reachable
- **THEN** portal exposes a tailnet-accessible endpoint that proxies traffic to
  the local service
- **AND** the endpoint remains private to the tailnet

#### Scenario: Tailnet proxy is available from non-CLI configuration
- **WHEN** a developer provides the service port via
  `~/.portal/config.yml` or `PORTAL_PORT` and the local service is reachable
- **THEN** portal exposes a tailnet-accessible endpoint that proxies traffic to
  the local service
- **AND** the endpoint remains private to the tailnet

### Requirement: Optional Public Proxy Via Tailscale Funnel
portal MUST support explicit opt-in public exposure of a local developer
service via Tailscale Funnel, while preserving proxying and request visibility
behavior, regardless of whether funnel is enabled by CLI, config file, or
environment variable. When a Funnel allowlist is configured, portal MUST enforce
it before proxying requests and MUST deny requests when a trustworthy client
source IP cannot be determined.

#### Scenario: Developer enables Funnel on a local service with CLI
- **WHEN** a developer runs `portal <port> --funnel`
- **THEN** portal configures Funnel-compatible serving for that service
- **AND** the service becomes reachable from the public internet

#### Scenario: Developer enables Funnel on a local service with config or environment
- **WHEN** a developer enables funnel in `~/.portal/config.yml` or sets
  `PORTAL_FUNNEL=true` with a valid service port
- **THEN** portal configures Funnel-compatible serving for that service
- **AND** the service becomes reachable from the public internet

#### Scenario: Request from allowlisted source is proxied in Funnel mode
- **WHEN** Funnel mode is enabled and the resolved client source IP matches a
  configured allowlist IP or CIDR
- **THEN** portal proxies the request to the local service

#### Scenario: Request from non-allowlisted source is denied in Funnel mode
- **WHEN** Funnel mode is enabled and the resolved client source IP does not
  match any configured allowlist IP or CIDR
- **THEN** portal returns an HTTP 403 response
- **AND** portal does not proxy the request to the local service

#### Scenario: Unresolvable source identity is denied in Funnel mode
- **WHEN** Funnel mode is enabled with an allowlist and no trustworthy client
  source IP can be resolved from available request metadata
- **THEN** portal returns an HTTP 403 response
- **AND** portal logs that the request was denied due to unresolved client
  identity
