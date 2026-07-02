# Ark Chat

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: zh,en -->

## 中文

这个示例通过 `gopact-ext/models/ark` 和 Volcengine Ark SDK 发起一次 chat completion。它和 `ark-streaming` 的区别是：这里走 Ark SDK provider；`ark-streaming` 把 Ark endpoint 当作 OpenAI-compatible Responses API。

```bash
GOPACT_ARK_API_KEY=your-ark-api-key \
GOPACT_ARK_MODEL=your-ark-endpoint-id \
go run ./quickstart/ark-chat
```

可选变量：

- `GOPACT_ARK_BASEURL`
- `GOPACT_ARK_REGION`
- `GOPACT_ARK_ACCESS_KEY`
- `GOPACT_ARK_SECRET_KEY`

如果提供 AK/SK，示例会使用 Ark SDK 的 AK/SK 初始化路径；否则使用 `GOPACT_ARK_API_KEY`。

## English

This example calls Ark through `gopact-ext/models/ark` and the Volcengine Ark SDK. Use this quickstart for the Ark SDK path; use `ark-streaming` when testing an Ark OpenAI-compatible Responses endpoint.
