# Tool Calling Quickstart

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

这个示例通过 `gopact-ext/models/openai` 演示两步 tool calling：第一次模型返回 tool call，本地执行工具，第二次把 tool result 交回模型生成最终回答。

```dotenv
GOPACT_LLM_BASEURL=https://api.openai.com/v1
GOPACT_LLM_TOKEN=your-token
GOPACT_LLM_MODEL=gpt-4o-mini
```

```bash
go run ./quickstart/tool-calling
```

它展示 provider-neutral tool schema、assistant tool call round-trip、`gopact.ToolMessage` 和最终模型调用。
