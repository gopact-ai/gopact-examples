# Generated Cluster

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: en -->

Chinese documentation: [README_zh.md](README_zh.md)

This example runs the core `gopact agent init-cluster` generator, verifies the generated local A2A HTTP cluster through `gopact agent verify`, and serves it through `gopact agent run`.

```bash
go run ./quickstart/generated-cluster
```

The generated project exposes `/agents.json`, planner/worker/reviewer A2A endpoints, health/readiness endpoints, and generated tests for mesh bootstrap, `GOPACT_A2A_REGISTRY_URL` registry override, routing, streaming, cancel, and shutdown.
