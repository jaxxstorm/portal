## 1. Docsify Documentation System

- [x] 1.1 Add Docsify scaffolding under `docs/` (entry page, sidebar/navigation, and local serve instructions).
- [x] 1.2 Create docs pages for operating modes (proxy, mock, local Tailscale, tsnet) including tailnet-default and funnel-opt-in behavior.
- [x] 1.3 Create docs pages for configuration (`~/.tgate/config.yml`, `TGATE_*`, precedence rules, and examples).

## 2. README Declutter And Style Cleanup

- [x] 2.1 Refactor `README.md` into a concise overview with installation, quick-start, and links into `docs/`.
- [x] 2.2 Remove all emoji characters from `README.md` content and headings.
- [x] 2.3 Move deep operational/troubleshooting sections from `README.md` into dedicated `docs/` pages.

## 3. Documentation Policy For Future Changes

- [x] 3.1 Add contributor-facing guidance stating that user-visible feature changes must include corresponding updates in `docs/`.
- [x] 3.2 Link documentation expectations from the main project entry points (README and/or contribution guidance).

## 4. Validation And Consistency Checks

- [x] 4.1 Validate that docs include tailnet-only usage guidance with `tgate <port>`.
- [x] 4.2 Validate that docs include funnel-public usage guidance with `tgate <port> --funnel` and prerequisite notes.
- [x] 4.3 Validate that README-to-docs links resolve and docs examples are consistent with current CLI behavior.
