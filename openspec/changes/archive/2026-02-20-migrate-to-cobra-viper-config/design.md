## Context

tgate currently uses `alecthomas/kong` for argument parsing and relies primarily on CLI flags/args for runtime configuration. This is workable for interactive local usage, but it is awkward in Docker and sidecar environments where immutable images, environment injection, and mounted configuration files are the norm.

The change must keep current operator-facing behavior for service exposure:
- Tailnet-private exposure remains the default.
- Funnel remains explicit opt-in with HTTPS and certificate prerequisites.
- Existing command usage such as `tgate 8080` and `tgate 8080 --funnel` must continue to work.

Both local Tailscale mode and tsnet mode must continue to satisfy existing exposure requirements, with configuration source changes limited to how startup options are provided.

## Goals / Non-Goals

**Goals:**
- Replace Kong with Cobra/Viper while preserving current CLI surface.
- Support config file loading from `~/.tgate/config.yml`.
- Support `TGATE_*` environment variables for every configurable runtime value.
- Define deterministic precedence across CLI args, env vars, config file, and defaults.
- Preserve tailnet and funnel behavior semantics across local Tailscale and tsnet operation.

**Non-Goals:**
- Changing core proxy behavior, request logging, or TUI functionality.
- Changing default exposure model (tailnet first, funnel opt-in).
- Introducing a broader plugin/config format redesign beyond immediate runtime settings.

## Decisions

1. **CLI framework migration to Cobra**
   - Replace Kong parsing with a Cobra root command that keeps current positional/flag UX intact.
   - Preserve command compatibility (`tgate <port>`, `tgate --mock`, `tgate <port> --funnel`) to avoid breaking scripts and docs.
   - Rationale: Cobra is a stable CLI foundation and integrates directly with Viper bindings.

2. **Centralized configuration assembly with Viper**
   - Add a single configuration loader that resolves values from:
     - defaults
     - `~/.tgate/config.yml`
     - `TGATE_*` environment variables
     - CLI args/flags
   - Precedence order: CLI > environment > config file > defaults.
   - Rationale: predictable 12-factor semantics while preserving explicit CLI overrides.

3. **Canonical key mapping for env vars**
   - Use `TGATE_` prefix and normalize keys so config keys map consistently (for example `funnel` -> `TGATE_FUNNEL`, `serve-port` -> `TGATE_SERVE_PORT`).
   - Ensure all existing runtime options can be set by environment variable, including mode toggles relevant to Docker/sidecar execution.
   - Rationale: complete env parity is required for non-interactive deployment targets.

4. **Typed configuration + validation boundary**
   - Keep an internal typed config struct with validation after merge to preserve existing safety checks.
   - Validation continues to enforce funnel prerequisites (HTTPS + certificate readiness path) before runtime setup.
   - Rationale: source flexibility should not weaken configuration correctness or startup guarantees.

5. **Mode behavior remains unchanged across local Tailscale and tsnet**
   - Local Tailscale and tsnet startup paths consume the same resolved config object.
   - Tailnet default and funnel opt-in semantics remain identical regardless of whether values came from CLI, env, or file.
   - Rationale: configuration transport should not alter networking/security behavior.

6. **Failure handling for HTTPS certificate and Funnel setup remains explicit**
   - If funnel is enabled without valid HTTPS/certificate prerequisites, startup fails with actionable errors.
   - Errors are surfaced consistently no matter whether `funnel` was set by CLI flag, config file, or `TGATE_FUNNEL`.
   - Rationale: preserve current security constraints and reduce misconfiguration ambiguity.

## Risks / Trade-offs

- [Risk] CLI compatibility drift during parser migration -> Mitigation: add compatibility tests for existing command examples and flags.
- [Risk] Ambiguous source precedence surprises users -> Mitigation: document and test precedence (CLI > env > file > defaults) with explicit cases.
- [Risk] Partial env key mapping leaves options unconfigurable in containers -> Mitigation: table-driven tests that assert every config field has a `TGATE_*` mapping.
- [Risk] Funnel setup failures become harder to debug when enabled indirectly -> Mitigation: include source-aware startup logging and preserve actionable certificate error messages.
- [Trade-off] Added dependency surface (Cobra + Viper) increases startup complexity -> Mitigation: isolate configuration loading into one package and keep runtime wiring unchanged.

## Migration Plan

1. Introduce Cobra command wiring with equivalent args/flags and keep existing runtime entry points.
2. Implement Viper-backed config loader for file + env + CLI merge.
3. Migrate existing config validation into the new loader boundary.
4. Add compatibility tests for tailnet default and funnel opt-in behavior across all config sources.
5. Update docs with `~/.tgate/config.yml` and `TGATE_*` examples for Docker/sidecar usage.

## Open Questions

- Should we support an explicit `--config` path override now, or defer to a follow-up change?
- Do we want startup output to display the effective config source per key for easier debugging in CI/container environments?
