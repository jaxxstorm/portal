## Why

Tailnet membership or Funnel allowlists alone cannot protect a shared endpoint
with credentials. Portal needs an opt-in authorization boundary now, while
leaving room to support identity-provider-based authorization later.

## What Changes

- Add an authorization configuration surface with a disabled-by-default mode.
- Add HTTP Basic authorization backed by a configured username and password.
- Reject unauthorized requests before mock handling or forwarding to the local
  service.
- Establish an extensible authorization-provider boundary for future OAuth
  providers without adding OAuth configuration or behavior in this change.
- Document authorization behavior and configuration for tailnet-private and
  public Funnel endpoints.

## Capabilities

### New Capabilities
- `endpoint-authorization`: Opt-in request authorization for exposed portal endpoints, initially using HTTP Basic credentials and designed for future providers.

### Modified Capabilities
- `runtime-configuration`: Resolve and validate authorization settings from portal configuration sources.
- `local-service-exposure`: Apply configured authorization consistently before serving tailnet-private or Funnel-public requests.

## Impact

- `internal/config` gains authorization settings and validation.
- `internal/proxy` gains request authorization before proxying or mock responses.
- Startup wiring, tests, and user documentation gain authorization coverage.
