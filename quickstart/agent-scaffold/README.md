# Agent Scaffold

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: en -->

Chinese documentation: [README_zh.md](README_zh.md)

This is the smallest no-credential agent scaffold: typed graph, checkpointing, approval interrupt/resume, verification report, and A2A file registry in one local flow before the larger self-bootstrap release gate path in `quickstart/agent-cluster`.

The example records a `RunExport`, builds a verification report, and embeds the report in the release evidence bundle. Use it as the local scaffold before adopting the self-bootstrap release gate path in `quickstart/agent-cluster`.
