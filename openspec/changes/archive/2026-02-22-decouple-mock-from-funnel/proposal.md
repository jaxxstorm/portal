## Why

`--mock` currently communicates two concerns at once: backend behavior and exposure posture. This is confusing and can unintentionally broaden exposure; mock mode should only define the backend response behavior, while tailnet vs Funnel should remain an explicit, separate operator choice.

## What Changes

- Change mock semantics so `--mock` (or equivalent config/env) selects a mock backend only, without implicitly enabling or preferring Funnel exposure.
- Preserve explicit exposure controls:
  - `tgate <port>` remains tailnet-private by default.
  - `tgate <port> --funnel` remains explicit public exposure.
  - `tgate --mock` remains tailnet-private by default unless Funnel is explicitly enabled.
  - `tgate --mock --funnel` explicitly opts into public mock exposure.
- Update startup and operator-facing messaging so output clearly distinguishes backend mode (`proxy` vs `mock`) from exposure mode (`tailnet` vs `funnel`).
- Update docs and examples so mock usage is presented as backend simulation independent of exposure.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `local-service-exposure`: redefine mock behavior so mock mode does not imply public reachability and Funnel remains explicit opt-in.
- `startup-operator-feedback`: require startup summaries to report backend mode and exposure mode as separate dimensions for mock runs.
- `documentation-system`: update operating-mode and mock/Funnel documentation to reflect explicit separation of backend and exposure semantics.

## Impact

- Affected code:
  - CLI/config resolution and runtime mode wiring in `main.go` and config helpers.
  - Exposure setup paths in local-daemon and tsnet startup flows.
  - Startup summary/log output fields for capability and mode reporting.
  - Documentation under `docs/` and `README.md` examples that describe mock and Funnel behavior.
- Behavioral impact:
  - Operators must explicitly set Funnel for public mock endpoints.
  - Default mock behavior becomes safer and more predictable (tailnet-private unless explicitly public).
