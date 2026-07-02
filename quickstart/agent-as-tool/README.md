# Agent as Tool

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)


<!-- gopact:doc-language: zh,en -->

## 中文

本文档是 gopact 开源文档集的一部分，中文内容用于说明当前仓库约束、能力或维护流程。

## English

This document is part of the gopact open-source documentation set. The English section gives an entry point for readers who prefer English, while the remaining sections preserve the maintained technical details.


Run a parent ReAct agent that delegates one task to a Plan-Exec child agent exposed as a normal tool.

```bash
go run ./quickstart/agent-as-tool
```

This example uses only scripted local models. It demonstrates `agenttool.New`, `a2a.NewRunnableAgent`, child A2A completion evidence, failure evidence propagation, and runtime identity propagation without external credentials.
