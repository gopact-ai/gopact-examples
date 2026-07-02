# Agnes Chat Quickstart

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)


<!-- gopact:doc-language: zh,en -->

## 中文

本文档是 gopact 开源文档集的一部分，中文内容用于说明当前仓库约束、能力或维护流程。

## English

This document is part of the gopact open-source documentation set. The English section gives an entry point for readers who prefer English, while the remaining sections preserve the maintained technical details.


Run a single chat completion through `gopact-ext/models/agnes`.

## Configure

Create `.env` at the repository root:

```dotenv
GOPACT_LLM_BASEURL=https://apihub.agnes-ai.com/v1
GOPACT_LLM_TOKEN=your-agnes-token
GOPACT_LLM_MODEL=agnes-2.0-flash
```

You can also use Agnes-specific credentials:

```dotenv
GOPACT_AGNES_API_KEY=your-agnes-token
GOPACT_AGNES_SK=your-agnes-token
GOPACT_AGNES_MODEL=agnes-2.0-flash
```

## Run

```bash
go run ./quickstart/agnes-chat
```

## Local Integration

```bash
go test -tags=integration -count=1 ./quickstart/agnes-chat
```
