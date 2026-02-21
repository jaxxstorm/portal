## ADDED Requirements

### Requirement: Unified Tailscale Lifecycle Logging
tgate MUST emit Tailscale and tsnet lifecycle logs through the same primary application logger channel used for other runtime events. This requirement MUST apply in tailnet mode and Funnel mode, and log entries MUST include structured context that identifies the Tailscale-related component and startup phase.

#### Scenario: Tailnet mode startup logs stay in the primary channel
- **WHEN** a developer runs `tgate <port>` without Funnel enabled and Tailscale initialization succeeds
- **THEN** Tailscale/tsnet startup lifecycle entries are emitted through the main application logger channel
- **AND** the application does not emit separate out-of-band Tailscale lifecycle output to a second stdout/stderr stream

#### Scenario: Funnel setup failures are visible in the primary channel
- **WHEN** a developer runs `tgate <port> --funnel` and HTTPS certificate or Funnel setup fails
- **THEN** tgate emits a structured error entry for that failure through the main application logger channel
- **AND** the failure entry includes enough context to identify that Funnel startup could not complete
