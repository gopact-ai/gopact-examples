# Plan-Exec

Run a Plan-Execute workflow through `gopact-ext/agents/planexec`.

This example uses local planner/executor functions. A real application can make the planner or executor model-backed while keeping the same graph event stream.

The runnable example covers one-shot replan after an execution failure. The test suite also covers approval resume and cancel propagation so the example stays aligned with production workflow behavior.

```bash
go run ./quickstart/plan-exec
```
