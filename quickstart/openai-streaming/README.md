# OpenAI Streaming

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: zh,en -->

## 中文

这个示例通过 `gopact-ext/models/openai` 同时演示两种 OpenAI API surface 的 streaming：

- `openai.WithChatCompletionsAPI()` 使用 `/chat/completions` 和 `messages` payload。
- `openai.WithResponsesAPI()` 使用 `/responses` 和 `input` payload。

```dotenv
GOPACT_LLM_BASEURL=https://api.openai.com/v1
GOPACT_LLM_TOKEN=your-token
GOPACT_LLM_MODEL=gpt-4o-mini
```

```bash
go run ./quickstart/openai-streaming
```

优先使用 Chat Completions 兼容面接通多数 provider；当 provider 支持 Responses、rich content part 或 reasoning delta 时，再使用 Responses。

## English

This example streams both OpenAI API surfaces through `gopact-ext/models/openai`: Chat Completions and Responses. It demonstrates that `Stream` uses the same provider-neutral `gopact.ModelRequest` contract.
