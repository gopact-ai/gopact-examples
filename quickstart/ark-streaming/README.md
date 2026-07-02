# Ark Streaming

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: zh,en -->

## 中文

这个示例把 Ark endpoint 作为 OpenAI-compatible Responses API 使用，通过 `gopact-ext/models/openai` 演示 streaming。它会分别运行 `thinking=disabled` 和 `thinking=enabled`，并使用 `max_output_tokens=1024` 避免 thinking 模型只返回 reasoning 而没有可见文本。

```bash
GOPACT_LLM_BASEURL=https://ark.cn-beijing.volces.com/api/v3 \
GOPACT_LLM_TOKEN=your-ark-api-key \
GOPACT_LLM_MODEL=your-ark-endpoint-id \
go run ./quickstart/ark-streaming
```

如果要测试 Ark SDK provider，请使用 `quickstart/ark-chat`。

## English

This example treats an Ark endpoint as an OpenAI-compatible Responses API and streams through `gopact-ext/models/openai`. It runs both thinking-disabled and thinking-enabled requests with a reasonable output budget.
