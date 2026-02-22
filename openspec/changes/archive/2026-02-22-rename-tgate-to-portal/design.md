## Context

The codebase and user-facing interfaces currently use `tgate` as the product,
binary, configuration namespace, and documentation identity. The requested
change is a full rename to `portal`, including command usage, config/env
surfaces, and implementation references. This is a cross-cutting change that
touches CLI behavior, runtime configuration, startup output, docs, tests, and
release/build metadata.

## Goals / Non-Goals

**Goals:**
- Rename all maintained project references from `tgate` to `portal`.
- Ensure runtime behavior and security semantics remain unchanged after rename
  (tailnet default, Funnel opt-in, same prerequisite handling).
- Migrate user-facing configuration namespace:
  - `~/.tgate/config.yml` -> `~/.portal/config.yml`
  - `TGATE_*` -> `PORTAL_*`
- Update startup and documentation examples to use `portal` consistently.
- Keep the implementation verifiable with focused regression tests.

**Non-Goals:**
- Changing exposure semantics, networking behavior, or policy defaults.
- Introducing new runtime capabilities unrelated to naming.
- Supporting long-term dual-brand output in docs/logs.

## Decisions

### Decision: Treat the rename as a breaking identity migration
- Command, docs, config path, and env prefixes move to `portal`.
- Existing `tgate` references are removed from maintained user-facing surfaces.

Why:
- The user requested full replacement of references.
- Partial aliasing creates prolonged ambiguity and larger future cleanup scope.

Alternatives considered:
- Keep `tgate` aliases indefinitely: rejected to avoid dual-identity drift.
- Introduce phased migration windows in this change: rejected to keep scope
  focused on complete rename.

### Decision: Preserve behavior while renaming only identity and namespace
- Runtime mode resolution, exposure selection, allowlist enforcement, and tsnet
  validation remain semantically unchanged.
- Only names, prefixes, defaults, and examples are updated.

Why:
- Reduces risk by separating brand/identity migration from behavior changes.

Alternatives considered:
- Combine rename with functional refactors: rejected due to higher regression
  risk and harder rollback.

### Decision: Update defaults and metadata tied to product identity
- Default device/service naming values move from `tgate` to `portal`.
- Logging component labels and startup examples move to portal naming.
- Release/build/package references are updated to emit `portal` artifacts.

Why:
- Avoids hidden residual references that contradict user-visible branding.

Alternatives considered:
- Leave internal defaults/components unchanged: rejected as incomplete rename.

## Risks / Trade-offs

- [Risk] User scripts and automation break on command/env/path rename
  → Mitigation: document breaking changes and migration examples clearly.
- [Risk] Hidden references remain in less-visible files/tests
  → Mitigation: repository-wide audit and dedicated verification tasks.
- [Trade-off] No compatibility aliases increases migration burden
  → Mitigation: provide explicit migration mapping in docs/changelog notes.

## Migration Plan

1. Rename user-facing command/docs/config/env surfaces to `portal`.
2. Update defaults and internal naming references tied to project identity.
3. Update tests and fixtures to match new naming.
4. Run full test suite plus targeted repository search validation for lingering
   `tgate` references.

Rollback:
- Revert this change set to restore prior naming if critical downstream
  compatibility issues are discovered.

## Open Questions

- Should any short-lived compatibility alias (`tgate` command shim or
  `TGATE_*` fallback) be added in a follow-up change, or should migration remain
  strict and immediate?
