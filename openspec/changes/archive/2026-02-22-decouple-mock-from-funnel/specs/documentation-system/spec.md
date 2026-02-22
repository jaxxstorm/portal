## MODIFIED Requirements

### Requirement: Operating Mode Documentation Coverage
Project documentation MUST describe tgate operating modes, including tailnet-private and funnel-public usage, and MUST describe mock backend behavior independently from exposure mode.

#### Scenario: Tailnet mode behavior is documented
- **WHEN** a user reads operating mode documentation
- **THEN** the documentation explains tailnet-private default behavior and how to run `tgate <port>`

#### Scenario: Funnel mode behavior is documented
- **WHEN** a user reads operating mode documentation
- **THEN** the documentation explains funnel opt-in behavior and how to run `tgate <port> --funnel`
- **AND** the documentation calls out funnel prerequisites and public exposure implications

#### Scenario: Mock backend behavior is documented as exposure-independent
- **WHEN** a user reads operating mode documentation for mock workflows
- **THEN** the documentation explains that `--mock` selects backend simulation only
- **AND** the documentation states that `tgate --mock` is tailnet-private by default
- **AND** the documentation includes explicit public mock guidance using `tgate --mock --funnel`
