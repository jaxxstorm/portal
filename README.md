# portal

portal is an HTTP proxy and testing tool for exposing local services over
Tailscale. It is private by default (tailnet-only) and can be made public with
explicit Funnel opt-in.

## Installation

### Homebrew

```bash
brew tap jaxxstorm/tap
brew install portal
```

### From Source

```bash
git clone https://github.com/jaxxstorm/portal.git
cd portal
go mod tidy
go build -o portal main.go
```

## Quick Start

```bash
# Tailnet-only proxy to local service
portal 8080

# Public internet access with Funnel
portal 8080 --funnel

# Mock endpoint for webhook testing (tailnet-only by default)
portal --mock

# Mock endpoint with explicit public Funnel exposure
portal --mock --funnel
```

## Documentation

Detailed documentation is in `docs/` and can be served with Docsify.

- [Docs Home](docs/README.md)
- [Operating Modes](docs/operating-modes.md)
- [Configuration](docs/configuration.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Documentation Policy](docs/documentation-policy.md)

### Run Docs Locally

```bash
npm install -g docsify-cli
docsify serve docs
```

Open `http://localhost:3000`.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).  
Any user-visible behavior change must include matching updates in `docs/`.

## License

MIT. See `LICENSE`.
