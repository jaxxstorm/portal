# Troubleshooting

## Basic Checks

```bash
portal --version
tailscale status
```

## Tailnet-Only Connectivity

Verify your local target service:

```bash
curl localhost:8080
portal 8080 --verbose
```

If `--funnel` is not set, access remains tailnet-only.

## Tailscale And TSNet Log Location

Tailscale and tsnet lifecycle logs are emitted through portal's main logger:
- In TUI mode: the **Application Logs** panel
- In non-TUI mode: standard portal console log output

If you are debugging startup, certificate, or Funnel setup issues, use:

```bash
portal 8080 --verbose
```

or in TUI mode:

```bash
portal 8080
```

Healthy startup should include a definitive `Startup ready` log entry. If that
entry is missing, startup did not complete and reachability information may be
incomplete.

## Funnel Setup Issues

For public access (`--funnel`), verify prerequisites:
- HTTPS certificates enabled in Tailscale admin DNS settings
- Funnel entitlement available for your tailnet

Useful command:

```bash
portal 8080 --funnel --verbose
```

## Service Mode + Funnel Conflict

This combination is invalid:

```bash
portal 8080 --listen-mode service --funnel
```

Service mode and Funnel are mutually exclusive. Use one:
- service mode for tailnet service exposure
- listener mode with Funnel for public exposure

## Service Mode Host Identity

Service mode requires a tag-based host identity. If a node is authenticated as
a user device (no `tag:*` identity), portal fails startup before advertising
the service.

Check node tags:

```bash
tailscale status --json | jq '.Self.Tags'
```

If tags are empty, authenticate this node with a tagged identity and retry.

## Funnel Allowlist Denials (`403`)

If Funnel allowlist is configured, requests are denied when:
- Source IP is not in the allowlist
- Source IP cannot be resolved from trusted request metadata or socket remote address
- PROXY protocol is expected (root-path Funnel allowlist mode) but missing

Verify configuration:

```bash
echo "$PORTAL_FUNNEL_ALLOWLIST"
```

Example config file key:

```yaml
funnel-allowlist:
  - 203.0.113.10
  - 198.51.100.0/24
```

If you set a non-root `set-path` (for example `/api`), portal cannot use the
Funnel TCP+PROXY mode and falls back to trusted HTTP source metadata.

For full configuration and behavior details, see
[IP Whitelisting](ip-whitelisting.md).

## TUI Display Problems

Use console mode:

```bash
portal 8080 --no-tui --verbose
```

## Reset Serve State

```bash
portal --cleanup-serve
```

Then retry with a minimal configuration:

```bash
portal 8080
```
