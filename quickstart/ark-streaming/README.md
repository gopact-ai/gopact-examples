# Ark Streaming

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)


<!-- gopact:doc-language: zh,en -->

## 中文

本文档是 gopact 开源文档集的一部分，中文内容用于说明当前仓库约束、能力或维护流程。

## English

This document is part of the gopact open-source documentation set. The English section gives an entry point for readers who prefer English, while the remaining sections preserve the maintained technical details.


Stream an Ark OpenAI-compatible Responses call through `gopact-ext/models/openai`.
The example runs both `thinking=disabled` and `thinking=enabled` with `max_output_tokens=1024`.
Reasoning tokens count against the output budget, so tiny values like `64` can finish with reasoning-only output on thinking models.

```bash
GOPACT_LLM_BASEURL=https://ark.cn-beijing.volces.com/api/v3 \
GOPACT_LLM_TOKEN=your-ark-api-key \
GOPACT_LLM_MODEL=your-ark-endpoint-id \
go run ./quickstart/ark-streaming
```
