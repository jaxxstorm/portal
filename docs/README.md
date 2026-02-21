# tgate Documentation

tgate exposes local development services over Tailscale, with private-by-default
tailnet access and optional public access through Tailscale Funnel.

## Start Here

- [Operating Modes](operating-modes.md)
- [Configuration](configuration.md)
- [IP Whitelisting](ip-whitelisting.md)
- [Troubleshooting](troubleshooting.md)
- [Documentation Policy](documentation-policy.md)

## Local Docs Preview

Use Docsify to serve this folder locally:

```bash
# Install docsify once
npm install -g docsify-cli

# Serve docs site
docsify serve docs
```

Then open `http://localhost:3000`.
