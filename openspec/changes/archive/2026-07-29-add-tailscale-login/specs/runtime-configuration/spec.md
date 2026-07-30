## MODIFIED Requirements

### Requirement: Authorization Configuration Validation
portal MUST support `none`, `basic`, and `tailscale` authorization modes and
MUST validate authorization configuration before serving. `basic` MUST require a
non-empty username and password. `tailscale` MUST require non-empty Tailscale
OAuth client ID, OAuth client secret, and session signing secret. `none` MUST
reject supplied Basic or Tailscale authorization credentials, and `basic` and
`tailscale` MUST reject credentials belonging to the other provider.

#### Scenario: Incomplete HTTP Basic configuration fails startup
- **WHEN** `auth-mode=basic` is configured without a non-empty username or password
- **THEN** portal exits with a configuration error that identifies the missing authorization setting

#### Scenario: Incomplete Tailscale configuration fails startup
- **WHEN** `auth-mode=tailscale` is configured without a non-empty OAuth client ID, OAuth client secret, or session signing secret
- **THEN** portal exits with a configuration error that identifies the missing authorization setting

#### Scenario: Tailscale credentials with disabled authorization fail startup
- **WHEN** `auth-mode=none` is configured with any Tailscale OAuth or session signing credential
- **THEN** portal exits with a configuration error that identifies that Tailscale credentials require `auth-mode=tailscale`

#### Scenario: Unsupported authorization mode fails startup
- **WHEN** an authorization mode other than `none`, `basic`, or `tailscale` is configured
- **THEN** portal exits with a configuration error that identifies the unsupported mode

## ADDED Requirements

### Requirement: Tailscale Authorization Settings Are Available From All Sources
portal MUST resolve Tailscale authorization client ID, client secret, and session
signing secret from CLI flags, `PORTAL_*` environment variables, and
`~/.portal/config.yml` with the existing deterministic configuration precedence.

#### Scenario: Tailscale authorization is configured from environment
- **WHEN** `PORTAL_AUTH_MODE=tailscale`, `PORTAL_TAILSCALE_OAUTH_CLIENT_ID`, `PORTAL_TAILSCALE_OAUTH_CLIENT_SECRET`, and `PORTAL_TAILSCALE_SESSION_SECRET` are set with a valid service configuration
- **THEN** portal starts with Tailscale endpoint authorization enabled

#### Scenario: CLI Tailscale credentials override lower-precedence sources
- **WHEN** configuration file and environment define Tailscale authorization settings and the user supplies a CLI value for one of those settings
- **THEN** portal uses the CLI value for that setting

### Requirement: Funnel Login Prerequisites Are Clear
When `auth-mode=tailscale` and Funnel are enabled, portal MUST report that the
Tailscale OAuth application needs the exact public HTTPS callback URL. Portal
MUST retain existing startup failures for unavailable HTTPS certificates or
Funnel setup and MUST NOT report the endpoint as ready after such a failure.

#### Scenario: Funnel startup fails before login is available
- **WHEN** a user runs `portal <port> --funnel` with `auth-mode=tailscale` and Funnel HTTPS certificate prerequisites are not satisfied
- **THEN** portal reports the HTTPS or Funnel setup failure
- **AND** Portal does not report a ready authenticated endpoint
