# Workflow Graph

Run a typed workflow with `github.com/gopact-ai/gopact/graph`.

This example does not call a model. It shows the workflow runtime shape that agent templates build on: typed state, named nodes, dynamic branch fan-out, DAG fan-in, runnable subgraphs with nested events, branch-driven loops, step-limit guarding, and event streaming at each step boundary.

```bash
go run ./quickstart/workflow-graph
```
