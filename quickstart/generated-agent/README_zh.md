# Generated Agent Quickstart

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

## 中文

这个示例调用 core CLI 的 `gopact agent init`，在临时目录生成一个 A2A HTTP agent，然后执行生成项目的 `go mod tidy`、`go test ./...`，并通过 `gopact agent run` 拉起服务做 smoke test。

```bash
go run ./quickstart/generated-agent
```

它验证：

- 当前 core SDK 版本可以生成可编译 agent 项目。
- 生成项目包含 `README.md`、`.env.example`、`.gitignore`、`agents.json`、`main.go`、`main_test.go`。
- 生成项目的 README 包含 `gopact agent run .` 用法。
- 生成的 HTTP agent 暴露 `/agents.json` registry endpoint。
