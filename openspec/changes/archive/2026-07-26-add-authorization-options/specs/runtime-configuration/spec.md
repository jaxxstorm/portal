## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Authorization Configuration Validation
portal MUST support `none` and `basic` authorization modes and MUST validate
authorization configuration before serving. `basic` MUST require a non-empty
username and password, while `none` MUST reject supplied authorization
credentials.

#### Scenario: Incomplete HTTP Basic configuration fails startup
- **WHEN** `auth-mode=basic` is configured without a non-empty username or password
- **THEN** portal exits with a configuration error that identifies the missing authorization setting

#### Scenario: Unsupported authorization mode fails startup
- **WHEN** an authorization mode other than `none` or `basic` is configured
- **THEN** portal exits with a configuration error that identifies the unsupported mode
