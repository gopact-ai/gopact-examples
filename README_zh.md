# 🧪 gopact-examples

<!-- gopact:doc-language: zh -->

[English documentation](README.md)

当前 `gopact` 新 API 的可运行示例仓库。

> **仅支持 Go 1.27+。** 本项目围绕泛型方法构建，也借此庆祝我们眼中 Go 近十年来最具影响力的语言演进之一。Go 1.27 正式发布前，本项目需要开发版工具链，应视为预览而非稳定版本。

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
| [`concepts/session-correlation`](./concepts/session-correlation) | 用 Session 关联多个独立 Run，再检查并恢复选中的 Run |

Session 查询用于列出相关 Run。Snapshot 和恢复操作必须用 `RunID` 选择具体 Run；不存在 Session Snapshot。共享 `workflow.MemoryStore` 仅保存进程生命周期内的执行检查点和日志记录，不是语义 Memory。

## 集成

| 示例 | 你将学到什么 |
| --- | --- |
| [`integrations/otel`](./integrations/otel) | 把 Workflow 身份和事件映射到调用方拥有的 OpenTelemetry span |
| [`integrations/mem0`](./integrations/mem0) | 显式检索语义 Memory，并由业务代码构造 Agent Context |

## OpenTelemetry 集成

应用已经拥有 OpenTelemetry 配置，并希望把 Workflow 事件关联到当前 span 时，可使用 `integrations/otel`。示例把 `SessionID` 映射为 `gen_ai.conversation.id`，把 `RunID` 映射为 `gopact.run.id`，把 Workflow definition ID 映射为 `gopact.workflow.name`。

这种方式不会把遥测身份写入领域 Event 或存储 schema，core 也无需依赖 OpenTelemetry，并可接入任意 SDK exporter。限制是 adapter 只能增强调用 `context.Context` 中的有效 span；没有有效 span 时，OpenTelemetry API 只执行空操作。

## Mem0 集成

Agent 需要从 Mem0 或兼容 HTTP 服务获取语义 Memory 时，可使用 `integrations/mem0`。示例用显式 typed topology 解决检索和作用域映射：

```text
load-memory（HTTP I/O）-> build-model-request（纯函数）-> model
```

决定模型能看到什么的 Agent Context 由业务构造；Memory 只是 Context 的一个输入，不是框架拥有的容器或 provider 接口。

检索节点通过 `workflow.RunInfoFromContext` 读取 SessionID 与 Workflow RunID，业务 Context 不重复保存执行 meta。调用方用 RunOptions 指定身份，框架再把最终身份传播给节点。

| 应用身份 | Mem0 / model 映射 |
| --- | --- |
| UserID | `user_id` |
| Agent identity | `agent_id` |
| SessionID | Mem0 `run_id` |
| Workflow RunID | 写入 `ModelRequest.Metadata` 的 `gopact.workflow.run_id`，仅用于 provenance |

优点：Workflow 中能直接看到 I/O 边界，provider 策略保留在业务代码里，core 和 ext 不引入 Mem0 依赖。缺点与限制：结果选择、prompt 构造、HTTP 兼容和失败策略都由业务负责；这个最小 client 只演示一次 `POST /search` 合约，不是完整 Mem0 SDK。为了避免 API key 泄露，它拒绝所有重定向（包括同源重定向），应直接配置最终 endpoint URL。

确定性示例使用离线结果。若要运行有 15 秒超时的外部 smoke test，可先加载仓库根目录的本地 `.env`：

```bash
set -a; [ ! -f .env ] || . ./.env; set +a
MEM0_INTEGRATION=1 go test -tags=integration ./integrations/mem0 -run TestMem0Smoke -count=1 -v
```

`MEM0_BASE_URL` 缺省为 `http://localhost:8888`，`MEM0_API_KEY` 可选。

## 运行全部示例

```bash
go test ./...
```
