## 1. TSNet service-mode capability

- [x] 1.1 Add runtime configuration keys for tsnet listen mode and service name across CLI, `TGATE_*` env vars, and `~/.tgate/config.yml`.
- [x] 1.2 Extend tsnet server setup to select listener mode (`Listen`/`ListenTLS`/`ListenFunnel`) or service mode (`ListenService`) from resolved config.
- [x] 1.3 Implement service-mode serving path that proxies requests to the same backend handler used by existing tsnet mode.
- [x] 1.4 Populate tsnet startup URL details for service mode using returned service FQDN/address information.

## 2. Compatibility, validation, and startup output

- [x] 2.1 Validate tsnet service names before startup and fail with clear configuration errors on invalid values.
- [x] 2.2 Handle Funnel + service-mode configuration deterministically by failing validation with a clear incompatibility error.
- [x] 2.3 Surface runtime `ListenService` prerequisite failures (for example untagged host or approval issues) with actionable startup errors.
- [x] 2.4 Extend startup summary/log output to include configured/effective tsnet listen mode and service identity when applicable.

## 3. Verification and documentation

- [x] 3.1 Add configuration tests verifying precedence and resolution for tsnet listen mode/service-name across CLI, env, and config file.
- [x] 3.2 Add tsnet serving tests for tailnet-private service mode success path.
- [x] 3.3 Add tsnet serving/config tests for Funnel-enabled runs with service mode configured, verifying incompatibility failure behavior and startup signaling.
- [x] 3.4 Update operating modes and configuration docs with tsnet service-mode setup, prerequisites, and Funnel interaction behavior.
