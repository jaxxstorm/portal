## Purpose

Define how tgate resolves runtime configuration from CLI arguments, environment
variables, and `~/.tgate/config.yml` while preserving tailnet-default and
funnel-opt-in behavior.

## Requirements

### Requirement: Unified Runtime Configuration Sources
tgate MUST resolve runtime configuration from defaults, `~/.tgate/config.yml`, `TGATE_*` environment variables, and CLI arguments/flags, including Funnel allowlist settings.

#### Scenario: Config file is loaded from the default path
- **WHEN** `~/.tgate/config.yml` exists and the user runs tgate without conflicting env or CLI values
- **THEN** tgate uses values from `~/.tgate/config.yml` as the effective runtime configuration

#### Scenario: Funnel allowlist is loaded from config file
- **WHEN** `~/.tgate/config.yml` contains a Funnel allowlist and Funnel mode is enabled
- **THEN** tgate applies the configured allowlist as part of effective runtime configuration

### Requirement: Deterministic Configuration Precedence
tgate MUST apply deterministic precedence in this order: CLI arguments/flags, then environment variables, then config file values, then built-in defaults.

#### Scenario: Precedence resolves conflicting values
- **WHEN** `serve-port` is set to `80` in `~/.tgate/config.yml`, `TGATE_SERVE_PORT=443` is present, and the user runs `tgate 8080 --serve-port 8443`
- **THEN** tgate uses `serve-port=8443`

#### Scenario: Environment variables override config-file Funnel allowlist
- **WHEN** `~/.tgate/config.yml` defines a Funnel allowlist and `TGATE_FUNNEL_ALLOWLIST` is also set
- **THEN** tgate uses `TGATE_FUNNEL_ALLOWLIST` as the effective Funnel allowlist

### Requirement: Environment Variable Coverage
tgate MUST provide `TGATE_*` environment variable mappings for all supported runtime configuration keys, including Funnel allowlist configuration.

#### Scenario: Funnel is enabled via environment variable
- **WHEN** `TGATE_PORT=8080` and `TGATE_FUNNEL=true` are set and required HTTPS prerequisites are satisfied
- **THEN** tgate starts with funnel enabled and exposes the configured service publicly

#### Scenario: Funnel allowlist is configured via environment variable
- **WHEN** `TGATE_FUNNEL=true` and `TGATE_FUNNEL_ALLOWLIST=203.0.113.10,198.51.100.0/24` are set with a valid service configuration
- **THEN** tgate enables Funnel allowlist enforcement using those entries

### Requirement: Funnel Allowlist Entries Are Validated At Startup
tgate MUST validate Funnel allowlist entries as IP addresses or CIDR blocks before accepting runtime configuration.

#### Scenario: Invalid Funnel allowlist entry fails startup
- **WHEN** Funnel allowlist configuration contains an invalid IP or CIDR value
- **THEN** tgate exits with a configuration error
- **AND** the error identifies the invalid entry

### Requirement: Exposure Semantics Are Source-Agnostic
tgate MUST preserve tailnet-default and funnel-opt-in semantics regardless of whether configuration comes from CLI, env vars, or config file.

#### Scenario: Tailnet-private default from config file
- **WHEN** `~/.tgate/config.yml` sets a proxy target port and does not enable funnel
- **THEN** tgate exposes the service to the tailnet only

### Requirement: TSNet ListenService Configuration Surface
tgate MUST expose tsnet service-mode configuration through CLI flags,
`TGATE_*` environment variables, and `~/.tgate/config.yml`, with deterministic
precedence consistent with existing runtime configuration behavior.

#### Scenario: TSNet service mode is configured from non-CLI sources
- **WHEN** `~/.tgate/config.yml` or `TGATE_*` values set listen mode to
  `service` with a valid service name and no overriding CLI flags
- **THEN** tgate resolves effective configuration with `service` mode enabled
- **AND** tsnet startup attempts `ListenService` behavior

#### Scenario: CLI overrides environment and config for tsnet listen mode
- **WHEN** config file and environment define listen mode values and the
  user provides a CLI listen-mode value
- **THEN** tgate uses the CLI value as effective tsnet listen mode

### Requirement: TSNet Service Configuration Validation
tgate MUST validate tsnet service-mode configuration at startup and MUST fail
fast on invalid service identifiers.

#### Scenario: Invalid service name fails startup
- **WHEN** listen mode is `service` and the configured service name is
  not a valid Tailscale service identifier
- **THEN** tgate exits with a configuration error before serving starts
- **AND** the error identifies the invalid service value

#### Scenario: Service prerequisites failure is surfaced clearly
- **WHEN** listen mode is `service` with a valid service name but runtime
  prerequisites (for example tagged-node identity or service approval) are not
  satisfied
- **THEN** tgate fails startup with a clear runtime error
- **AND** no startup-ready event is emitted

#### Scenario: Service mode and Funnel conflict fails startup
- **WHEN** listen mode is `service` and Funnel is enabled in effective
  configuration
- **THEN** tgate fails startup with a configuration error indicating the
  combination is not allowed
- **AND** no startup-ready event is emitted
