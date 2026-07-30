## Purpose

Define how portal resolves runtime configuration from CLI arguments, environment
variables, and `~/.portal/config.yml` while preserving tailnet-default and
funnel-opt-in behavior.

## Requirements

### Requirement: Unified Runtime Configuration Sources
portal MUST resolve runtime configuration from defaults, `~/.portal/config.yml`, `PORTAL_*` environment variables, and CLI arguments/flags, including Funnel allowlist and authorization settings.

#### Scenario: Config file is loaded from the default path
- **WHEN** `~/.portal/config.yml` exists and the user runs portal without conflicting env or CLI values
- **THEN** portal uses values from `~/.portal/config.yml` as the effective runtime configuration

#### Scenario: Funnel allowlist is loaded from config file
- **WHEN** `~/.portal/config.yml` contains a Funnel allowlist and Funnel mode is enabled
- **THEN** portal applies the configured allowlist as part of effective runtime configuration

#### Scenario: HTTP Basic authorization is loaded from config file
- **WHEN** `~/.portal/config.yml` configures `auth-mode: basic` with an authorization username and password
- **THEN** portal applies those values as the effective endpoint authorization configuration

### Requirement: Deterministic Configuration Precedence
portal MUST apply deterministic precedence in this order: CLI arguments/flags, then environment variables, then config file values, then built-in defaults.

#### Scenario: Precedence resolves conflicting values
- **WHEN** `serve-port` is set to `80` in `~/.portal/config.yml`, `PORTAL_SERVE_PORT=443` is present, and the user runs `portal 8080 --serve-port 8443`
- **THEN** portal uses `serve-port=8443`

#### Scenario: Environment variables override config-file Funnel allowlist
- **WHEN** `~/.portal/config.yml` defines a Funnel allowlist and `PORTAL_FUNNEL_ALLOWLIST` is also set
- **THEN** portal uses `PORTAL_FUNNEL_ALLOWLIST` as the effective Funnel allowlist

### Requirement: Environment Variable Coverage
portal MUST provide `PORTAL_*` environment variable mappings for all supported runtime configuration keys, including Funnel allowlist and authorization configuration.

#### Scenario: Funnel is enabled via environment variable
- **WHEN** `PORTAL_PORT=8080` and `PORTAL_FUNNEL=true` are set and required HTTPS prerequisites are satisfied
- **THEN** portal starts with funnel enabled and exposes the configured service publicly

#### Scenario: Funnel allowlist is configured via environment variable
- **WHEN** `PORTAL_FUNNEL=true` and `PORTAL_FUNNEL_ALLOWLIST=203.0.113.10,198.51.100.0/24` are set with a valid service configuration
- **THEN** portal enables Funnel allowlist enforcement using those entries

#### Scenario: HTTP Basic authorization is configured via environment variables
- **WHEN** `PORTAL_AUTH_MODE=basic`, `PORTAL_AUTH_USERNAME`, and `PORTAL_AUTH_PASSWORD` are set with a valid service configuration
- **THEN** portal enables HTTP Basic authorization using those credentials

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
Portal MUST accept the public hostname and derive the fixed HTTPS callback URL
when its OAuth App bootstrap is used for `auth-mode=tailscale` with Funnel
enabled. Portal MUST retain existing startup failures for unavailable HTTPS
certificates or Funnel setup and MUST NOT report the endpoint as ready after
such a failure.

#### Scenario: Funnel startup fails before login is available
- **WHEN** a user runs `portal <port> --funnel` with `auth-mode=tailscale` and Funnel HTTPS certificate prerequisites are not satisfied
- **THEN** portal reports the HTTPS or Funnel setup failure
- **AND** Portal does not report a ready authenticated endpoint

### Requirement: Funnel Allowlist Entries Are Validated At Startup
portal MUST validate Funnel allowlist entries as IP addresses or CIDR blocks before accepting runtime configuration.

