# ReAct Agent

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: zh,en -->

## 中文

这个示例通过 `gopact-ext/agents/react` 运行一个 ReAct-style model/tool loop。它使用 scripted local model 和本地 `uppercase` tool，因此可以在 CI 中无凭据运行。

```bash
go run ./quickstart/react-agent
```

执行过程：

1. scripted model 请求调用 `local.uppercase`。
2. tool registry 执行工具并返回结果。
3. model 生成最终回答。
4. 示例打印事件链、tool result 和 final answer。

## English

This example runs a ReAct-style model/tool loop with a scripted local model and a local uppercase tool. It is the recommended first no-credential quickstart.
