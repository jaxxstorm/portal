## Context

`tgate` currently treats mock runs as a special UX path, and operator
expectations have drifted toward "mock means public webhook mode." That blends
two independent concerns:
- backend behavior (`proxy` vs `mock`)
- exposure posture (`tailnet` vs `funnel`)

The product baseline is tailnet-private by default with explicit Funnel opt-in.
This design restores that baseline for mock runs so local-daemon and tsnet
paths evaluate exposure consistently from explicit Funnel configuration only.

## Goals / Non-Goals

**Goals:**
- Make mock mode select backend behavior only, never implicit exposure.
- Keep exposure resolution source-agnostic: CLI, env, and config all follow the
  same default tailnet + explicit Funnel semantics.
- Preserve existing Funnel setup/error handling in both local-daemon and tsnet
  startup paths.
- Improve startup/operator feedback by reporting backend mode and exposure mode
  as separate dimensions.

**Non-Goals:**
- Changing Funnel prerequisites, certificate issuance, or HTTPS requirements.
- Reworking tsnet listen-mode/service-mode compatibility rules.
- Replacing mock response format or request-capture behavior.

## Decisions

### Decision: Model backend mode and exposure mode as independent runtime dimensions
- Keep `mock` as a backend selector (`proxy` vs `mock`).
- Keep `funnel` as exposure selector (`tailnet` vs `funnel`).
- Resolve both through existing precedence rules (CLI > env > config > defaults).

Why:
- This aligns behavior with security defaults and avoids accidental public
  exposure.
- It keeps mode resolution simpler and easier to explain in docs.

Alternatives considered:
- Continue implicit public mock behavior: rejected due to unclear security
  posture and surprising behavior.
- Add a new "public-mock" flag: rejected because it duplicates existing Funnel
  controls and increases configuration complexity.

### Decision: Apply the same exposure evaluation in local-daemon and tsnet flows
- Local-daemon path: `SetupServe` uses explicit Funnel configuration only.
- tsnet path: server startup uses explicit Funnel configuration only.
- In both cases, mock backend can run with either tailnet (default) or Funnel
  (explicit opt-in).

Why:
- Ensures mode parity and reduces implementation drift between serving paths.

Alternatives considered:
- Divergent behavior by serving mode: rejected due to operator confusion and
  inconsistent docs/test expectations.

### Decision: Keep Funnel failure behavior unchanged and explicit
- If Funnel is enabled (including `--mock --funnel`) and prerequisites fail
  (HTTPS/cert/setup), startup fails before ready summary.
- If Funnel is not enabled, mock runs remain tailnet-private and do not execute
  Funnel setup paths.

Why:
- Preserves predictable failure semantics and existing safety controls.

Alternatives considered:
- Silent fallback from Funnel to tailnet on Funnel failure: rejected because it
  hides operator intent and can mask security/reachability issues.

### Decision: Extend startup summary semantics for mode clarity
- Startup-ready output includes both backend mode and exposure mode, so mock
  tailnet vs mock funnel is unambiguous in TUI, text logs, and JSON logs.

Why:
- Prevents ambiguity that currently leads operators to infer wrong exposure.

Alternatives considered:
- Keep existing capability-only summary: rejected because it does not clearly
  distinguish backend vs exposure intent during mock runs.

## Risks / Trade-offs

- [Risk] Existing users expecting public mock by default may perceive regression
  → Mitigation: update docs/examples and release notes with explicit
  `--mock --funnel` guidance.
- [Risk] Startup summary field additions could impact log consumers
  → Mitigation: preserve existing keys and add stable new keys rather than
  renaming/removing old ones.
- [Trade-off] More explicit configuration required for public mock testing
  → Mitigation: this is intentional for safer defaults and clearer operator
  intent.

## Migration Plan

1. Update runtime mode resolution so mock does not influence Funnel.
2. Update startup summary/event payload to include backend+exposure separation.
3. Update docs (`docs/operating-modes.md`, `docs/configuration.md`, README
   examples) to show `--mock` vs `--mock --funnel`.
4. Add/adjust tests for mode resolution and startup summary semantics.

Rollback:
- Revert mode-resolution and startup-summary changes in a single release if
  unexpected compatibility issues are discovered.

## Open Questions

- Should we add an explicit warning when `--mock` is used without Funnel to
  remind operators that the endpoint is tailnet-private?
