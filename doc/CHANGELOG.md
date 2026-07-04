# Changelog

<!-- gopact:doc-language: en -->

Chinese documentation: [CHANGELOG_zh.md](CHANGELOG_zh.md)

This changelog records user-visible changes for `gopact-examples`. The current unreleased work improves repository documentation while preserving mock-only CI and local opt-in provider integration.

## Unreleased

- Add an ecosystem self-bootstrap mock suite that validates `gopact`, `gopact-ext`, and `gopact-examples` through their mock-only self-bootstrap gates.
- Add `quickstart/release-bundle` to demonstrate core `gopact release-bundle` using a recorded run export and observed verification report.
- Add `quickstart/generated-cluster` to exercise core `gopact agent init-cluster`, `agent verify`, and `agent run` against a generated local A2A cluster.
- The generated agent and generated cluster quickstarts now use core `v0.0.53`, exercise default module paths, and cover generated cluster `GOPACT_A2A_REGISTRY_URL` bootstrap behavior.
- Add `quickstart/background-scheduler` to demonstrate leased background jobs with retry, dead-letter, drain, and schedule verification evidence.
- Add `quickstart/self-bootstrap` to demonstrate the reusable Dev Agent self-bootstrap workflow with policy-approved plan patch apply, quickstart release requirements, diff, file snapshot, command, CI gate, run export, and verification report evidence.
- The workflow graph quickstart now demonstrates completed step export/import resume and interrupted resume without provider credentials.
- Examples now track `gopact` core `v0.0.53` and the matching `gopact-ext` release tags for that core line.
- The agent cluster quickstart now demonstrates mesh-level HTTP options for environment-driven A2A discovery.
- Add `quickstart/agent-node` to demonstrate mounting an A2A child agent as a typed graph node with nested evidence.
- The agent cluster quickstart now covers `Mesh.SyncEnv` and `Mesh.SyncEnvEvery` for env-driven A2A discovery, HTTP agent registration, readiness pruning, and registry changes.
- The agent cluster quickstart demonstrates A2A lease heartbeat evidence plus replay and command evidence.
- Add `quickstart/supervisor` to demonstrate routing work to named Plan-Execute child agents without provider credentials.
- Default documentation is English-only, with Chinese translations maintained in sibling `_zh.md` files.
- The root README has a credential-free scaffold path, complete quickstart command list, environment variable reference, and CI gate summary.
- The feature matrix records tested quickstarts, provider paths, A2A discovery paths, and opt-in integration commands.
