# OpenAI Chat Quickstart

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

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
