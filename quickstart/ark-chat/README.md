# Ark Chat

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)


<!-- gopact:doc-language: zh,en -->

## 中文

本文档是 gopact 开源文档集的一部分，中文内容用于说明当前仓库约束、能力或维护流程。

## English

This document is part of the gopact open-source documentation set. The English section gives an entry point for readers who prefer English, while the remaining sections preserve the maintained technical details.


Run a single Ark Chat Completions call through `gopact-ext/models/ark` and the official Ark SDK.

```bash
GOPACT_ARK_API_KEY=your-ark-api-key \
GOPACT_ARK_MODEL=your-ark-endpoint-id \
go run ./quickstart/ark-chat
```

Optional: set `GOPACT_ARK_BASEURL`, `GOPACT_ARK_REGION`, or use `GOPACT_ARK_ACCESS_KEY` + `GOPACT_ARK_SECRET_KEY` instead of `GOPACT_ARK_API_KEY`.
