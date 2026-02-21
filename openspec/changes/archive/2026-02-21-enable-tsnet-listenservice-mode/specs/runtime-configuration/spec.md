## ADDED Requirements

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
