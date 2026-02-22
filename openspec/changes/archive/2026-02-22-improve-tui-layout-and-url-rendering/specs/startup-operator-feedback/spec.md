## MODIFIED Requirements

### Requirement: Startup Summary MUST Be Rendered Consistently Across Output Modes
`portal` MUST render the same startup summary semantics across TUI, non-TUI text logging, and non-TUI JSON logging. In TUI mode, startup summary fields for service and Web UI reachability MUST be presented with dedicated labels and readable URL formatting.

#### Scenario: Non-TUI text output includes startup completion and reachability
- **WHEN** portal runs with TUI disabled and JSON logging disabled
- **THEN** logs include a human-readable startup-complete message
- **AND** logs include the service URL and web UI status/location

#### Scenario: Non-TUI JSON output emits structured startup summary
- **WHEN** portal runs with TUI disabled and JSON logging enabled
- **THEN** portal emits a structured startup event with stable keys for mode, exposure, service URL, web UI status, and capabilities
- **AND** startup details are not emitted as unstructured decorative output

#### Scenario: TUI mode presents startup summary in the UI channel
- **WHEN** portal runs with TUI enabled
- **THEN** startup summary information is visible in TUI-managed output
- **AND** startup details are not duplicated as conflicting non-TUI legacy output
- **AND** service and Web UI endpoint fields are rendered with stable labels and readable URL formatting

#### Scenario: TUI mode renders unavailable Web UI status explicitly
- **WHEN** portal runs with TUI enabled and Web UI is unavailable in the active serving mode
- **THEN** startup summary output in TUI includes explicit web UI unavailable status
- **AND** the output includes a reason instead of presenting an empty or ambiguous endpoint field
