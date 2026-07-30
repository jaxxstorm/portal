## ADDED Requirements

### Requirement: Command-Oriented Workflow Documentation
Project documentation MUST describe the command-oriented Portal interface and
the configuration-template workflow, including compatible legacy invocations.

#### Scenario: Serving workflow is documented
- **WHEN** a user reads Portal configuration documentation
- **THEN** the documentation includes `portal serve <port>` and `portal mock` examples
- **AND** the documentation explains that `portal <port>` and `portal --mock` remain supported

#### Scenario: Configuration template workflow is documented
- **WHEN** a user reads Portal configuration documentation
- **THEN** the documentation includes `portal config init` and explains the default file path, `--force` behavior, precedence, and secret-handling guidance

#### Scenario: Public exposure remains explicit in documentation
- **WHEN** a user reads command-oriented Funnel examples
- **THEN** the documentation shows `portal serve <port> --funnel` or equivalent explicit configuration
- **AND** the documentation calls out HTTPS prerequisites and public-exposure implications
