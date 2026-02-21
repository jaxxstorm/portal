## ADDED Requirements

### Requirement: Startup Output Documentation Coverage
Project documentation MUST define how startup-complete information is presented in TUI mode, non-TUI text logging mode, and non-TUI JSON logging mode.

#### Scenario: Tailnet startup output is documented
- **WHEN** a user reads startup and operating-mode documentation
- **THEN** the docs include a concrete `tgate <port>` startup example
- **AND** the docs explain where to find the effective web UI location when `--ui-port` is auto-assigned

#### Scenario: Funnel startup output is documented with security implications
- **WHEN** a user reads startup and operating-mode documentation for public exposure
- **THEN** the docs include a concrete `tgate <port> --funnel` startup example
- **AND** the docs call out that Funnel exposure is public and prerequisite-dependent
- **AND** the docs explain how startup output confirms enabled capabilities and reachability
