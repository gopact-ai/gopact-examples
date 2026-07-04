# Human Review

<!-- gopact:doc-language: zh -->

这个 quickstart 使用 `gopact-ext/agents/humanreview` 演示无凭据人工审批 gate。

```bash
go run ./quickstart/human-review
```

示例构造 `draft -> review -> publish` typed graph。`review` 节点产生 `humanreview` approval interrupt，然后分别通过 step export/import 和 checkpoint resume 恢复同一个审批 gate。
