# Background Scheduler

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

这个示例演示如何使用 `agents/scheduler` 执行带 lease 的后台任务。示例使用内存队列和内存 lease backend，不需要真实 provider credential，CI 中可以稳定运行。

```bash
go run ./quickstart/background-scheduler
```

它覆盖：

- 使用 leased worker 执行有界后台 drain。
- 失败任务 retry，并在下一次 attempt 完成。
- 永久失败任务进入 dead-letter。
- 记录 schedule verification evidence。
- 每次 worker pass 后释放 ownership lease。
