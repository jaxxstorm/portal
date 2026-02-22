## Purpose

Define responsive TUI layout and endpoint-rendering behavior so key operator
information remains readable across compact and large terminal sizes.

## Requirements

### Requirement: TUI Layout MUST Adapt To Terminal Size
`portal` MUST select a deterministic TUI layout profile from terminal dimensions so critical operator information remains visible on compact terminals and uses space effectively on large monitors.

#### Scenario: Compact terminal uses a readability-first layout
- **WHEN** a user runs `portal <port>` with TUI enabled in a compact terminal window
- **THEN** portal renders a compact profile that prioritizes endpoint visibility and recent logs without pane overlap
- **AND** pane content remains readable without requiring horizontal scrolling for core status labels

#### Scenario: Large monitor uses an expanded multi-pane layout
- **WHEN** a user runs `portal <port>` with TUI enabled on a large terminal window
- **THEN** portal renders an expanded profile with dedicated space for endpoint/status details, request details, and logs
- **AND** the additional width/height is used to increase visible request and log context

#### Scenario: Runtime resize recalculates layout without losing state
- **WHEN** a user resizes the terminal while portal is running in TUI mode
- **THEN** portal recalculates layout profile and pane sizes deterministically
- **AND** previously collected request and log data remain available after reflow

### Requirement: Endpoint URLs MUST Be Rendered As Dedicated TUI Fields
`portal` MUST render service and Web UI endpoint information in dedicated labeled fields in TUI mode, and URL formatting MUST preserve scheme and host when truncation or wrapping is required.

#### Scenario: Tailnet mode shows private service and Web UI endpoints
- **WHEN** a user runs `portal <port>` and startup succeeds with UI enabled
- **THEN** the TUI endpoint section shows the tailnet service URL in a labeled field
- **AND** the TUI endpoint section shows Web UI status and URL in a labeled field when available

#### Scenario: Funnel mode shows public exposure endpoint context
- **WHEN** a user runs `portal <port> --funnel` and Funnel prerequisites are satisfied
- **THEN** the TUI endpoint section shows that service exposure is public via Funnel
- **AND** the service URL remains clearly legible in the endpoint section even for long URLs

#### Scenario: Tsnet mode shows unavailable Web UI state explicitly
- **WHEN** portal runs in tsnet mode where Web UI exposure is unavailable
- **THEN** the TUI endpoint section shows Web UI status as unavailable
- **AND** the unavailable state includes an explicit reason instead of showing a blank URL field

### Requirement: TUI Endpoint Rendering MUST Keep Exposure Semantics Explicit
`portal` MUST present exposure semantics in the endpoint section so tailnet-private default behavior and Funnel opt-in public behavior are unambiguous.

#### Scenario: Tailnet mode remains the default private baseline
- **WHEN** a user runs `portal <port>` without `--funnel`
- **THEN** the TUI endpoint section identifies exposure as tailnet-private
- **AND** the rendered state does not imply public internet reachability

#### Scenario: Funnel mode is explicit opt-in public exposure
- **WHEN** a user runs `portal <port> --funnel`
- **THEN** the TUI endpoint section identifies exposure as Funnel/public
- **AND** portal surfaces that public exposure depends on successful Funnel/HTTPS setup
