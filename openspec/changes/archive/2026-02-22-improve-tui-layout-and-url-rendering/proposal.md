## Why

The current TUI uses a mostly fixed layout split that does not adapt well to very small or very large terminals, making key information harder to scan. The Web UI URL is shown as plain text in a crowded stats pane, so operators can miss where to open the dashboard during startup and live operation.

## What Changes

- Introduce an adaptive TUI layout that reflows panes for compact terminals and uses available space more effectively on large monitors.
- Improve endpoint presentation so service and Web UI URLs are rendered in a dedicated, high-signal section with predictable wrapping/truncation behavior.
- Ensure `portal <port>` presents the tailnet service URL and (when available) Web UI URL clearly in the TUI.
- Ensure `portal <port> --funnel` presents both public exposure context and URL information clearly, while keeping Funnel as explicit opt-in behavior.
- Preserve existing startup security semantics: tailnet remains the default baseline; Funnel output continues to surface prerequisite and public-exposure implications.

## Capabilities

### New Capabilities
- `tui-adaptive-layout`: Define responsive TUI layout behavior and URL rendering requirements for small and large terminal sizes.

### Modified Capabilities
- `startup-operator-feedback`: Tighten TUI startup summary presentation requirements so Web UI and service URLs remain legible and discoverable within TUI-managed output.

## Impact

- Affected code:
  - `internal/tui/model.go` (layout computation, pane composition, URL rendering)
  - `main.go` (startup TUI output wiring where needed)
  - `internal/tui/logging.go` (if endpoint emphasis requires structured message formatting updates)
- Affected tests:
  - New/expanded tests for TUI layout behavior across terminal sizes and URL rendering edge cases.
- No external API changes; behavior changes are in operator-facing terminal UX and startup observability output.
