## 1. Authorization Configuration

- [x] 1.1 Add authorization mode, username, and password fields to runtime configuration with CLI, environment, and YAML bindings.
- [x] 1.2 Validate disabled, HTTP Basic, incomplete, and unsupported authorization configurations.
- [x] 1.3 Add configuration tests for default, config-file, environment, CLI precedence, and validation behavior.

## 2. Request Authorization

- [x] 2.1 Define a provider-neutral request authorizer interface and HTTP Basic implementation with constant-time credential comparison.
- [x] 2.2 Wire the configured authorizer into proxy construction and enforce it after Funnel allowlist checks and before backend dispatch.
- [x] 2.3 Add proxy tests for disabled authorization, valid HTTP Basic credentials, rejected credentials, mock dispatch, and Funnel authorization behavior.

## 3. Documentation And Validation

- [x] 3.1 Document authorization configuration, HTTP Basic security guidance, tailnet usage, and Funnel HTTPS prerequisites.
- [x] 3.2 Run formatting and the relevant Go test suite for tailnet and Funnel authorization paths.
