# self-bootstrap

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

这个 quickstart 使用 `gopact-ext/devagent/selfbootstrap` 和 `gopact-ext/devagent/workspace` 运行一条无凭据 Dev Agent self-bootstrap workflow。示例会创建临时 git 仓库，由 planner 产出 patch proposal，经本地 policy 授权后通过 workspace adapter 应用 approved plan patch，采集 repo-relative worktree diff 与 file snapshot，在临时 workspace 内执行 `go test ./...`，并输出 run export 和 verification evidence 摘要。

```bash
go run ./quickstart/self-bootstrap
```

示例不会调用模型，也不会修改当前 examples 仓库。plan patch apply 和命令执行都发生在临时 workspace 中，运行结束后会删除。
