# Plan-Exec

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

这个示例通过 `gopact-ext/agents/planexec` 运行一个本地 Plan-Execute workflow。planner、executor 和 replanner 都是本地函数，不需要真实 provider。

```bash
go run ./quickstart/plan-exec
```

示例主流程覆盖一次执行失败后的 replan。测试还覆盖 approval resume 和 cancel propagation，确保 quickstart 不是只有 happy path。

真实应用可以把 planner 或 executor 替换成 `gopact.ResponseModel`，同时保留相同的 event stream 和 runtime IDs。
