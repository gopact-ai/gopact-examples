# 🧪 gopact-examples

<!-- gopact:doc-language: zh -->

[英文文档](README.md)

这是基于新版 `gopact` API 的可运行示例仓库。

> **需要 Go 1.27 或更新版本。** 这些示例使用了仅在该工具链中提供的泛型方法。

下面的本地命令使用你安装的 Go 1.27 工具链。CI 使用 `setup-go` 固定到某个具体的 Go 1.27 工具链版本。

手动触发源码端到端流程时，必须为三个仓库分别传入经过审查的 40 位
提交 SHA。流程会检出这些精确提交，打印对应 SHA，再用临时 Go 工作区
联调。日常模块测试会在关闭 Go 工作区（`GOWORK=off`）的情况下使用本
模块声明的不可变稳定版本；另一项必需的源码兼容性检查则单独验证三个
仓库相互配套的源码。

发布顺序固定为 `gopact` → `gopact-ext` 各模块 → `gopact-examples`。
本模块锁定经过审核的不可变依赖版本；创建自己的发布标签前，必须在
`GOWORK=off` 下通过测试。

所有示例默认都能离线运行。

## 快速入门

| 示例 | 你将学到什么 |
| --- | --- |
| [`quickstart/model-basic`](./quickstart/model-basic) | 实现并调用 `gopact.Model` 的最小接口 |
| [`quickstart/workflow-basic`](./quickstart/workflow-basic) | 构建并运行类型明确、能发出可观测事件的 Workflow |
| [`quickstart/react-basic`](./quickstart/react-basic) | 通过 ReAct Agent 连接模型与工具 |

## 核心概念

| 示例 | 你将学到什么 |
| --- | --- |
| [`concepts/durable-resume`](./concepts/durable-resume) | 从检查点恢复一个被中断的 Run |
| [`concepts/run-control`](./concepts/run-control) | 通过 `Retry` 或 `Fork` 从失败的 Run 创建新 Run，并保留来源 |
| [`concepts/session-correlation`](./concepts/session-correlation) | 用 Session 关联多个独立 Run，再检查并恢复指定 Run |

`durable-resume` 只演示公开的中断与恢复路径。新进程解码、栅栏保护、崩溃窗口和副作用幂等性由核心库及 Store 集成测试负责，不在概念示例里重复搭建。为了离线运行，示例使用 `MemoryStore`；需要跨进程恢复时，必须换成持久化 Store。

`run-control` 保持失败的源 Run 不可变。`Retry` 把失败节点的一次执行重放到一个新 Run 中；`Fork` 从可安全重放的根输入启动另一个新 Run，并修改 Workflow 输入。两个新 Run 都通过 `SourceRunID` 保留来源。

按 Session 查询可以列出相关 Run。读取快照和恢复时，必须用 `RunID` 选择具体 Run；不存在 Session 级快照。共享的 `workflow.MemoryStore` 只保存当前进程生命周期内的执行检查点和日志记录，不是语义 Memory，也只适合测试或短生命周期进程。SQLite 适合单机，或能够安全共享同一个本地数据库文件的多进程；多主机部署必须使用支持原子 `Claim` 和栅栏保护的分布式数据库 Store。

## 集成

| 示例 | 你将学到什么 |
| --- | --- |
| [`integrations/otel`](./integrations/otel) | 把 Workflow 身份和事件映射到由调用方管理的 OpenTelemetry span |
| [`integrations/mem0`](./integrations/mem0) | 显式检索语义 Memory，并由业务代码构造 Agent 上下文 |

## OpenTelemetry 集成

应用已经自行配置 OpenTelemetry 时，可使用 `integrations/otel`。示例把 Workflow 领域事件投影到运行 span：`SessionID` 映射为 `gen_ai.conversation.id`，`RunID` 映射为 `gopact.run.id`，Workflow 定义 ID 映射为 `gopact.workflow.name`。应用适配器另由基础设施 span 包装，这个 span 不会伪造 Workflow 事件。

这种方式不会把遥测身份写入领域事件或存储结构，核心库也无需依赖 OpenTelemetry，并可接入任意 SDK 导出器。两种投影都沿用调用方的 `context.Context`；使用 OpenTelemetry 的 no-op provider 时不会产生运行时遥测。

## Mem0 集成

Agent 需要从 Mem0 或兼容 HTTP 服务获取语义 Memory 时，可使用 `integrations/mem0`。示例通过显式、类型明确的数据流完成检索和作用域映射：

```text
load-memory（加载 Memory，HTTP I/O）-> build-model-request（构造模型请求，纯函数）-> model（调用模型）
```

决定模型能看到什么的 Agent 上下文由业务代码构造。Memory 只是上下文的一项输入，不是框架持有的容器或 provider 接口。检索到的 Memory 来自外部，可能变化，也不可信。示例只把固定的应用策略放入 `system` 角色；召回的 Memory 放在一条独立的 `user` 角色消息中，并明确标为不可信证据；当前用户消息放在最后。普通 Workflow 输入只携带用户文本，不接受带角色的模型消息，因此请求数据不能自行选择更高信任级别。Memory 也绝不会被提升为 `system` 指令。

检索节点通过 `workflow.RunInfoFromContext` 读取 `SessionID` 和 Workflow `RunID`，业务上下文不重复保存执行元数据。用户和 Agent 身份由应用持有，存放在 `memoryWorkflowConfig` 中，不进入普通的 `workflowInput`。演示程序使用固定值；真实服务必须从经过认证的服务端状态取得这些值，不能直接复制普通请求体中的字段。应用还应在确定调用方作用域后，通过 `gopact.WithSessionID` 分配 `SessionID`。

| 应用身份 | Mem0 / 模型映射 |
| --- | --- |
| 应用配置中经过认证的 `UserID` | `user_id` |
| 由应用管理的 Agent 身份 | `agent_id` |
| 应用分配的 `SessionID` | Mem0 `run_id` |
| Workflow `RunID` | 写入 `ModelRequest.Metadata` 的 `gopact.workflow.run_id`，仅用于追溯来源 |

优点：Workflow 中能清楚看到 I/O 边界，Mem0 的使用策略留在业务代码里，`gopact` 和 `gopact-ext` 不引入 Mem0 依赖。限制：角色隔离只是纵深防御，既不足以完整防住提示词注入，也不能代替权限校验。应用仍需负责身份认证、作用域授权、结果筛选、来源验证、排序、提示词构造、HTTP 兼容和失败处理策略。这个最小客户端只演示 `POST /search` 的调用契约，不是完整的 Mem0 SDK。为了避免 API 密钥泄露，它拒绝所有重定向，包括同源重定向；请直接配置最终服务地址。

确定性示例使用离线结果。若要运行带有 15 秒超时的外部冒烟测试，可先加载仓库根目录的本地 `.env`：

```bash
set -a; [ ! -f .env ] || . ./.env; set +a
MEM0_INTEGRATION=1 go test -tags=integration ./integrations/mem0 -run TestMem0Smoke -count=1 -v
```

`MEM0_BASE_URL` 默认为 `http://localhost:8888`，`MEM0_API_KEY` 可选。

## 运行全部示例

检出发布版本后执行：

```bash
GOWORK=off go mod download
GOWORK=off go test -count=1 ./...
```

创建发布标签前，源码端到端流程会检出三个相互配套的仓库版本，建立临时 Go 工作区，然后执行：

```bash
go test ./...
```
