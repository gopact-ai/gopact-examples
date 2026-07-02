# Agnes Chat Quickstart

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

这个示例通过 `gopact-ext/models/agnes` 调用 Agnes provider。默认单元测试使用 fake server；真实 Agnes 服务测试需要本地 `.env`。

配置仓库根目录 `.env`：

```dotenv
GOPACT_LLM_BASEURL=https://apihub.agnes-ai.com/v1
GOPACT_LLM_TOKEN=your-agnes-token
GOPACT_LLM_MODEL=agnes-2.0-flash
```

也可以使用 Agnes-specific override：

```dotenv
GOPACT_AGNES_API_KEY=your-agnes-token
GOPACT_AGNES_SK=your-agnes-token
GOPACT_AGNES_MODEL=agnes-2.0-flash
```

运行：

```bash
go run ./quickstart/agnes-chat
```

本地真实服务测试：

```bash
go test -tags=integration -count=1 ./quickstart/agnes-chat
```
