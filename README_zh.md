# 🧪 gopact-examples

<!-- gopact:doc-language: zh -->

[English documentation](README.md)

当前 `gopact` 新 API 的可运行示例仓库。

> **仅支持 Go 1.27+。** 本项目围绕泛型方法构建，也借此庆祝我们眼中 Go 近十年来最具影响力的语言演进之一。Go 1.27 正式发布前，本项目需要开发版工具链，应视为预览而非稳定版本。

协调发布各 RC 模块之前，手动 source E2E workflow 要求传入已 review 的 core 与 ext
40 位 commit SHA，checkout 对应精确提交，打印三个仓库的 SHA，再用临时 Go workspace
联调。普通 CI 使用 `GOWORK=off` 消费 examples module 声明的 immutable versions。

发布顺序固定为 core → 两个 ext module → examples。获批的 immutable dependency tags 发布后，本 module 必须固定这些精确版本，并在 `GOWORK=off` 下通过。该 post-tag 门禁目前尚未通过；Go 1.27 stable 验证和 RC burn-in 完成前，RC 只能称为 production evaluation candidate。

当前示例默认全部离线可运行。

## Quickstart

| 示例 | 你将学到什么 |
| --- | --- |
| [`quickstart/model-basic`](./quickstart/model-basic) | 实现并调用最小的 `gopact.Model` contract |
| [`quickstart/workflow-basic`](./quickstart/workflow-basic) | 构建带可观测事件的 typed Workflow |
| [`quickstart/react-basic`](./quickstart/react-basic) | 通过 ReAct Agent 连接 model 与 tool |

## 核心概念

| 示例 | 你将学到什么 |
| --- | --- |
| [`concepts/durable-resume`](./concepts/durable-resume) | 从 checkpoint 恢复一个 interrupted Run |
| [`concepts/run-control`](./concepts/run-control) | 把 failed Run Retry 或 Fork 为带 source lineage 的新 Run |
| [`concepts/session-correlation`](./concepts/session-correlation) | 用 Session 关联多个独立 Run，再检查并恢复选中的 Run |

durable-resume 只保留公开的 interrupt/resume 主路径。fresh-process decode、fencing、crash window 与副作用幂等由 core 和 Store integration suites 验证，不在概念示例中重复搭建。示例为离线运行使用 `MemoryStore`；需要跨进程恢复时必须替换成 durable Store。

run-control 保持 failed source Run 不可变。`Retry` 把一次失败的 node activation 重放到新 Run，`Fork` 则从 replay-safe root 创建另一个新 Run并修改 workflow input；两个新 Run 都保留 `SourceRunID` lineage。

Session 查询用于列出相关 Run。Snapshot 和恢复操作必须用 `RunID` 选择具体 Run；不存在 Session Snapshot。共享 `workflow.MemoryStore` 仅保存进程生命周期内的执行检查点和日志记录，不是语义 Memory，也只适合测试或短生命周期进程。SQLite 适用于单机，或安全共享同一个本地数据库文件的多进程；多主机必须使用支持原子 Claim 与 fencing 的分布式数据库 Store。

## 集成

| 示例 | 你将学到什么 |
| --- | --- |
| [`integrations/otel`](./integrations/otel) | 把 Workflow 身份和事件映射到调用方拥有的 OpenTelemetry span |
| [`integrations/mem0`](./integrations/mem0) | 显式检索语义 Memory，并由业务代码构造 Agent Context |

## OpenTelemetry 集成

应用已经拥有 OpenTelemetry 配置时，可使用 `integrations/otel`。示例把 Workflow domain Events 投影到 run span：`SessionID` 映射为 `gen_ai.conversation.id`，`RunID` 映射为 `gopact.run.id`，Workflow definition ID 映射为 `gopact.workflow.name`。应用 adapter 另由 infrastructure span 包装，这个 span 不会伪造 Workflow Event。

这种方式不会把遥测身份写入领域 Event 或存储 schema，core 也无需依赖 OpenTelemetry，并可接入任意 SDK exporter。两种投影都沿用调用方的 `context.Context`；使用 OpenTelemetry no-op provider 时不会产生运行时遥测。

## Mem0 集成

Agent 需要从 Mem0 或兼容 HTTP 服务获取语义 Memory 时，可使用 `integrations/mem0`。示例用显式 typed topology 解决检索和作用域映射：

```text
load-memory（HTTP I/O）-> build-model-request（纯函数）-> model
```

决定模型能看到什么的 Agent Context 由业务构造；Memory 只是 Context 的一个输入，不是框架拥有的容器或 provider 接口。检索到的 Memory 是外部、可变且不可信的数据。示例只把固定的 application policy 放入 system role，把 recalled Memory 放入单独的 user-role message 并标记为 untrusted evidence，最后再放当前 user message。普通 workflow input 只携带 user text，而不是可指定 role 的 model message，因此请求数据不能自行选择更高信任级别；Memory 永远不会被提升成 system instruction。

检索节点通过 `workflow.RunInfoFromContext` 读取 SessionID 与 Workflow RunID，业务 Context 不重复保存执行 meta。User 与 Agent identity 位于 application-owned `memoryWorkflowConfig`，不进入普通 `workflowInput`。可运行 demo 使用固定值；真实服务必须从认证后的服务端状态取得这些值，不能复制普通请求 body 字段。application 也应在确定调用方 scope 后通过 RunOptions 分配 SessionID。

| 应用身份 | Mem0 / model 映射 |
| --- | --- |
| application config 中已认证的 UserID | `user_id` |
| application-owned Agent identity | `agent_id` |
| application 分配的 SessionID | Mem0 `run_id` |
| Workflow RunID | 写入 `ModelRequest.Metadata` 的 `gopact.workflow.run_id`，仅用于 provenance |

优点：Workflow 中能直接看到 I/O 边界，provider 策略保留在业务代码里，core 和 ext 不引入 Mem0 依赖。缺点与限制：role separation 只是纵深防御，不是完整的 prompt-injection 防护或 authorization。identity authentication、scope authorization、结果选择、provenance 验证、ranking、prompt 构造、HTTP 兼容和失败策略仍由业务负责；这个最小 client 只演示一次 `POST /search` 合约，不是完整 Mem0 SDK。为了避免 API key 泄露，它拒绝所有重定向（包括同源重定向），应直接配置最终 endpoint URL。

确定性示例使用离线结果。若要运行有 15 秒超时的外部 smoke test，可先加载仓库根目录的本地 `.env`：

```bash
set -a; [ ! -f .env ] || . ./.env; set +a
MEM0_INTEGRATION=1 go test -tags=integration ./integrations/mem0 -run TestMem0Smoke -count=1 -v
```

`MEM0_BASE_URL` 缺省为 `http://localhost:8888`，`MEM0_API_KEY` 可选。

## 运行全部示例

发布后的 checkout 使用：

```bash
GOWORK=off go mod download
GOWORK=off go test -count=1 ./...
```

tag 前的 source E2E 则对三个协调 source checkout 创建临时 workspace，并执行：

```bash
go test ./...
```
