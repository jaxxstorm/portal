## Purpose

Define portal's core product behavior as "ngrok over Tailscale": proxy local developer services privately to the tailnet by default, with explicit opt-in public exposure via Tailscale Funnel.
## Requirements
### Requirement: Tailnet Proxy For Local Developer Services
portal MUST proxy a developer's local HTTP service to their Tailscale tailnet so authenticated tailnet users can access it securely, regardless of whether the service configuration is provided by CLI, `~/.portal/config.yml`, or `PORTAL_*` environment variables. Funnel request allowlist controls MUST NOT affect tailnet-private request handling.

#### Scenario: Tailnet proxy is available from CLI configuration
- **WHEN** a developer runs `portal <port>` and the local service is reachable
- **THEN** portal exposes a tailnet-accessible endpoint that proxies traffic to the local service
- **AND** the endpoint remains private to the tailnet

#### Scenario: Tailnet proxy is available from non-CLI configuration
- **WHEN** a developer provides the service port via `~/.portal/config.yml` or `PORTAL_PORT` and the local service is reachable
- **THEN** portal exposes a tailnet-accessible endpoint that proxies traffic to the local service
- **AND** the endpoint remains private to the tailnet

### Requirement: Optional Public Proxy Via Tailscale Funnel
portal MUST support explicit opt-in public exposure of a local developer service via Tailscale Funnel, while preserving proxying and request visibility behavior, regardless of whether funnel is enabled by CLI, config file, or environment variable. When a Funnel allowlist is configured, portal MUST enforce it before proxying requests and MUST deny requests when a trustworthy client source IP cannot be determined.

#### Scenario: Developer enables Funnel on a local service with CLI
- **WHEN** a developer runs `portal <port> --funnel`
- **THEN** portal configures Funnel-compatible serving for that service
- **AND** the service becomes reachable from the public internet

#### Scenario: Developer enables Funnel on a local service with config or environment
- **WHEN** a developer enables funnel in `~/.portal/config.yml` or sets `PORTAL_FUNNEL=true` with a valid service port
- **THEN** portal configures Funnel-compatible serving for that service
- **AND** the service becomes reachable from the public internet

#### Scenario: Request from allowlisted source is proxied in Funnel mode
- **WHEN** Funnel mode is enabled and the resolved client source IP matches a configured allowlist IP or CIDR
- **THEN** portal proxies the request to the local service

#### Scenario: Request from non-allowlisted source is denied in Funnel mode
- **WHEN** Funnel mode is enabled and the resolved client source IP does not match any configured allowlist IP or CIDR
- **THEN** portal returns an HTTP 403 response
- **AND** portal does not proxy the request to the local service

#### Scenario: Unresolvable source identity is denied in Funnel mode
- **WHEN** Funnel mode is enabled with an allowlist and no trustworthy client source IP can be resolved from available request metadata
- **THEN** portal returns an HTTP 403 response
- **AND** portal logs that the request was denied due to unresolved client identity

### Requirement: Clear Security Defaults Across Exposure Modes
portal MUST default to private tailnet access and MUST require explicit user action to expose a service publicly across all supported configuration sources, including mock backend runs.

#### Scenario: No public exposure without explicit intent
- **WHEN** funnel is not explicitly enabled by CLI, config file, or `PORTAL_*` environment variables
- **THEN** portal does not create a public endpoint
- **AND** access is limited to tailnet clients

#### Scenario: Mock mode defaults to tailnet-private exposure
- **WHEN** a developer runs `portal --mock` or enables mock mode through supported non-CLI configuration without enabling Funnel
- **THEN** portal serves mock responses using a tailnet-private endpoint
- **AND** portal does not create a public endpoint
- **AND** request details remain available for inspection

#### Scenario: Mock mode supports explicit Funnel opt-in
- **WHEN** a developer runs `portal --mock --funnel` or enables both mock mode and Funnel through supported non-CLI configuration
- **THEN** portal configures Funnel-compatible serving for the mock backend
- **AND** the mock endpoint is reachable from the public internet
- **AND** Funnel security controls and prerequisites apply

### Requirement: Unified Tailscale Lifecycle Logging
portal MUST emit Tailscale and tsnet lifecycle logs through the same primary application logger channel used for other runtime events. This requirement MUST apply in tailnet mode and Funnel mode, and log entries MUST include structured context that identifies the Tailscale-related component and startup phase.

#### Scenario: Tailnet mode startup logs stay in the primary channel
- **WHEN** a developer runs `portal <port>` without Funnel enabled and Tailscale initialization succeeds
- **THEN** Tailscale/tsnet startup lifecycle entries are emitted through the main application logger channel
- **AND** the application does not emit separate out-of-band Tailscale lifecycle output to a second stdout/stderr stream

#### Scenario: Funnel setup failures are visible in the primary channel
- **WHEN** a developer runs `portal <port> --funnel` and HTTPS certificate or Funnel setup fails
- **THEN** portal emits a structured error entry for that failure through the main application logger channel
- **AND** the failure entry includes enough context to identify that Funnel startup could not complete

### Requirement: TSNet Service Listen Mode
When running in tsnet mode, portal MUST support an explicit service listening
mode that uses `tsnet.Server.ListenService` for tailnet/private exposure.

#### Scenario: Tailnet tsnet service mode serves traffic with ListenService
- **WHEN** a user runs portal in tsnet mode with `listen-mode=service`,
  `funnel=false`, and a valid service name
- **THEN** portal creates the serving listener using `ListenService`
- **AND** requests are proxied to the configured local or mock backend
- **AND** startup output identifies that tsnet service mode is active

#### Scenario: Funnel run with service mode configured is rejected
- **WHEN** a user runs portal in tsnet mode with Funnel enabled and
  `listen-mode=service`
- **THEN** portal fails startup with a configuration error indicating that
  service mode and Funnel are mutually exclusive
- **AND** portal does not start serving traffic
