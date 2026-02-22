## MODIFIED Requirements

### Requirement: Clear Security Defaults Across Exposure Modes
tgate MUST default to private tailnet access and MUST require explicit user action to expose a service publicly across all supported configuration sources, including mock backend runs.

#### Scenario: No public exposure without explicit intent
- **WHEN** funnel is not explicitly enabled by CLI, config file, or `TGATE_*` environment variables
- **THEN** tgate does not create a public endpoint
- **AND** access is limited to tailnet clients

#### Scenario: Mock mode defaults to tailnet-private exposure
- **WHEN** a developer runs `tgate --mock` or enables mock mode through supported non-CLI configuration without enabling Funnel
- **THEN** tgate serves mock responses using a tailnet-private endpoint
- **AND** tgate does not create a public endpoint
- **AND** request details remain available for inspection

#### Scenario: Mock mode supports explicit Funnel opt-in
- **WHEN** a developer runs `tgate --mock --funnel` or enables both mock mode and Funnel through supported non-CLI configuration
- **THEN** tgate configures Funnel-compatible serving for the mock backend
- **AND** the mock endpoint is reachable from the public internet
- **AND** Funnel security controls and prerequisites apply
