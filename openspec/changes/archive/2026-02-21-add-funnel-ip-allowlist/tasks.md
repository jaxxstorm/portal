## 1. Funnel allowlist configuration surface

- [x] 1.1 Add config fields for Funnel allowlist entries (IP/CIDR list) and wire them into the parsed runtime config model.
- [x] 1.2 Bind and normalize `TGATE_FUNNEL_ALLOWLIST` so comma-separated values map to the same runtime field as `~/.tgate/config.yml`.
- [x] 1.3 Validate Funnel allowlist entries at startup and return clear configuration errors for invalid IP/CIDR values.
- [x] 1.4 Add/extend config tests for file loading, env loading, and precedence (`TGATE_FUNNEL_ALLOWLIST` over config file).

## 2. Funnel request enforcement

- [x] 2.1 Implement request-origin resolver for Funnel requests using trusted request metadata and remote address fallback.
- [x] 2.2 Add Funnel-only allowlist middleware that returns HTTP 403 when source IP is not allowlisted.
- [x] 2.3 Enforce fail-closed behavior (HTTP 403) when allowlist is enabled and no trustworthy source IP can be resolved.
- [x] 2.4 Add structured logging for allow/deny outcomes, including source-signal type and reason for denial.

## 3. Behavior validation and documentation

- [x] 3.1 Add tests proving tailnet-private mode (`tgate <port>`) is unchanged by Funnel allowlist configuration.
- [x] 3.2 Add tests for Funnel mode (`tgate <port> --funnel`) covering allowlisted, non-allowlisted, and unresolved-source requests.
- [x] 3.3 Update `docs/configuration.md` with Funnel allowlist keys, `TGATE_FUNNEL_ALLOWLIST`, format examples, and precedence behavior.
- [x] 3.4 Update Funnel docs (`docs/operating-modes.md` and `docs/troubleshooting.md`) with source-IP limitations and fail-closed behavior.
