# OpenAI Chat Quickstart

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: zh,en -->

## 中文

这个示例通过 `gopact-ext/models/openai` 发起一次 OpenAI-compatible chat completion。默认测试使用 fake server，真实服务运行需要 `.env`。

```dotenv
GOPACT_LLM_BASEURL=https://api.openai.com/v1
GOPACT_LLM_TOKEN=your-token
GOPACT_LLM_MODEL=gpt-4o-mini
```

```bash
go run ./quickstart/openai-chat
```

它演示 provider 初始化、system/user message 构造、per-call temperature 设置和 response text 读取。

## English

This example makes one OpenAI-compatible chat completion call through `gopact-ext/models/openai`. Default tests use a fake server; real provider runs read `GOPACT_LLM_BASEURL`, `GOPACT_LLM_TOKEN`, and `GOPACT_LLM_MODEL` from `.env`.
