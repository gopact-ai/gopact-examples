# ReAct Agent

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)


<!-- gopact:doc-language: zh,en -->

## 中文

本文档是 gopact 开源文档集的一部分，中文内容用于说明当前仓库约束、能力或维护流程。

## English

This document is part of the gopact open-source documentation set. The English section gives an entry point for readers who prefer English, while the remaining sections preserve the maintained technical details.


Run a ReAct-style model/tool loop through `gopact-ext/agents/react`.

This example uses a scripted local model so it can run in CI without credentials. Real applications can inject any `gopact.ChatModel`, including the OpenAI adapter from `gopact-ext/models/openai`.

```bash
go run ./quickstart/react-agent
```