#### Scenario: Invalid Funnel allowlist entry fails startup
- **WHEN** Funnel allowlist configuration contains an invalid IP or CIDR value
- **THEN** portal exits with a configuration error
- **AND** the error identifies the invalid entry

### Requirement: Exposure Semantics Are Source-Agnostic
portal MUST preserve tailnet-default and funnel-opt-in semantics regardless of whether configuration comes from CLI, env vars, or config file.

#### Scenario: Tailnet-private default from config file
- **WHEN** `~/.portal/config.yml` sets a proxy target port and does not enable funnel
- **THEN** portal exposes the service to the tailnet only

### Requirement: Command-Aware Configuration Resolution
Portal MUST apply the existing configuration precedence to serving commands and
MUST validate only settings relevant to the selected command. OAuth App,
cleanup, and configuration-management commands MUST NOT require a service port
or validate a configured proxy target.

#### Scenario: Serve command uses configuration-file port
- **WHEN** `~/.portal/config.yml` supplies a valid port and a user runs `portal serve` without a positional port
- **THEN** Portal uses the configured port for serving

#### Scenario: Cleanup command ignores invalid serving configuration
- **WHEN** `~/.portal/config.yml` contains an invalid serving setting and a user runs `portal cleanup`
- **THEN** Portal performs cleanup without rejecting the unrelated serving setting

### Requirement: Configuration Template Generation
Portal MUST provide `portal config init` to generate a commented YAML template
at `~/.portal/config.yml`. Portal MUST create the parent directory as needed,
MUST generate a tailnet-private default configuration, and MUST NOT overwrite an
existing configuration file unless `--force` is provided.

#### Scenario: New configuration template is created
- **WHEN** a user runs `portal config init` and the default configuration file does not exist
- **THEN** Portal creates a commented YAML template with Funnel disabled by default
- **AND** Portal reports the generated file location

#### Scenario: Existing configuration is protected
- **WHEN** a user runs `portal config init` and the default configuration file already exists without `--force`
- **THEN** Portal exits with an error and leaves the existing file unchanged

#### Scenario: Explicit replacement is allowed
- **WHEN** a user runs `portal config init --force`
- **THEN** Portal replaces the existing configuration file with the template

### Requirement: TSNet ListenService Configuration Surface
portal MUST expose tsnet service-mode configuration through CLI flags,
`PORTAL_*` environment variables, and `~/.portal/config.yml`, with deterministic
precedence consistent with existing runtime configuration behavior.

#### Scenario: TSNet service mode is configured from non-CLI sources
- **WHEN** `~/.portal/config.yml` or `PORTAL_*` values set listen mode to
  `service` with a valid service name and no overriding CLI flags
- **THEN** portal resolves effective configuration with `service` mode enabled
- **AND** tsnet startup attempts `ListenService` behavior

#### Scenario: CLI overrides environment and config for tsnet listen mode
- **WHEN** config file and environment define listen mode values and the
  user provides a CLI listen-mode value
- **THEN** portal uses the CLI value as effective tsnet listen mode

### Requirement: TSNet Service Configuration Validation
portal MUST validate tsnet service-mode configuration at startup and MUST fail
fast on invalid service identifiers.

#### Scenario: Invalid service name fails startup
- **WHEN** listen mode is `service` and the configured service name is
  not a valid Tailscale service identifier
- **THEN** portal exits with a configuration error before serving starts
- **AND** the error identifies the invalid service value

#### Scenario: Service prerequisites failure is surfaced clearly
- **WHEN** listen mode is `service` with a valid service name but runtime
  prerequisites (for example tagged-node identity or service approval) are not
  satisfied
- **THEN** portal fails startup with a clear runtime error
- **AND** no startup-ready event is emitted

#### Scenario: Service mode and Funnel conflict fails startup
- **WHEN** listen mode is `service` and Funnel is enabled in effective
  configuration
- **THEN** portal fails startup with a configuration error indicating the
  combination is not allowed
- **AND** no startup-ready event is emitted
