# Agent as Tool

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

这个示例演示父 ReAct agent 如何把一个 Plan-Execute 子 agent 当作普通 tool 调用。整个流程使用 scripted local model，不需要真实 provider credential。

```bash
go run ./quickstart/agent-as-tool
```

它覆盖：

- `a2a.NewRunnableAgent` 将 runnable agent 包装成 A2A agent。
- `agenttool.New` 将 A2A agent 转成 `gopact.ToolFunc`。
- 父 agent 通过 tool call 委托子 agent。
- 子 agent 的 completion evidence、failure evidence 和 runtime IDs 回传给父 run。
