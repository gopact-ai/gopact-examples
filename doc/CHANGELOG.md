# Changelog

<!-- gopact:doc-language: en -->

Chinese documentation: [CHANGELOG_zh.md](CHANGELOG_zh.md)

This changelog records user-visible changes for `gopact-examples`. The current unreleased work improves repository documentation while preserving mock-only CI and local opt-in provider integration.

## Unreleased

- Examples now track `gopact` core `v0.0.40` and the current `gopact-ext` release tags.
- The agent cluster quickstart now covers `Mesh.SyncEnv` and `Mesh.SyncEnvEvery` for env-driven A2A discovery, HTTP agent registration, readiness pruning, and registry changes.
- The agent cluster quickstart demonstrates A2A lease heartbeat evidence plus replay and command evidence.
- Add `quickstart/supervisor` to demonstrate routing work to named Plan-Execute child agents without provider credentials.
- Default documentation is English-only, with Chinese translations maintained in sibling `_zh.md` files.
- The root README has a credential-free scaffold path, complete quickstart command list, environment variable reference, and CI gate summary.
- The feature matrix records tested quickstarts, provider paths, A2A discovery paths, and opt-in integration commands.
