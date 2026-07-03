# self-bootstrap

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/gopact-ai/gopact-examples.svg)](https://pkg.go.dev/github.com/gopact-ai/gopact-examples)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: en -->

This quickstart runs a credential-free Dev Agent self-bootstrap workflow using `gopact-ext/devagent/selfbootstrap`. It injects local analyze, plan, write, test, and review stages, then prints the run export and verification evidence summary.

```bash
go run ./quickstart/self-bootstrap
```

The example does not call a model, execute commands, or modify the workspace. It demonstrates how hosts pass already-observed diff, file snapshot, command, CI gate, and review results into the reusable self-bootstrap workflow.
