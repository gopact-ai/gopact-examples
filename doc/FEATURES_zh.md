# Feature Coverage

<!-- gopact:doc-language: zh -->

[英文文档](./FEATURES.md)

## 中文

这个矩阵是 `gopact-examples` 的可执行能力契约。CI 只运行 mock、本地 fake server 和 scripted agent；真实 provider 测试必须通过 integration build tag 手动执行。

| Capability | Path | Mock test | Local integration |
| --- | --- | --- | --- |
| dotenv configuration | `internal/exampleenv` | `go test -count=1 ./internal/exampleenv` | - |
| scripted ReAct loop | `quickstart/react-agent` | `go test -count=1 ./quickstart/react-agent` | - |
| workflow graph branch, dynamic fan-out, fan-in, loop, subgraph, step limit, step export/import, and interrupted resume | `quickstart/workflow-graph` | `go test -count=1 ./quickstart/workflow-graph` | - |
| checkpoint approval resume | `quickstart/agent-scaffold` | `go test -count=1 ./quickstart/agent-scaffold` | - |
| verification bundle | `quickstart/agent-scaffold` | `go test -count=1 ./quickstart/agent-scaffold` | - |
| A2A file registry scaffold | `quickstart/agent-scaffold` | `go test -count=1 ./quickstart/agent-scaffold` | - |
| core agent init/verify/run scaffold | `quickstart/generated-agent` | `go test -count=1 ./quickstart/generated-agent` | - |
| Plan-Execute workflow with replan, approval resume, and cancel | `quickstart/plan-exec` | `go test -count=1 ./quickstart/plan-exec` | - |
| Supervisor routing to named Plan-Execute child agents | `quickstart/supervisor` | `go test -count=1 ./quickstart/supervisor` | - |
| agent as tool success and failure evidence | `quickstart/agent-as-tool` | `go test -count=1 ./quickstart/agent-as-tool` | - |
| leased background scheduler with retry, dead-letter, drain, lease release, and schedule evidence | `quickstart/background-scheduler` | `go test -count=1 ./quickstart/background-scheduler` | - |
| Dev Agent self-bootstrap workflow with policy-approved plan patch apply, quickstart release requirements, diff, file snapshot, command, CI gate, run export, failure attribution, and verification report evidence | `quickstart/self-bootstrap` | `go test -count=1 ./quickstart/self-bootstrap` | - |
| A2A child agent as typed graph node with nested evidence | `quickstart/agent-node` | `go test -count=1 ./quickstart/agent-node` | - |
| A2A local cluster + multi-source discovery + tag route + fallback + cancel | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| A2A env mesh sync with mesh-level HTTP options and readiness pruning | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| A2A continuous env mesh sync with registry changes | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| A2A local cluster expiry-aware discovery and lease heartbeat evidence | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| A2A local cluster readiness-gated endpoint discovery | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| A2A local cluster run export golden trajectory | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| A2A local cluster policy deny and review | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| A2A local cluster retry evidence | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| Dev Agent test and review evidence | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| Dev Agent replay and command evidence | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| OpenAI-compatible chat | `quickstart/openai-chat` | `go test -count=1 ./quickstart/openai-chat` | - |
| OpenAI-compatible streaming | `quickstart/openai-streaming` | `go test -count=1 ./quickstart/openai-streaming` | - |
| tool calling | `quickstart/tool-calling` | `go test -count=1 ./quickstart/tool-calling` | - |
| structured output | `quickstart/structured-output` | `go test -count=1 ./quickstart/structured-output` | - |
| Ark SDK provider | `quickstart/ark-chat` | `go test -count=1 ./quickstart/ark-chat` | - |
| Ark OpenAI-compatible streaming | `quickstart/ark-streaming` | `go test -count=1 ./quickstart/ark-streaming` | - |
| Agnes provider | `quickstart/agnes-chat` | `go test -count=1 ./quickstart/agnes-chat` | `go test -tags=integration -count=1 ./quickstart/agnes-chat` |

覆盖原则：

- 每个 quickstart 都必须有 `main.go`、`main_test.go` 和 `README.md`。
- provider quickstart 的默认测试必须使用 fake server 或 mock 数据。
- 真实 provider 凭据只能通过 `.env` 或显式环境变量进入本地 integration 测试。
