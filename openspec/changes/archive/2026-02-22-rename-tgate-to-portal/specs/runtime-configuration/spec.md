## MODIFIED Requirements

### Requirement: Unified Runtime Configuration Sources
portal MUST resolve runtime configuration from defaults,
`~/.portal/config.yml`, `PORTAL_*` environment variables, and CLI
arguments/flags, including Funnel allowlist settings.

#### Scenario: Config file is loaded from the default path
- **WHEN** `~/.portal/config.yml` exists and the user runs portal without
  conflicting env or CLI values
- **THEN** portal uses values from `~/.portal/config.yml` as the effective
  runtime configuration

#### Scenario: Funnel allowlist is loaded from config file
- **WHEN** `~/.portal/config.yml` contains a Funnel allowlist and Funnel mode
  is enabled
- **THEN** portal applies the configured allowlist as part of effective runtime
  configuration

### Requirement: Deterministic Configuration Precedence
portal MUST apply deterministic precedence in this order: CLI
arguments/flags, then environment variables, then config file values, then
built-in defaults.

#### Scenario: Precedence resolves conflicting values
- **WHEN** `serve-port` is set to `80` in `~/.portal/config.yml`,
  `PORTAL_SERVE_PORT=443` is present, and the user runs
  `portal 8080 --serve-port 8443`
- **THEN** portal uses `serve-port=8443`

#### Scenario: Environment variables override config-file Funnel allowlist
- **WHEN** `~/.portal/config.yml` defines a Funnel allowlist and
  `PORTAL_FUNNEL_ALLOWLIST` is also set
- **THEN** portal uses `PORTAL_FUNNEL_ALLOWLIST` as the effective Funnel
  allowlist

### Requirement: Environment Variable Coverage
portal MUST provide `PORTAL_*` environment variable mappings for all supported
runtime configuration keys, including Funnel allowlist configuration.

#### Scenario: Funnel is enabled via environment variable
- **WHEN** `PORTAL_PORT=8080` and `PORTAL_FUNNEL=true` are set and required
  HTTPS prerequisites are satisfied
- **THEN** portal starts with funnel enabled and exposes the configured service
  publicly

#### Scenario: Funnel allowlist is configured via environment variable
- **WHEN** `PORTAL_FUNNEL=true` and
  `PORTAL_FUNNEL_ALLOWLIST=203.0.113.10,198.51.100.0/24` are set with a valid
  service configuration
- **THEN** portal enables Funnel allowlist enforcement using those entries

### Requirement: TSNet ListenService Configuration Surface
portal MUST expose tsnet service-mode configuration through CLI flags,
`PORTAL_*` environment variables, and `~/.portal/config.yml`, with
deterministic precedence consistent with existing runtime configuration
behavior.

#### Scenario: TSNet service mode is configured from non-CLI sources
- **WHEN** `~/.portal/config.yml` or `PORTAL_*` values set listen mode to
  `service` with a valid service name and no overriding CLI flags
- **THEN** portal resolves effective configuration with `service` mode enabled
- **AND** tsnet startup attempts `ListenService` behavior

#### Scenario: CLI overrides environment and config for tsnet listen mode
- **WHEN** config file and environment define listen mode values and the user
  provides a CLI listen-mode value
- **THEN** portal uses the CLI value as effective tsnet listen mode
