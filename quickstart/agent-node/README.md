# Agent Node

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: en -->

Chinese documentation: [README_zh.md](README_zh.md)

This example shows an A2A child agent mounted as a typed graph node through `agentnode.New`. It uses a local scripted A2A agent and does not require provider credentials.

```bash
go run ./quickstart/agent-node
```

It covers:

- Mapping graph state into an `a2a.Task`.
- Streaming A2A status and completion events through the parent graph run.
- Mapping the terminal `a2a.Result` back into typed workflow state.
- Preserving child agent evidence on parent graph events.

