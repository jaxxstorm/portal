## MODIFIED Requirements

### Requirement: Tailnet Proxy For Local Developer Services
tgate MUST proxy a developer's local HTTP service to their Tailscale tailnet so authenticated tailnet users can access it securely, regardless of whether the service configuration is provided by CLI, `~/.tgate/config.yml`, or `TGATE_*` environment variables.

#### Scenario: Tailnet proxy is available from CLI configuration
- **WHEN** a developer runs `tgate <port>` and the local service is reachable
- **THEN** tgate exposes a tailnet-accessible endpoint that proxies traffic to the local service
- **AND** the endpoint remains private to the tailnet

#### Scenario: Tailnet proxy is available from non-CLI configuration
- **WHEN** a developer provides the service port via `~/.tgate/config.yml` or `TGATE_PORT` and the local service is reachable
- **THEN** tgate exposes a tailnet-accessible endpoint that proxies traffic to the local service
- **AND** the endpoint remains private to the tailnet

### Requirement: Optional Public Proxy Via Tailscale Funnel
tgate MUST support explicit opt-in public exposure of a local developer service via Tailscale Funnel, while preserving proxying and request visibility behavior, regardless of whether funnel is enabled by CLI, config file, or environment variable.

#### Scenario: Developer enables Funnel on a local service with CLI
- **WHEN** a developer runs `tgate <port> --funnel`
- **THEN** tgate configures Funnel-compatible serving for that service
- **AND** the service becomes reachable from the public internet

#### Scenario: Developer enables Funnel on a local service with config or environment
- **WHEN** a developer enables funnel in `~/.tgate/config.yml` or sets `TGATE_FUNNEL=true` with a valid service port
- **THEN** tgate configures Funnel-compatible serving for that service
- **AND** the service becomes reachable from the public internet

### Requirement: Clear Security Defaults Across Exposure Modes
tgate MUST default to private tailnet access and MUST require explicit user action to expose a service publicly across all supported configuration sources.

#### Scenario: No public exposure without explicit intent
- **WHEN** funnel is not explicitly enabled by CLI, config file, or `TGATE_*` environment variables
- **THEN** tgate does not create a public endpoint
- **AND** access is limited to tailnet clients

#### Scenario: Mock mode supports public webhook testing
- **WHEN** a developer runs `tgate --mock` or enables mock mode through supported non-CLI configuration
- **THEN** tgate provides a publicly reachable mock endpoint for external webhook providers
- **AND** request details are available for inspection
