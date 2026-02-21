## 1. Runtime Configuration Capability

- [x] 1.1 Define the canonical config schema and defaults covering existing runtime options (port, funnel, mock, tsnet/local tailscale, logging, serve settings).
- [x] 1.2 Implement Viper loading from `~/.tgate/config.yml` into the canonical config model.
- [x] 1.3 Implement `TGATE_*` environment variable bindings for every supported configuration key with consistent key normalization.
- [x] 1.4 Implement and test precedence resolution (CLI > env > config file > defaults), including conflict cases.

## 2. CLI Compatibility Capability

- [x] 2.1 Replace Kong parsing with Cobra commands/flags while preserving existing command usage (`tgate <port>`, `tgate <port> --funnel`, `tgate --mock`).
- [x] 2.2 Route resolved configuration into existing startup paths without changing tailnet-default or funnel-opt-in semantics.
- [x] 2.3 Preserve and verify HTTPS certificate and funnel prerequisite failure behavior when funnel is enabled via CLI, config file, or env var.

## 3. Exposure Behavior Validation

- [x] 3.1 Add/refresh tests for tailnet-only exposure when configured by CLI, config file, and `TGATE_*` environment variables.
- [x] 3.2 Add/refresh tests for funnel-public exposure when configured by CLI, config file, and `TGATE_*` environment variables.
- [x] 3.3 Add compatibility regression tests to confirm current CLI scripts and docs examples continue to work unchanged.

## 4. Documentation And Rollout

- [x] 4.1 Update README and usage docs to describe `~/.tgate/config.yml`, `TGATE_*` mappings, and precedence rules.
- [x] 4.2 Add Docker/sidecar-focused examples showing env-only and mounted-config operation.
- [x] 4.3 Document migration guidance from Kong-based startup flags to the Cobra/Viper-backed configuration model.
