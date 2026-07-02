# Structured Output Quickstart

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: zh,en -->

## 中文

这个示例通过 `gopact-ext/models/openai` 发送 JSON schema response contract，并在本地再次校验模型返回是否满足 schema。

```dotenv
GOPACT_LLM_BASEURL=https://api.openai.com/v1
GOPACT_LLM_TOKEN=your-token
GOPACT_LLM_MODEL=gpt-4o-mini
```

```bash
go run ./quickstart/structured-output
```

它展示：

- client 默认启用 structured output capability。
- request 级别传入 `gopact.WithResponseSchema`。
- 返回文本先 `json.Unmarshal`，再调用 core schema validator。

## English

This example sends a JSON schema response contract through the OpenAI-compatible provider and validates the returned JSON locally.
