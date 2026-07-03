# Workflow Graph

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

这个示例直接使用 `github.com/gopact-ai/gopact/graph` 构造 typed workflow，不调用模型。它展示 agent template 底层依赖的 graph runtime 形态。

```bash
go run ./quickstart/workflow-graph
```

覆盖能力：

- typed state 和 named nodes。
- dynamic branch fan-out。
- DAG fan-in。
- runnable subgraph 和 nested events。
- branch-driven loop。
- step-limit guard。
- completed step export/import resume。
- interrupted step export/import resume。
- step boundary event streaming。
