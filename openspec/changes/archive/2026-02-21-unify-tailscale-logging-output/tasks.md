## 1. Unified Tailscale Log Channel

- [x] 1.1 Add a shared Tailscale/tsnet log adapter that routes lifecycle output into the primary application logger.
- [x] 1.2 Wire the adapter into both local-daemon and tsnet startup paths so tailnet mode uses the same logging channel end-to-end.
- [x] 1.3 Remove or gate legacy direct stdout/stderr writes from Tailscale lifecycle setup to prevent split or duplicate output.

## 2. Structured Lifecycle And Failure Logging

- [x] 2.1 Standardize structured fields for Tailscale lifecycle events (component, mode, phase, and relevant identifiers).
- [x] 2.2 Ensure HTTPS certificate and Funnel setup failures emit actionable structured errors through the primary logger before startup exits.
- [x] 2.3 Keep severity mapping and message preservation consistent so diagnostics remain usable after adapter routing.

## 3. Validation And Documentation

- [x] 3.1 Add/extend tests for tailnet-only startup to verify Tailscale lifecycle logs appear in the main logger channel without out-of-band stream output.
- [x] 3.2 Add/extend tests for Funnel startup and failure paths to verify certificate/Funnel errors are logged through the same main channel.
- [ ] 3.3 Validate behavior across `tgate <port>`, `tgate <port> --funnel`, and `tgate --mock` in both TUI and non-TUI runs.
- [x] 3.4 Update operator documentation and troubleshooting guidance to reflect unified Tailscale log output location.
