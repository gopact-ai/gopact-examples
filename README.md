# gopact-examples

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/gopact-ai/gopact-examples.svg)](https://pkg.go.dev/github.com/gopact-ai/gopact-examples)

<!-- gopact:doc-language: zh,en -->

## 中文

`gopact-examples` 提供 `github.com/gopact-ai/gopact` 和官方 extension 的可运行示例。CI 使用本地 fake LLM server 和 mock 数据，不需要真实 provider credential。

## 配置

示例会读取这些环境变量：

- `GOPACT_LLM_BASEURL`：OpenAI-shaped `/v1` API base URL。
- `GOPACT_LLM_TOKEN`：API token。
- `GOPACT_LLM_MODEL`：model name。
- `GOPACT_A2A_REGISTRY_FILE`：`quickstart/agent-cluster` 可选 A2A agent-card JSON 文件。
- `GOPACT_A2A_REGISTRY_URL`：`quickstart/agent-cluster` 可选 HTTP A2A agent-card registry URL。
- `GOPACT_A2A_ENDPOINTS`：`quickstart/agent-cluster` 可选逗号分隔 A2A HTTP agent endpoints。

`quickstart/ark-chat` 使用 Ark SDK 变量：`GOPACT_ARK_API_KEY`，或 `GOPACT_ARK_ACCESS_KEY` + `GOPACT_ARK_SECRET_KEY`，以及 `GOPACT_ARK_MODEL`。

默认情况下，示例会从当前目录或父目录加载 `.env`。`.env` 已被 git ignore。

```bash
cp .env.example .env
```

## Scaffold Path

Start without credentials:

```bash
go run ./quickstart/react-agent
go run ./quickstart/agent-scaffold
go run ./quickstart/generated-agent
go run ./quickstart/plan-exec
go run ./quickstart/agent-as-tool
go run ./quickstart/agent-cluster
```

This path grows from one scripted tool-using agent to a checkpointed approval/resume scaffold, the core `gopact agent init` generator, a Plan-Execute workflow, an agent-as-tool bridge, and a local A2A cluster. Use provider quickstarts after `.env` is configured.

## 示例

```bash
go run ./quickstart/agent-as-tool
go run ./quickstart/agent-cluster
go run ./quickstart/agent-scaffold
go run ./quickstart/agnes-chat
go run ./quickstart/ark-chat
go run ./quickstart/ark-streaming
go run ./quickstart/generated-agent
go run ./quickstart/openai-chat
go run ./quickstart/openai-streaming
go run ./quickstart/plan-exec
go run ./quickstart/react-agent
go run ./quickstart/structured-output
go run ./quickstart/tool-calling
go run ./quickstart/workflow-graph
```

## 本地集成测试

CI stays mock-only. To verify the provider-backed Agnes quickstart locally, put one of `GOPACT_AGNES_API_KEY`, `GOPACT_AGNES_SK`, or `GOPACT_LLM_TOKEN` in `.env`, then run:

```bash
go test -tags=integration -count=1 ./quickstart/agnes-chat
```

## 文档索引

- [doc/README.md](./doc/README.md)：完整文档索引。
- [doc/FEATURES.md](./doc/FEATURES.md)：可执行能力覆盖矩阵。
- [doc/CONTRIBUTING.md](./doc/CONTRIBUTING.md)：贡献指南。
- [doc/SECURITY.md](./doc/SECURITY.md)：安全策略。
- [doc/CHANGELOG.md](./doc/CHANGELOG.md)：变更记录。
- [doc/maintainers/repository-governance.md](./doc/maintainers/repository-governance.md)：PR、CI、自动合并和公开前检查规则。

## 开发

```bash
git diff --check
go mod tidy
git diff --exit-code
go test -count=1 ./...
go test -race -count=1 ./...
go vet ./...
golangci-lint run ./...
go test -coverprofile=coverage.out ./...
govulncheck ./...
```

## English

`gopact-examples` provides runnable examples for `github.com/gopact-ai/gopact` and the official extensions. CI uses local fake LLM servers and mock data, so real provider credentials are not required.

Use the scaffold path first when no credentials are available. Configure `.env` only when running provider-backed quickstarts. The full documentation index is [doc/README.md](./doc/README.md), and the executable capability matrix is [doc/FEATURES.md](./doc/FEATURES.md).
