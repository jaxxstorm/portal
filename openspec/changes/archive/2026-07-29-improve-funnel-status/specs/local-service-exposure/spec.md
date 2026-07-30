## ADDED Requirements

### Requirement: Funnel Certificate And Readiness Feedback
When a user explicitly enables Funnel, portal MUST report whether its HTTPS
certificate prerequisite and Funnel serving setup are verified before identifying
the endpoint as publicly reachable. The status MUST be emitted through the
primary application logger and included in the startup summary across local
Tailscale and tsnet modes.

#### Scenario: Funnel certificate and public endpoint are confirmed
- **WHEN** a user runs `portal <port> --funnel` and HTTPS certificate and Funnel setup succeed
- **THEN** portal reports that the certificate is verified and Funnel is ready
- **AND** portal reports the public service URL only after Funnel serving is active

#### Scenario: Certificate prerequisite failure is explicit
- **WHEN** a user runs `portal <port> --funnel` and the HTTPS certificate prerequisite cannot be confirmed
- **THEN** portal reports that certificate verification failed and includes the prerequisite error
- **AND** portal does not report the service as publicly reachable

#### Scenario: Tailnet mode does not claim Funnel readiness
- **WHEN** a user runs `portal <port>` without Funnel enabled
- **THEN** portal reports that Funnel is disabled and certificate verification is not required
- **AND** portal does not report a public endpoint
