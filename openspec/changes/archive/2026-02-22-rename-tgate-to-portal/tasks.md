## 1. Runtime Identity And Configuration Namespace

- [x] 1.1 Rename CLI/binary-facing project references from `tgate` to `portal` in command usage/help output and user-visible naming defaults.
- [x] 1.2 Update configuration file default location from `~/.tgate/config.yml` to `~/.portal/config.yml`.
- [x] 1.3 Rename environment-variable namespace from `TGATE_*` to `PORTAL_*` across config parsing and related tests.
- [x] 1.4 Update default identity values tied to product name (for example device-name/service-name defaults) from `tgate` to `portal`.

## 2. Exposure Behavior Surface And Startup Output

- [x] 2.1 Update startup/log output fields and messages to use portal naming while preserving existing tailnet-default and funnel-opt-in behavior.
- [x] 2.2 Update command/example references in runtime behavior paths so tailnet and Funnel scenarios use `portal <port>` and `portal <port> --funnel`.
- [x] 2.3 Ensure startup-ready output and related tests remain consistent after rename for proxy and mock backends.

## 3. Documentation And Project Metadata

- [x] 3.1 Rename README and docs references from `tgate` to `portal`, including all command examples.
- [x] 3.2 Update configuration docs to reflect `~/.portal/config.yml` and `PORTAL_*` variables.
- [x] 3.3 Update contributor/build/release metadata references (where present) to emit and describe `portal` naming.

## 4. Verification And Cleanup

- [x] 4.1 Run focused tests for config parsing, startup summary/logging, and renamed defaults.
- [x] 4.2 Validate tailnet-only exposure behavior still works with renamed command/config surfaces.
- [x] 4.3 Validate Funnel-public behavior still works with renamed command/config surfaces.
- [x] 4.4 Perform repository-wide audit to identify and resolve remaining maintained `tgate` references.
