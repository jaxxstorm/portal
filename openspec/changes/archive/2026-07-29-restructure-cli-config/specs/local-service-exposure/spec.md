## ADDED Requirements

### Requirement: Command-Oriented Exposure Defaults
Portal MUST preserve tailnet-private exposure by default for `portal serve
<port>` and `portal mock`. Funnel-public exposure MUST require explicit
`--funnel` command configuration or an effective configuration-file or
environment setting.

#### Scenario: Serve command stays private by default
- **WHEN** a user runs `portal serve 8080` without effective Funnel configuration
- **THEN** Portal exposes the local service only to the tailnet

#### Scenario: Mock command enables public exposure explicitly
- **WHEN** a user runs `portal mock --funnel` and Funnel HTTPS prerequisites are satisfied
- **THEN** Portal exposes the mock endpoint through Funnel
- **AND** Portal reports the endpoint as publicly reachable
