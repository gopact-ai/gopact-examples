# Agnes Chat Quickstart

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: en -->

Chinese documentation: [README_zh.md](README_zh.md)

This example calls Agnes through `gopact-ext/models/agnes`. Default tests use a fake server. Real Agnes verification is opt-in with local credentials from `.env`, using `GOPACT_AGNES_API_KEY`, `GOPACT_AGNES_SK`, or `GOPACT_LLM_TOKEN`.

Run the deterministic mock test:

```bash
go test -count=1 ./quickstart/agnes-chat
```

Run the local provider-backed test after configuring `.env`:

```bash
go test -tags=integration -count=1 ./quickstart/agnes-chat
```
