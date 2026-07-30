## 1. Readiness Status

- [x] 1.1 Add certificate and Funnel status values to the shared startup summary model and structured output fields.
- [x] 1.2 Render the same certificate, Funnel, and public URL readiness information in non-TUI text, JSON, and TUI startup output.
- [x] 1.3 Add summary tests for tailnet-private disabled/not-required status and Funnel verified/ready status.

## 2. Funnel Lifecycle Reporting

- [x] 2.1 Thread confirmed local-daemon HTTPS certificate and Funnel setup milestones into startup readiness reporting.
- [x] 2.2 Thread tsnet Funnel readiness into the same reporting contract without claiming unsupported certificate detail.
- [x] 2.3 Add structured failure status and actionable context for certificate verification and Funnel activation failures while preserving the no-ready-summary behavior.

## 3. Validation And Documentation

- [x] 3.1 Add local-daemon and tsnet tests for successful Funnel readiness and certificate/Funnel failure paths.
- [x] 3.2 Update operator documentation with `portal <port>` and `portal <port> --funnel` status examples, prerequisites, and readiness limitations.
- [x] 3.3 Run the Go test suite and OpenSpec validation.
