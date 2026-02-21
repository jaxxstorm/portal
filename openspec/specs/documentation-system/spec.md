## Purpose

Define tgate's documentation system so detailed guidance lives in `docs/`,
`README.md` remains concise, and future user-visible changes include matching
documentation updates.

## Requirements

### Requirement: Docs Site For Detailed Project Documentation
The project MUST provide a Docsify-based documentation site under `docs/` as the primary home for detailed documentation.

#### Scenario: Docsify documentation site exists
- **WHEN** a contributor opens the repository documentation assets
- **THEN** a Docsify-compatible documentation structure exists under `docs/`
- **AND** the docs home page links to operating modes and configuration documentation

### Requirement: README Is A Concise Entry Point
`README.md` MUST be streamlined to a concise project entry point and MUST not include emojis.

#### Scenario: README uses clean, concise style
- **WHEN** a user reads `README.md`
- **THEN** the README presents a compact overview and quick-start information
- **AND** the README contains no emoji characters
- **AND** the README links users to `docs/` for detailed guidance

### Requirement: Operating Mode Documentation Coverage
Project documentation MUST describe tgate operating modes, including both tailnet-private and funnel-public usage.

#### Scenario: Tailnet mode behavior is documented
- **WHEN** a user reads operating mode documentation
- **THEN** the documentation explains tailnet-private default behavior and how to run `tgate <port>`

#### Scenario: Funnel mode behavior is documented
- **WHEN** a user reads operating mode documentation
- **THEN** the documentation explains funnel opt-in behavior and how to run `tgate <port> --funnel`
- **AND** the documentation calls out funnel prerequisites and public exposure implications

### Requirement: Configuration Documentation Coverage
Project documentation MUST explain the configuration model, including config file location, environment variable format, and precedence rules.

#### Scenario: Configuration sources and precedence are documented
- **WHEN** a user reads configuration documentation
- **THEN** the documentation explains `~/.tgate/config.yml`
- **AND** the documentation explains `TGATE_*` environment variable mapping
- **AND** the documentation explains precedence order of CLI over environment over config file over defaults

### Requirement: Documentation Maintenance Expectation
The project MUST define a contributor-facing expectation that future feature changes include corresponding documentation updates in `docs/`.

#### Scenario: Future behavior changes include docs updates
- **WHEN** a contributor introduces or changes user-visible behavior
- **THEN** contributor guidance states that relevant `docs/` content is updated as part of the same change

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
