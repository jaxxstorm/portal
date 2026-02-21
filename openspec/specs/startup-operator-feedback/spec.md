## Purpose

Define deterministic startup readiness reporting so operators can reliably see
service reachability, web UI status, and enabled capabilities across TUI and
non-TUI execution modes.

## Requirements

### Requirement: Deterministic Startup Readiness Summary
`tgate` MUST emit a single startup-ready summary after the serving path is active, and the summary MUST include execution mode, exposure type, service reachability URL, web UI status, and enabled capabilities.

#### Scenario: Tailnet mode startup summary is explicit
- **WHEN** a user runs `tgate <port>` with UI enabled
- **THEN** tgate emits one startup-ready summary with `exposure=tailnet`
- **AND** the summary includes the reachable service URL
- **AND** the summary includes web UI status and URL (or explicit UI disabled/unavailable status)

#### Scenario: Funnel mode startup summary is explicit
- **WHEN** a user runs `tgate <port> --funnel` and Funnel prerequisites are satisfied
- **THEN** tgate emits one startup-ready summary with `exposure=funnel`
- **AND** the summary identifies public reachability
- **AND** the summary includes enabled capability flags (including Funnel)

### Requirement: Startup Summary MUST Be Rendered Consistently Across Output Modes
`tgate` MUST render the same startup summary semantics across TUI, non-TUI text logging, and non-TUI JSON logging.

#### Scenario: Non-TUI text output includes startup completion and reachability
- **WHEN** tgate runs with TUI disabled and JSON logging disabled
- **THEN** logs include a human-readable startup-complete message
- **AND** logs include the service URL and web UI status/location

#### Scenario: Non-TUI JSON output emits structured startup summary
- **WHEN** tgate runs with TUI disabled and JSON logging enabled
- **THEN** tgate emits a structured startup event with stable keys for mode, exposure, service URL, web UI status, and capabilities
- **AND** startup details are not emitted as unstructured decorative output

#### Scenario: TUI mode presents startup summary in the UI channel
- **WHEN** tgate runs with TUI enabled
- **THEN** startup summary information is visible in TUI-managed output
- **AND** startup details are not duplicated as conflicting non-TUI legacy output

### Requirement: Startup Failure MUST Not Emit Ready Summary
`tgate` MUST NOT emit a startup-ready summary when startup fails before serving is active.

#### Scenario: Funnel prerequisite failure blocks ready summary
- **WHEN** a user runs `tgate <port> --funnel` and HTTPS certificate or Funnel setup fails
- **THEN** tgate logs an explicit startup failure reason
- **AND** tgate does not emit a startup-ready summary
