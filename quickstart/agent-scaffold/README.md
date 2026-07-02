# Agent Scaffold

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)


<!-- gopact:doc-language: zh,en -->

## 中文

本文档是 gopact 开源文档集的一部分，中文内容用于说明当前仓库约束、能力或维护流程。

## English

This document is part of the gopact open-source documentation set. The English section gives an entry point for readers who prefer English, while the remaining sections preserve the maintained technical details.


Minimal no-credential agent skeleton with graph nodes, checkpointing, approval interrupt, resume, and a verification report.

```bash
go run ./quickstart/agent-scaffold
```

This is the smallest local scaffold path before adding a real provider. It records a `RunExport`, builds a verification report, embeds the report into a bundle, and writes a local A2A file registry entry for the scaffold agent; use `quickstart/agent-cluster` when the same scaffold needs A2A task evidence and a self-bootstrap release gate bundle. Use `quickstart/react-agent`, `quickstart/plan-exec`, and provider quickstarts to expand it.
