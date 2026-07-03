# Generated Agent Quickstart

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: en -->

Chinese documentation: [README_zh.md](README_zh.md)

This example runs the core `gopact agent init` generator, verifies the generated A2A HTTP agent project through `gopact agent verify`, and serves it through `gopact agent run`.

Run it from the repository root:

```bash
go run ./quickstart/generated-agent
```

The generated project exposes `/agents.json`, validates through `gopact agent verify .`, and can be served with `gopact agent run .`.
