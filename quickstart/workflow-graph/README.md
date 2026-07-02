# Workflow Graph

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)


<!-- gopact:doc-language: zh,en -->

## 中文

本文档是 gopact 开源文档集的一部分，中文内容用于说明当前仓库约束、能力或维护流程。

## English

This document is part of the gopact open-source documentation set. The English section gives an entry point for readers who prefer English, while the remaining sections preserve the maintained technical details.


Run a typed workflow with `github.com/gopact-ai/gopact/graph`.

This example does not call a model. It shows the workflow runtime shape that agent templates build on: typed state, named nodes, dynamic branch fan-out, DAG fan-in, runnable subgraphs with nested events, branch-driven loops, step-limit guarding, and event streaming at each step boundary.

```bash
go run ./quickstart/workflow-graph
```
