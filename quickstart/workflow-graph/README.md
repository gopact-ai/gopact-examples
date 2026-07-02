# Workflow Graph

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: zh,en -->

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
- step boundary event streaming。

## English

This example uses the core graph runtime directly. It demonstrates typed state, branch fan-out, fan-in, nested runnable subgraphs, loop control, step limits, and event streaming without calling a model.
