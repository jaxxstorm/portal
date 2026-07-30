## ADDED Requirements

### Requirement: Command-Oriented Portal Interface
Portal MUST provide discoverable `serve`, `mock`, `oauth-app`, `cleanup`, and
`config` commands. Each command's help output MUST describe only settings and
arguments relevant to that command.

#### Scenario: Serve command exposes a local service privately
- **WHEN** a user runs `portal serve 8080` without Funnel configuration
- **THEN** Portal starts proxy serving for local port 8080
- **AND** the endpoint remains tailnet-private

#### Scenario: Mock command avoids a local service port
- **WHEN** a user runs `portal mock` without Funnel configuration
- **THEN** Portal starts the mock backend
- **AND** the endpoint remains tailnet-private

#### Scenario: OAuth App commands do not require serving settings
- **WHEN** a user runs `portal oauth-app create --help`
- **THEN** Portal displays OAuth App bootstrap options
- **AND** Portal does not require a service port or validate serving configuration

### Requirement: Legacy Serving Invocations Remain Supported
Portal MUST preserve `portal <port>`, `portal --mock`, and `portal
--cleanup-serve` behavior as compatibility invocations that use the same
underlying handlers as the corresponding command-oriented interface.

#### Scenario: Positional port remains compatible
- **WHEN** a user runs `portal 8080`
- **THEN** Portal behaves equivalently to `portal serve 8080`

#### Scenario: Legacy cleanup remains compatible
- **WHEN** a user runs `portal --cleanup-serve`
- **THEN** Portal performs the same cleanup action as `portal cleanup`
