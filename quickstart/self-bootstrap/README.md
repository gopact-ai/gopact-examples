# self-bootstrap

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/gopact-ai/gopact-examples.svg)](https://pkg.go.dev/github.com/gopact-ai/gopact-examples)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: en -->

This quickstart runs a credential-free Dev Agent self-bootstrap workflow using `gopact-ext/devagent/selfbootstrap` and `gopact-ext/devagent/workspace`. It creates a temporary git repository, applies a caller-provided unified diff through the workspace adapter, captures a repo-relative worktree diff and file snapshot, executes `go test ./...` inside that temporary workspace, and prints the run export and verification evidence summary.

```bash
go run ./quickstart/self-bootstrap
```

The example does not call a model or modify the checked-out examples repository. Patch apply and command activity stay inside a temporary workspace that is deleted after the run.
