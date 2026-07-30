## 1. Command Interface

- [x] 1.1 Build one root Cobra command tree with `serve`, `mock`, `oauth-app`, `cleanup`, and `config` commands.
- [x] 1.2 Move serving configuration parsing and validation behind the `serve` and `mock` command handlers.
- [x] 1.3 Preserve `portal <port>`, `portal --mock`, and `portal --cleanup-serve` compatibility paths through the new handlers.
- [x] 1.4 Add command help and parsing tests for scoped flags, legacy aliases, tailnet-private defaults, and explicit Funnel opt-in.

## 2. Configuration Workflow

- [x] 2.1 Add `portal config init` to write a commented, tailnet-private YAML template at the default configuration path.
- [x] 2.2 Protect existing configuration from overwrite unless `--force` is set, use user-only permissions, and exclude secret values from the template.
- [x] 2.3 Add configuration-template tests for creation, overwrite protection, forced replacement, and command-aware validation.

## 3. Command Migration

- [x] 3.1 Move OAuth App bootstrap into the unified root command tree without changing its management OAuth credential behavior.
- [x] 3.2 Move Tailscale Serve cleanup into `portal cleanup` while retaining the legacy cleanup flag.
- [x] 3.3 Add end-to-end command dispatch tests for serve, mock, OAuth App, cleanup, and config commands.

## 4. Documentation And Verification

- [x] 4.1 Update README and docs with command-oriented serve, mock, cleanup, OAuth App, and configuration-template examples.
- [x] 4.2 Document configuration precedence, secret-handling guidance, tailnet-private defaults, and explicit Funnel prerequisites.
- [x] 4.3 Run the Go test suite and OpenSpec validation.
