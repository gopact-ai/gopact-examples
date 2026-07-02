# Structured Output Quickstart

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

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
