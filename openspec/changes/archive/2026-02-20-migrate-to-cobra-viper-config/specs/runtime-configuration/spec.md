## ADDED Requirements

### Requirement: Unified Runtime Configuration Sources
tgate MUST resolve runtime configuration from defaults, `~/.tgate/config.yml`, `TGATE_*` environment variables, and CLI arguments/flags.

#### Scenario: Config file is loaded from the default path
- **WHEN** `~/.tgate/config.yml` exists and the user runs tgate without conflicting env or CLI values
- **THEN** tgate uses values from `~/.tgate/config.yml` as the effective runtime configuration

### Requirement: Deterministic Configuration Precedence
tgate MUST apply deterministic precedence in this order: CLI arguments/flags, then environment variables, then config file values, then built-in defaults.

#### Scenario: Precedence resolves conflicting values
- **WHEN** `serve-port` is set to `80` in `~/.tgate/config.yml`, `TGATE_SERVE_PORT=443` is present, and the user runs `tgate 8080 --serve-port 8443`
- **THEN** tgate uses `serve-port=8443`

### Requirement: Environment Variable Coverage
tgate MUST provide `TGATE_*` environment variable mappings for all supported runtime configuration keys.

#### Scenario: Funnel is enabled via environment variable
- **WHEN** `TGATE_PORT=8080` and `TGATE_FUNNEL=true` are set and required HTTPS prerequisites are satisfied
- **THEN** tgate starts with funnel enabled and exposes the configured service publicly

### Requirement: Exposure Semantics Are Source-Agnostic
tgate MUST preserve tailnet-default and funnel-opt-in semantics regardless of whether configuration comes from CLI, env vars, or config file.

#### Scenario: Tailnet-private default from config file
- **WHEN** `~/.tgate/config.yml` sets a proxy target port and does not enable funnel
- **THEN** tgate exposes the service to the tailnet only
