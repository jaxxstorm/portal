## Context

Portal currently parses one root command with a positional port and a broad set
of flags. It already reads `~/.portal/config.yml`, but users must know the
complete key set and have no supported way to generate a safe starter file.
OAuth App bootstrap is separately dispatched before normal configuration parsing,
which makes command discovery inconsistent.

## Goals / Non-Goals

**Goals:**
- Provide clear `serve`, `mock`, `oauth-app`, `cleanup`, and `config` command
  entry points with scoped help.
- Add `config init` to generate an editable starter configuration file without
  overwriting an existing file by default.
- Preserve existing serving behavior and use the same resolved configuration for
  local Tailscale and tsnet modes.
- Keep tailnet-private exposure as the default and require an explicit Funnel
  setting in the command or configuration file.

**Non-Goals:**
- Remove existing `portal <port>`, `portal --mock`, or `--cleanup-serve`
  invocations in this change.
- Add interactive secret storage, credential discovery, or a configuration
  editor.
- Change configuration precedence, Funnel prerequisites, or authentication
  policy.

## Decisions

### Use a single Cobra command tree

Refactor command construction so `main` executes one root Cobra tree rather
than dispatching special commands before normal configuration parsing. The tree
will expose `serve [port]`, `mock`, `oauth-app`, `cleanup`, and `config init`.
Serving flags remain attached only to serving commands; OAuth bootstrap and
cleanup commands do not parse or validate irrelevant serve settings.

The root's existing positional-port and `--mock` behavior remains a compatibility
path that resolves to the same handlers as `serve` and `mock`. This avoids
breaking scripts and documented invocations while directing new users toward
the scoped commands.

### Generate but never implicitly overwrite configuration

`portal config init` writes a commented YAML template to
`~/.portal/config.yml`, creating the parent directory with user-only
permissions. It exits with a clear error when that file already exists unless
the operator explicitly passes `--force`. The template documents common
tailnet-private proxy, mock, Funnel, tsnet, and authorization settings without
including generated secrets.

This provides a discoverable configuration workflow without changing the
existing source precedence: command flags, then `PORTAL_*` environment values,
then file values, then defaults.

### Keep public exposure command-neutral

`portal serve <port>` and `portal mock` are tailnet-private unless their
effective configuration explicitly enables Funnel. `--funnel` remains visible
on the relevant serving commands, and generated configuration defaults it to
false. Funnel startup continues to require HTTPS/certificate readiness and
reports the existing public-readiness fields.

## Risks / Trade-offs

- [Compatibility parser ambiguity] -> Reserve known subcommand names and keep
  numeric positional arguments on the legacy root path.
- [Template files become a secret store] -> Generate only placeholder values,
  use user-only file permissions, and document environment/secret-manager use.
- [Users mistake `mock` for public exposure] -> Keep mock and Funnel independent
  in command help, generated configuration comments, and documentation.
- [Duplicate configuration behavior] -> Route every serving command through one
  configuration-resolution and validation path.

## Migration Plan

1. Build the unified root command tree and adapt the existing compatibility
   parser to use it.
2. Add command-specific configuration resolution, `config init`, and tests for
   precedence and overwrite behavior.
3. Move OAuth App and cleanup dispatch into the tree, preserving their current
   behavior.
4. Update CLI and configuration documentation with command examples.

Rollback is retaining the existing compatibility paths and reverting the new
commands; existing configuration files remain valid.

## Open Questions

None.
