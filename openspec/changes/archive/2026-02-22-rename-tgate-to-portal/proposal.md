## Why

The project name, command examples, and configuration namespace still use
`tgate`, while product direction now requires a unified `portal` identity.
Renaming all user-visible and implementation references now avoids long-term
brand inconsistency and migration churn.

## What Changes

- Rename the project identity from `tgate` to `portal` across code,
  documentation, CLI examples, logging/component labels, and packaging metadata.
- Rename the CLI command and user-facing command examples from `tgate` to
  `portal`.
- Rename configuration surfaces tied to project identity:
  - **BREAKING** `~/.tgate/config.yml` -> `~/.portal/config.yml`
  - **BREAKING** `TGATE_*` environment variables -> `PORTAL_*`
- Rename default runtime identity values where they are currently `tgate` (for
  example default node/device name and service naming defaults).
- Update tests, build/release references, and docs to ensure no remaining
  `tgate` references in maintained project assets.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `runtime-configuration`: rename command/config/env/default naming references
  to `portal`, including new config path and environment-variable namespace.
- `local-service-exposure`: update normative command examples and product naming
  to `portal` for tailnet/funnel behavior.
- `startup-operator-feedback`: update startup summary expectations to use portal
  identity in user-visible output.
- `documentation-system`: require docs and contributor-facing guidance to use
  `portal` naming consistently.

## Impact

- Affected code:
  - CLI/config parsing, defaults, and environment binding.
  - Startup/logging strings and component identifiers.
  - Docs, README, and operating/configuration examples.
  - Tests asserting names, command usage, env keys, and default values.
  - Build/release metadata and packaging references using project/binary name.
- External impact:
  - Existing users must migrate config path and env variable names.
  - Automation/scripts that invoke `tgate` will need to move to `portal`.
