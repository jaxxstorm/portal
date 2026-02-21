## 1. Startup summary capability

- [x] 1.1 Add a canonical startup summary model that captures mode, exposure, service URL, web UI status/URL, enabled capabilities, and readiness state.
- [x] 1.2 Populate the summary from local-daemon startup flow (`tailscaled` path) once serving is ready.
- [x] 1.3 Populate the summary from tsnet startup flow once serving is ready.
- [x] 1.4 Ensure startup failure paths (including HTTPS/Funnel prerequisite failures) do not emit a ready summary.

## 2. Mode-specific startup rendering

- [x] 2.1 Render startup summary in TUI-managed output so startup completion and endpoints are visible in the UI channel.
- [x] 2.2 Render startup summary in non-TUI text logging as a concise startup-complete statement with service and UI reachability.
- [x] 2.3 Render startup summary in non-TUI JSON logging as a structured event with stable keys.
- [x] 2.4 Remove or refactor conflicting legacy startup prints so there is one definitive startup-ready output path.

## 3. Validation and documentation

- [x] 3.1 Add tests for tailnet mode (`tgate <port>`) verifying startup summary fields and web UI location reporting.
- [x] 3.2 Add tests for funnel mode (`tgate <port> --funnel`) verifying startup summary fields, public exposure signaling, and failure behavior when prerequisites are missing.
- [x] 3.3 Add tests for non-TUI JSON mode to verify startup output remains structured and machine-readable.
- [x] 3.4 Update docs to show startup output expectations for tailnet and Funnel usage, including auto-assigned `--ui-port` behavior.
