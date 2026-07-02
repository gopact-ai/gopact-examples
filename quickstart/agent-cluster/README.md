# Agent Cluster

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: en -->

Chinese documentation: [README_zh.md](README_zh.md)

This example runs a local A2A-style agent cluster with planner, research, code, and review agents. It demonstrates multi-source discovery, routing, policy evidence, retry/cancel evidence, development-agent evidence, and a self-bootstrap release-gate bundle without requiring provider credentials.

The discovery path can be overridden with:

- `GOPACT_A2A_REGISTRY_FILE`
- `GOPACT_A2A_REGISTRY_URL`
- `GOPACT_A2A_ENDPOINTS`

Run it from the repository root:

```bash
go run ./quickstart/agent-cluster
go test -count=1 ./quickstart/agent-cluster
```
