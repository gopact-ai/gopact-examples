# Plan-Exec

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)


<!-- gopact:doc-language: zh,en -->

## 中文

本文档是 gopact 开源文档集的一部分，中文内容用于说明当前仓库约束、能力或维护流程。

## English

This document is part of the gopact open-source documentation set. The English section gives an entry point for readers who prefer English, while the remaining sections preserve the maintained technical details.


Run a Plan-Execute workflow through `gopact-ext/agents/planexec`.

This example uses local planner/executor functions. A real application can make the planner or executor model-backed while keeping the same graph event stream.

The runnable example covers one-shot replan after an execution failure. The test suite also covers approval resume and cancel propagation so the example stays aligned with production workflow behavior.

```bash
go run ./quickstart/plan-exec
```
