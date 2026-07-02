# gopact-examples

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

`gopact-examples` 是 `github.com/gopact-ai/gopact` 和 `gopact-ext` 的可运行示例仓库。它的目标不是展示孤立代码片段，而是把 core workflow、agent template、provider adapter、A2A discovery、verification 和 dev-agent evidence 串成可以本地执行、可以被 CI 固化的用法。

CI 使用 fake LLM server、scripted model 和本地 A2A agent，不需要真实 provider credential。真实 provider 示例保留为本地 opt-in 测试，必须由 `.env` 提供凭据。

## Scaffold Path

Start without credentials:

```bash
go run ./quickstart/react-agent
go run ./quickstart/plan-exec
go run ./quickstart/agent-as-tool
go run ./quickstart/agent-cluster
```

这条路径从单个 scripted ReAct agent 开始，逐步扩展到 Plan-Execute、agent-as-tool 委托和本地 A2A agent cluster。Use provider quickstarts after `.env` is configured.

## Quickstarts

所有示例都可以从仓库根目录运行：

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

| 示例 | 说明 | 是否需要真实凭据 |
| --- | --- | --- |
| `quickstart/react-agent` | scripted ReAct loop，演示本地 tool calling。 | 否 |
| `quickstart/workflow-graph` | typed graph、branch fan-out/fan-in、subgraph、loop、step limit。 | 否 |
| `quickstart/agent-scaffold` | checkpoint、approval interrupt/resume、verification bundle、A2A file registry。 | 否 |
| `quickstart/generated-agent` | 调用 core `gopact agent init`，验证生成 agent 的 run 和 registry。 | 否 |
| `quickstart/plan-exec` | Plan-Execute、replan、approval resume、cancel 测试覆盖。 | 否 |
| `quickstart/agent-as-tool` | 父 ReAct agent 将 Plan-Execute 子 agent 当作 tool 调用。 | 否 |
| `quickstart/agent-cluster` | 本地 A2A cluster、multi-source discovery、policy、retry、cancel、dev-agent evidence。 | 否 |
| `quickstart/openai-chat` | OpenAI-compatible chat completions。 | 是 |
| `quickstart/openai-streaming` | OpenAI Chat Completions 和 Responses 两种 streaming API。 | 是 |
| `quickstart/tool-calling` | OpenAI-compatible model tool calling。 | 是 |
| `quickstart/structured-output` | JSON schema structured output。 | 是 |
| `quickstart/ark-chat` | Ark SDK provider。 | 是 |
| `quickstart/ark-streaming` | Ark OpenAI-compatible Responses streaming。 | 是 |
| `quickstart/agnes-chat` | Agnes provider。 | 是 |

## 配置

默认情况下，示例会从当前目录或父目录加载 `.env`。`.env` 已在 `.gitignore` 中排除；仓库只提交 `.env.example`。

```bash
cp .env.example .env
```

通用 OpenAI-shaped provider 示例读取：

- `GOPACT_LLM_BASEURL`
- `GOPACT_LLM_TOKEN`
- `GOPACT_LLM_MODEL`

Agnes 示例支持通用变量，也支持 provider-specific override：

- `GOPACT_AGNES_API_KEY`
- `GOPACT_AGNES_SK`
- `GOPACT_AGNES_MODEL`

Ark SDK 示例读取：

- `GOPACT_ARK_API_KEY`
- `GOPACT_ARK_ACCESS_KEY`
- `GOPACT_ARK_SECRET_KEY`
- `GOPACT_ARK_MODEL`
- `GOPACT_ARK_REGION`

A2A cluster discovery 支持：

- `GOPACT_A2A_REGISTRY_FILE`
- `GOPACT_A2A_REGISTRY_URL`
- `GOPACT_A2A_ENDPOINTS`

## 本地集成测试

CI 必须保持 mock-only。真实 provider 测试只在本地显式运行：

```bash
go test -tags=integration -count=1 ./quickstart/agnes-chat
```

## 开发验证

提交 PR 前运行：

```bash
git diff --check
./scripts/public-readiness-check.sh
go mod tidy
git diff --exit-code
go test -count=1 ./...
go test -race -count=1 ./...
go vet ./...
golangci-lint run ./...
go test -coverprofile=coverage.out ./...
govulncheck ./...
```

## 文档索引

- [doc/README.md](./doc/README.md)：文档地图与推荐阅读顺序。
- [doc/FEATURES.md](./doc/FEATURES.md)：可执行能力覆盖矩阵。
- [doc/CONTRIBUTING.md](./doc/CONTRIBUTING.md)：贡献流程、本地验证和 PR 要求。
- [doc/SECURITY.md](./doc/SECURITY.md)：安全策略与漏洞报告方式。
- [doc/CHANGELOG.md](./doc/CHANGELOG.md)：变更记录。
- [doc/maintainers/repository-governance.md](./doc/maintainers/repository-governance.md)：PR-only、CI 门禁、admin auto-merge 和公开前检查。
