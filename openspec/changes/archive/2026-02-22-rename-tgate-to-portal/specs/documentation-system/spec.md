## MODIFIED Requirements

### Requirement: Operating Mode Documentation Coverage
Project documentation MUST describe portal operating modes, including
tailnet-private and funnel-public usage, and MUST describe mock backend behavior
independently from exposure mode.

#### Scenario: Tailnet mode behavior is documented
- **WHEN** a user reads operating mode documentation
- **THEN** the documentation explains tailnet-private default behavior and how
  to run `portal <port>`

#### Scenario: Funnel mode behavior is documented
- **WHEN** a user reads operating mode documentation
- **THEN** the documentation explains funnel opt-in behavior and how to run
  `portal <port> --funnel`
- **AND** the documentation calls out funnel prerequisites and public exposure
  implications

#### Scenario: Mock backend behavior is documented as exposure-independent
- **WHEN** a user reads operating mode documentation for mock workflows
- **THEN** the documentation explains that `--mock` selects backend simulation
  only
- **AND** the documentation states that `portal --mock` is tailnet-private by
  default
- **AND** the documentation includes explicit public mock guidance using
  `portal --mock --funnel`

### Requirement: Configuration Documentation Coverage
Project documentation MUST explain the configuration model, including config
file location, environment variable format, and precedence rules, using portal
namespaces and paths.

#### Scenario: Configuration sources and precedence are documented
- **WHEN** a user reads configuration documentation
- **THEN** the documentation explains `~/.portal/config.yml`
- **AND** the documentation explains `PORTAL_*` environment variable mapping
- **AND** the documentation explains precedence order of CLI over environment
  over config file over defaults

### Requirement: Startup Output Documentation Coverage
Project documentation MUST define how startup-complete information is presented
in TUI mode, non-TUI text logging mode, and non-TUI JSON logging mode.

#### Scenario: Tailnet startup output is documented
- **WHEN** a user reads startup and operating-mode documentation
- **THEN** the docs include a concrete `portal <port>` startup example
- **AND** the docs explain where to find the effective web UI location when
  `--ui-port` is auto-assigned

#### Scenario: Funnel startup output is documented with security implications
- **WHEN** a user reads startup and operating-mode documentation for public
  exposure
- **THEN** the docs include a concrete `portal <port> --funnel` startup example
- **AND** the docs call out that Funnel exposure is public and
  prerequisite-dependent
- **AND** the docs explain how startup output confirms enabled capabilities and
  reachability
