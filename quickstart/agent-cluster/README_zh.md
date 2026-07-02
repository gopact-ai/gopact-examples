# Agent Cluster

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

这个示例在本地拉起一个 A2A-style agent cluster，不需要外部服务或模型凭据。它展示 planner、research、code、review 四个垂域 agent 如何通过 `a2a.Mesh` 自动发现、路由和协作。

```bash
go run ./quickstart/agent-cluster
```

默认模式会创建临时本地 agent-card registry。也可以通过环境变量切换 discovery 来源：

```bash
GOPACT_A2A_REGISTRY_FILE=./agents.json
GOPACT_A2A_REGISTRY_URL=http://localhost:8080/agents.json
GOPACT_A2A_ENDPOINTS=http://localhost:8080,http://localhost:8081
```

如果多个 discovery 变量同时存在，示例按 file、registry URL、endpoint 的顺序加载并合并。

示例覆盖：

- multi-source A2A discovery、tag route、fallback 和 readiness-gated endpoint discovery。
- checkpoint、resume、policy allow/deny/review、retry evidence、cancel evidence。
- `RunExport` golden trajectory。
- git diff、file snapshot、dev-agent replay 和 command evidence。
- self-bootstrap release gate bundle。
