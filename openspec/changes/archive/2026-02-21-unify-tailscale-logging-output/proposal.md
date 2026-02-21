## Why

In TUI and console modes, some Tailscale/tsnet runtime messages bypass tgate's primary logger and write directly to a separate stdout/stderr stream. This splits operator visibility across channels, degrades UX, and makes troubleshooting harder in the exact moments where startup/connectivity logs matter most.

## What Changes

- Route all Tailscale-related runtime output (local daemon integration, tsnet lifecycle, and associated subsystem logs) through tgate's primary logging pipeline.
- Remove or prevent out-of-band Tailscale/tsnet log emission that currently appears outside the main application log channel.
- Preserve consistent behavior across run modes (`tgate <port>`, `tgate <port> --funnel`, `tgate --mock`) and across TUI and non-TUI execution.
- Standardize logging context for these events so startup/setup/failure logs remain searchable and coherent in one place.
- Update docs/troubleshooting guidance to reflect the unified logging behavior and expected output location.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `local-service-exposure`: tighten runtime visibility expectations so Tailscale/tsnet lifecycle logs are emitted through the same application logger channel used by other tgate runtime events.

## Impact

- Affected code: Tailscale/tsnet setup and lifecycle paths, logger adapters/wrappers in startup flows, and TUI log routing.
- Affected docs: troubleshooting and operating behavior docs where log location/visibility is described.
- Dependencies: no new external services expected; may require using existing logging hooks/options more consistently in Tailscale integration.
- Risk areas: accidentally suppressing important low-level diagnostics, duplicate log emission, or regressions between local-daemon and tsnet modes.
