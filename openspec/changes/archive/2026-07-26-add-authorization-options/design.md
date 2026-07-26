## Context

Portal parses CLI, environment, and YAML configuration into `config.Config`, then
constructs one `proxy.Server` that is used by both local-Tailscale and tsnet
serving paths. The proxy currently enforces Funnel source-IP allowlists before
dispatching to the mock backend or reverse proxy. Authorization must add a
credential boundary without changing tailnet-private-by-default or the explicit
Funnel-public exposure model.

## Goals / Non-Goals

**Goals:**
- Add disabled-by-default HTTP Basic authorization configured through the same
  sources and precedence as other runtime settings.
- Validate incomplete or unsupported authorization configuration before startup.
- Authenticate every incoming proxy request before it reaches either backend.
- Isolate authorization implementations behind a small request-authorizer
  interface so OAuth can be added without changing proxy dispatch.

**Non-Goals:**
- OAuth provider configuration, browser redirects, sessions, tokens, or user
  identity propagation.
- Protecting Portal's separate web UI endpoint.
- Encrypting, hashing, or persisting configured Basic-auth credentials.
- Changing Tailscale, tsnet, HTTPS-certificate, Funnel, or allowlist setup.

## Decisions

### Configure one explicit HTTP Basic mode

Add `auth-mode`, `auth-username`, and `auth-password` settings, exposed as
`--auth-mode`, `--auth-username`, `--auth-password`, matching `PORTAL_*`
variables, and YAML keys. `auth-mode` defaults to `none`; `basic` requires both
credentials and no credentials are accepted when the mode is `none`.

An explicit mode avoids treating an accidental username or password as a
security configuration, and makes a future `oauth` mode a compatible extension.
Using a standalone `--basic-auth` boolean was rejected because it couples the
configuration surface to one provider and makes future provider selection less
clear.

### Authorize in the proxy handler before request dispatch

Add an `Authorizer` interface with one request-authorization method and a
Basic implementation using `r.BasicAuth` plus constant-time comparisons. Pass
the constructed authorizer in `proxy.Config`. `Server.ServeHTTP` first performs
Funnel allowlist enforcement, then authorization, then mock or reverse-proxy
dispatch. An unsuccessful authorization writes `401 Unauthorized` and
`WWW-Authenticate: Basic realm="portal"` and does not forward or mock-handle
the request.

Keeping Funnel checks first retains existing source-restriction behavior and
does not reveal a credential challenge to a disallowed public source. Placing
the check in Tailscale or tsnet setup was rejected because both serving paths
share the proxy handler and only it reliably sees every request.

### Keep the authorization boundary provider-neutral

The proxy depends only on the `Authorizer` interface; configuration selects and
builds the initial Basic implementation during startup. A nil authorizer means
authorization is disabled. Future OAuth providers can implement the same
interface and add a configuration mode without changing handler control flow.

## Risks / Trade-offs

- [Basic credentials are static and sent on each request] → Require users to
  use HTTPS for public Funnel endpoints, document this limitation, and leave
  stronger provider types for a follow-up.
- [Configuration files and process environments can expose secrets] → Avoid
  logging credentials and document using environment variables or protected
  config-file permissions.
- [An authorization layer could inadvertently block readiness probes] → Apply
  it only to normal proxy traffic and test authorized and rejected mock paths.
- [Funnel HTTPS certificates can still fail before serving] → Authorization does
  not alter Funnel setup; existing certificate prerequisite failures remain
  visible through the primary startup logger.

## Migration Plan

Existing configurations default to `auth-mode=none`, so deployment does not
change behavior. Operators enable Basic authorization by configuring all three
settings and restarting Portal. Removing the settings or setting `auth-mode`
back to `none` rolls back the protection. Tailnet runs such as `portal <port>`
remain private by default; Funnel runs such as `portal <port> --funnel` remain
public only when Funnel and its HTTPS prerequisites are explicitly enabled.

## Open Questions

None for the initial HTTP Basic provider. OAuth provider names, credentials,
redirect URLs, and callback handling will be specified when that provider is
introduced.
