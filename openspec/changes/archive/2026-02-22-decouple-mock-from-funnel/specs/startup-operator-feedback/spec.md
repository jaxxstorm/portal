## MODIFIED Requirements

### Requirement: Deterministic Startup Readiness Summary
`tgate` MUST emit a single startup-ready summary after the serving path is active, and the summary MUST include execution mode, backend mode, exposure type, service reachability URL, web UI status, and enabled capabilities.

#### Scenario: Tailnet mode startup summary is explicit
- **WHEN** a user runs `tgate <port>` with UI enabled
- **THEN** tgate emits one startup-ready summary with `backend_mode=proxy` and `exposure=tailnet`
- **AND** the summary includes the reachable service URL
- **AND** the summary includes web UI status and URL (or explicit UI disabled/unavailable status)

#### Scenario: Funnel mode startup summary is explicit
- **WHEN** a user runs `tgate <port> --funnel` and Funnel prerequisites are satisfied
- **THEN** tgate emits one startup-ready summary with `backend_mode=proxy` and `exposure=funnel`
- **AND** the summary identifies public reachability
- **AND** the summary includes enabled capability flags (including Funnel)

#### Scenario: Mock mode startup summary separates backend and exposure
- **WHEN** a user runs `tgate --mock` without Funnel enabled
- **THEN** tgate emits one startup-ready summary with `backend_mode=mock` and `exposure=tailnet`
- **AND** the summary does not indicate public reachability

#### Scenario: Mock mode with Funnel reports public exposure explicitly
- **WHEN** a user runs `tgate --mock --funnel` and Funnel prerequisites are satisfied
- **THEN** tgate emits one startup-ready summary with `backend_mode=mock` and `exposure=funnel`
- **AND** the summary identifies public reachability
