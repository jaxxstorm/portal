# Contributing

## Development

```bash
go test ./...
```

## Documentation Requirement

When a change introduces or modifies user-visible behavior, update the relevant
documentation in `docs/` as part of the same change.

This includes:
- CLI usage changes
- Configuration behavior changes
- Operating mode behavior changes
- Security/prerequisite guidance changes

## Local Docs Preview

```bash
npm install -g docsify-cli
docsify serve docs
```

Open `http://localhost:3000` to review docs updates.
