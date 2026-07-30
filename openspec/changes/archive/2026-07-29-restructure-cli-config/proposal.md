## Why

Portal's single command accumulates serving, authentication, tsnet, cleanup,
and bootstrap flags in one place, making common workflows difficult to discover
and safely repeat. A command-oriented interface and a practical configuration
file workflow make routine operation concise while preserving explicit Funnel
opt-in.

## What Changes

- Introduce focused CLI commands for serving, OAuth App bootstrap, and Tailscale
  Serve cleanup instead of requiring unrelated flags on the default command.
- Make `portal serve` the documented serving entry point while retaining existing
  `portal <port>` and `portal --mock` invocation forms as compatibility aliases.
- Provide a generated, documented configuration-file template and command-aware
  configuration validation so users can keep recurring settings out of shell
  history.
- Keep `portal serve <port>` tailnet-private by default and require explicit
  `--funnel` or configuration-file opt-in for public exposure.

## Capabilities

### New Capabilities

- `command-oriented-cli`: Discoverable commands and scoped help for serving,
  OAuth App bootstrap, and maintenance actions.

### Modified Capabilities

- `runtime-configuration`: Resolve serving configuration through the new command
  surface and support generating a configuration-file template.
- `local-service-exposure`: Preserve tailnet-private and explicit Funnel behavior
  when serving through the new command syntax.
- `documentation-system`: Document command-oriented workflows and configuration
  templates.

## Impact

- Affects Cobra command construction, configuration parsing and validation,
  `main` command dispatch, CLI tests, documentation, and existing automation.
- Existing invocation forms remain supported; public Funnel exposure remains
  explicit and subject to HTTPS/certificate prerequisites.
