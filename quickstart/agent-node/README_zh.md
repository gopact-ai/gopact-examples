# Agent Node

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

这个示例演示如何通过 `agentnode.New` 把 A2A 子 agent 挂成 typed graph node。示例使用本地 scripted A2A agent，不需要真实 provider credential。

```bash
go run ./quickstart/agent-node
```

它覆盖：

- 将 graph state 映射为 `a2a.Task`。
- 将 A2A streaming status 和 completion event 注入父 graph run。
- 将终态 `a2a.Result` 映射回 typed workflow state。
- 在父 graph event 上保留子 agent evidence。

