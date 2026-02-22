## Purpose

Define deterministic startup readiness reporting so operators can reliably see
service reachability, web UI status, and enabled capabilities across TUI and
non-TUI execution modes.

## Requirements

### Requirement: Deterministic Startup Readiness Summary
`portal` MUST emit a single startup-ready summary after the serving path is active, and the summary MUST include execution mode, backend mode, exposure type, service reachability URL, web UI status, and enabled capabilities.

#### Scenario: Tailnet mode startup summary is explicit
- **WHEN** a user runs `portal <port>` with UI enabled
- **THEN** portal emits one startup-ready summary with `backend_mode=proxy` and `exposure=tailnet`
- **AND** the summary includes the reachable service URL
- **AND** the summary includes web UI status and URL (or explicit UI disabled/unavailable status)

#### Scenario: Funnel mode startup summary is explicit
- **WHEN** a user runs `portal <port> --funnel` and Funnel prerequisites are satisfied
- **THEN** portal emits one startup-ready summary with `backend_mode=proxy` and `exposure=funnel`
- **AND** the summary identifies public reachability
- **AND** the summary includes enabled capability flags (including Funnel)

#### Scenario: Mock mode startup summary separates backend and exposure
- **WHEN** a user runs `portal --mock` without Funnel enabled
- **THEN** portal emits one startup-ready summary with `backend_mode=mock` and `exposure=tailnet`
- **AND** the summary does not indicate public reachability

#### Scenario: Mock mode with Funnel reports public exposure explicitly
- **WHEN** a user runs `portal --mock --funnel` and Funnel prerequisites are satisfied
- **THEN** portal emits one startup-ready summary with `backend_mode=mock` and `exposure=funnel`
- **AND** the summary identifies public reachability

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

### Requirement: Startup Failure MUST Not Emit Ready Summary
`portal` MUST NOT emit a startup-ready summary when startup fails before serving is active.

#### Scenario: Funnel prerequisite failure blocks ready summary
- **WHEN** a user runs `portal <port> --funnel` and HTTPS certificate or Funnel setup fails
- **THEN** portal logs an explicit startup failure reason
- **AND** portal does not emit a startup-ready summary
