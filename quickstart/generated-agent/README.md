# Generated Agent Quickstart

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: zh,en -->

## 中文

这个示例调用 core CLI 的 `gopact agent init`，在临时目录生成一个 A2A HTTP agent，然后执行生成项目的 `go mod tidy` 和 `go test ./...`。

```bash
go run ./quickstart/generated-agent
```

它验证：

- 当前 core SDK 版本可以生成可编译 agent 项目。
- 生成项目包含 `README.md`、`.env.example`、`.gitignore`、`agents.json`、`main.go`、`main_test.go`。
- 生成项目的 README 包含 `gopact agent run .` 用法。
- 生成的 HTTP agent 暴露 `/agents.json` registry endpoint。

## English

This example runs the core `gopact agent init` generator, then verifies that the generated A2A HTTP agent project can tidy and test successfully.
