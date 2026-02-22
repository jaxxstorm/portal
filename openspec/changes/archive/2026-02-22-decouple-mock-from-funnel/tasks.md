## 1. Mock Mode Semantics

- [x] 1.1 Update runtime mode resolution so `mock` selects backend behavior only and does not implicitly enable or prefer Funnel.
- [x] 1.2 Ensure CLI/env/config precedence continues to resolve `mock` and `funnel` independently.
- [x] 1.3 Add/adjust config and mode-resolution unit tests for `--mock`, `--funnel`, and `--mock --funnel` combinations.

## 2. Exposure Wiring (Local Daemon + tsnet)

- [x] 2.1 Update local-daemon startup flow so exposure selection is driven only by explicit Funnel configuration for both proxy and mock backends.
- [x] 2.2 Update tsnet startup flow so mock backend remains tailnet-private by default and only uses public exposure when Funnel is explicitly enabled.
- [x] 2.3 Preserve and validate existing Funnel prerequisite/certificate failure handling for mock+funnel startup paths.

## 3. Startup Operator Feedback

- [x] 3.1 Extend startup-ready summary payloads to include separate backend mode and exposure mode fields.
- [x] 3.2 Ensure TUI, text logs, and JSON logs render the same backend/exposure semantics for mock and proxy runs.
- [x] 3.3 Add/update startup summary tests to verify no ready summary is emitted when Funnel startup fails.

## 4. Documentation Updates

- [x] 4.1 Update `docs/operating-modes.md` to describe `--mock` as backend simulation and `--mock --funnel` as explicit public mock exposure.
- [x] 4.2 Update `docs/configuration.md` and related examples to show independent mock and Funnel controls across CLI/env/config.
- [x] 4.3 Update `README.md` quick examples to reflect tailnet-default mock behavior and explicit Funnel opt-in.

## 5. Validation

- [x] 5.1 Verify tailnet-only behavior for `tgate --mock` (no public endpoint configured).
- [x] 5.2 Verify Funnel-public behavior for `tgate --mock --funnel` when prerequisites are satisfied.
- [x] 5.3 Run focused test suites for mode resolution, startup summaries, and serving-path behavior in both local-daemon and tsnet modes.
