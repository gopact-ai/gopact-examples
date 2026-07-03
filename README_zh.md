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
go run ./quickstart/supervisor
go run ./quickstart/agent-as-tool
go run ./quickstart/background-scheduler
go run ./quickstart/self-bootstrap
go run ./quickstart/agent-node
go run ./quickstart/agent-cluster
```

这条路径从单个 scripted ReAct agent 开始，逐步扩展到 Plan-Execute、supervisor 路由、agent-as-tool 委托、background scheduling、agent-as-graph-node 编排和本地 A2A agent cluster。配置 `.env` 后再运行 provider quickstart。

## Quickstarts

所有示例都可以从仓库根目录运行：

```bash
go run ./quickstart/agent-as-tool
go run ./quickstart/background-scheduler
go run ./quickstart/self-bootstrap
go run ./quickstart/agent-cluster
go run ./quickstart/agent-node
go run ./quickstart/agent-scaffold
go run ./quickstart/agnes-chat
go run ./quickstart/ark-chat
go run ./quickstart/ark-streaming
go run ./quickstart/generated-agent
go run ./quickstart/generated-cluster
go run ./quickstart/openai-chat
go run ./quickstart/openai-streaming
go run ./quickstart/plan-exec
go run ./quickstart/react-agent
go run ./quickstart/structured-output
go run ./quickstart/supervisor
go run ./quickstart/tool-calling
go run ./quickstart/workflow-graph
```

| 示例 | 说明 | 是否需要真实凭据 |
| --- | --- | --- |
| `quickstart/react-agent` | scripted ReAct loop，演示本地 tool calling。 | 否 |
| `quickstart/workflow-graph` | typed graph、branch fan-out/fan-in、subgraph、loop、step limit、step export/import resume。 | 否 |
| `quickstart/agent-scaffold` | checkpoint、approval interrupt/resume、verification bundle、A2A file registry。 | 否 |
| `quickstart/generated-agent` | 调用 core `gopact agent init`、`agent verify` 和 `agent run`，验证默认 module path、生成 agent 的测试、registry 与运行路径。 | 否 |
| `quickstart/generated-cluster` | 调用 core `gopact agent init-cluster`、`agent verify` 和 `agent run`，验证默认 module path、生成 cluster 的 registry、env registry bootstrap、mesh 和运行路径。 | 否 |
| `quickstart/plan-exec` | Plan-Execute、replan、approval resume、cancel 测试覆盖。 | 否 |
| `quickstart/supervisor` | supervisor 路由到具名 Plan-Execute 子 agent。 | 否 |
| `quickstart/agent-as-tool` | 父 ReAct agent 将 Plan-Execute 子 agent 当作 tool 调用。 | 否 |
| `quickstart/background-scheduler` | 带 lease 的后台任务，覆盖 retry、dead-letter、drain 和 schedule evidence。 | 否 |
| `quickstart/self-bootstrap` | Dev Agent self-bootstrap workflow，覆盖 policy-approved plan patch apply、quickstart release requirements、diff、file snapshot、command、CI gate、run export、failure attribution 和 verification report evidence。 | 否 |
| `quickstart/agent-node` | 将 A2A 子 agent 挂成 typed graph node，并保留嵌套 evidence。 | 否 |
| `quickstart/agent-cluster` | 本地 A2A cluster、mesh-level HTTP options、`Mesh.SyncEnv`/`Mesh.SyncEnvEvery` discovery、policy、retry、cancel、dev-agent replay 和 command evidence。 | 否 |
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

agent-cluster quickstart 通过 `WithMeshHTTPAgentOptions` 一次性配置 discovery，再使用 `Mesh.SyncEnv` 导入环境变量配置的 agent cards，注册可调用 HTTP agents，并在路由任务前剔除未就绪 endpoint；测试使用 `Mesh.SyncEnvEvery` 覆盖连续 registry refresh。

## 本地集成测试

CI 必须保持 mock-only。真实 provider 测试只在本地显式运行：

```bash
./scripts/local-agnes-integration.sh
go test -tags=integration -count=1 ./quickstart/agnes-chat
```

## 开发验证

提交 PR 前运行：

```bash
git diff --check
./scripts/public-readiness-check.sh
./scripts/self-bootstrap-mock-suite.sh
./scripts/ecosystem-self-bootstrap-mock-suite.sh
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
