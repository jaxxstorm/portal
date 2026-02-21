## Context

The current documentation is concentrated in `README.md`, which mixes quick-start content with detailed operational guidance. This makes it harder to scan and maintain, and contributes to documentation drift as features evolve.

This change introduces a Docsify-based `docs/` site as the long-form documentation system, while turning `README.md` into a concise entry point. Documentation content must explicitly cover:
- Operating modes (proxy mode, mock mode, local Tailscale mode, tsnet mode)
- Tailnet-private vs Funnel-public behavior and prerequisites
- Configuration model (`~/.tgate/config.yml`, `TGATE_*`, CLI precedence)
- Common failure handling topics (including HTTPS certificate and Funnel setup issues)

## Goals / Non-Goals

**Goals:**
- Establish `docs/` as the primary location for detailed project documentation using Docsify.
- Reduce `README.md` size and remove emojis while preserving essential project overview and entry commands.
- Document operating modes and configuration behavior clearly.
- Define contributor expectations so future feature changes include docs updates.
- Preserve current runtime and CLI behavior; only documentation structure and content expectations change.

**Non-Goals:**
- Changing proxy/networking behavior or security semantics.
- Changing CLI flags or configuration precedence behavior.
- Rewriting all examples from scratch if they can be relocated with minimal edits.

## Decisions

1. **Adopt Docsify for project docs**
   - Add `docs/` with standard Docsify entry files and navigation.
   - Keep documentation as markdown-first content in-repo.
   - Alternative considered: leave docs in README only. Rejected due to poor scalability and discoverability.

2. **Use README as a concise landing page**
   - Keep concise project summary, install, and quick-start pointers.
   - Move operational deep dives (modes, config internals, troubleshooting details) into `docs/`.
   - Remove all emojis for a cleaner, more maintainable style baseline.

3. **Document mode and security behavior explicitly**
   - Include clear sections for tailnet-private default behavior and funnel opt-in behavior.
   - Document local Tailscale and tsnet mode expectations and usage.
   - Include funnel prerequisites and failure handling guidance (certificate/admin setup issues).

4. **Documentation maintenance expectation for future changes**
   - Add contributor-facing documentation guidance that new behavior changes must include corresponding docs updates in `docs/`.
   - Keep this as process/documentation policy, not runtime enforcement.

## Risks / Trade-offs

- [Risk] README loses critical discoverability if over-trimmed -> Mitigation: keep install + quick-start + links to docs front-and-center.
- [Risk] Docsify setup adds minor tooling overhead -> Mitigation: keep setup minimal and document local preview command.
- [Risk] Docs may drift from implementation over time -> Mitigation: add explicit "update docs with feature changes" guidance in contributor documentation.
- [Risk] Funnel/certificate failure docs become stale -> Mitigation: reference stable behavior and prerequisites instead of brittle implementation details.

## Migration Plan

1. Create Docsify scaffolding under `docs/` and define navigation pages.
2. Split README content into concise overview + docs links.
3. Move mode/config/troubleshooting detail into dedicated docs pages.
4. Add contributor guidance requiring docs updates for future behavior changes.
5. Validate links and ensure README and docs are internally consistent.

## Open Questions

- Should Docsify be served through an npm script or a direct `docsify serve docs` command in contributor docs?
- Do we want to add CI link checks in a follow-up change to guard against docs drift?
