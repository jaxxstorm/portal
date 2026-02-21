## ADDED Requirements

### Requirement: TSNet Service Mode Documentation Coverage
Project documentation MUST describe tsnet service mode configuration,
prerequisites, and behavior in relation to tailnet and Funnel exposure modes.

#### Scenario: Service mode setup and prerequisites are documented
- **WHEN** a user reads operating modes and configuration documentation
- **THEN** docs explain how to enable `listen-mode=service`
- **AND** docs describe service-name format requirements
- **AND** docs note legacy `tsnet-*` aliases remain supported
- **AND** docs call out tagged-node and service-approval prerequisites

#### Scenario: Funnel interaction with service mode is documented
- **WHEN** a user reads documentation for Funnel and tsnet service mode
- **THEN** docs explain that Funnel and service mode are mutually exclusive
- **AND** docs describe that startup fails configuration validation when both
  are configured together
- **AND** docs include valid alternatives for public exposure and service
  exposure use cases
