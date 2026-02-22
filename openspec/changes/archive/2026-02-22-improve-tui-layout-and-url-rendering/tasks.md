## 1. Endpoint Summary UX

- [x] 1.1 Add structured endpoint state to the TUI data surface (service URL, exposure, web UI status, web UI reason) instead of relying on ad-hoc log text.
- [x] 1.2 Implement a dedicated endpoint summary section in `internal/tui/model.go` with stable labels for service URL and Web UI status/URL.
- [x] 1.3 Add URL rendering helpers that preserve scheme/host and apply deterministic wrapping/truncation for long endpoints.
- [x] 1.4 Ensure endpoint summary explicitly distinguishes tailnet-private default behavior from Funnel opt-in public behavior.

## 2. Adaptive Layout Profiles

- [x] 2.1 Refactor TUI size calculation into deterministic compact, standard, and wide layout profiles with clear width/height thresholds.
- [x] 2.2 Update pane composition so compact layouts prioritize endpoint summary and logs without pane overlap or unreadable clipping.
- [x] 2.3 Update standard/wide layouts to use extra terminal space for request context and logs while keeping endpoint details visible.
- [x] 2.4 Ensure window resize events recompute profile and pane sizes while preserving request/log history in viewports.

## 3. Startup Summary Consistency

- [x] 3.1 Wire startup readiness data from local-daemon and tsnet setup paths into endpoint summary state so TUI and startup output stay aligned.
- [x] 3.2 Ensure tsnet runs render explicit Web UI unavailable status with reason instead of blank endpoint fields.
- [x] 3.3 Ensure Funnel/HTTPS prerequisite failures do not render ready-like endpoint output and continue to surface clear startup failure messages.

## 4. Validation

- [x] 4.1 Add or update TUI model tests for compact/standard/wide profile selection and deterministic behavior across resize transitions.
- [x] 4.2 Add tests for tailnet mode (`portal <port>`) verifying readable service URL and Web UI endpoint rendering in TUI output.
- [x] 4.3 Add tests for Funnel mode (`portal <port> --funnel`) verifying public exposure labeling and long-URL rendering behavior.
- [x] 4.4 Add tests for tsnet/unavailable-UI and Funnel-prerequisite-failure cases verifying explicit unavailable/failure status rendering.
