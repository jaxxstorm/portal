## ADDED Requirements

### Requirement: TSNet Service Listen Mode
When running in tsnet mode, tgate MUST support an explicit service listening
mode that uses `tsnet.Server.ListenService` for tailnet/private exposure.

#### Scenario: Tailnet tsnet service mode serves traffic with ListenService
- **WHEN** a user runs tgate in tsnet mode with `listen-mode=service`,
  `funnel=false`, and a valid service name
- **THEN** tgate creates the serving listener using `ListenService`
- **AND** requests are proxied to the configured local or mock backend
- **AND** startup output identifies that tsnet service mode is active

#### Scenario: Funnel run with service mode configured is rejected
- **WHEN** a user runs tgate in tsnet mode with Funnel enabled and
  `listen-mode=service`
- **THEN** tgate fails startup with a configuration error indicating that
  service mode and Funnel are mutually exclusive
- **AND** tgate does not start serving traffic
