# Agent Scaffold

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](../../LICENSE)

<!-- gopact:doc-language: zh,en -->

## 中文

这个示例是最小的无凭据 agent scaffold：typed graph、checkpoint、approval interrupt、resume、verification report 和 A2A file registry 都在一个本地流程里完成。它是进入 `quickstart/agent-cluster` self-bootstrap release gate 路径前的最小骨架。

```bash
go run ./quickstart/agent-scaffold
```

它适合作为接入真实 provider 前的骨架：先把工作流边界、人工审批、恢复点和验证报告跑通，再把模型调用替换进具体节点。

验证点：

- 第一次运行在 approval 节点中断。
- resume 使用 checkpoint 和 approval payload 继续执行。
- `RunExport` 被写入 verification report，并嵌入 release bundle；the example embeds the report in a release bundle。
- scaffold agent card 被写入本地 A2A registry。

## English

This is the smallest no-credential agent scaffold: typed graph, checkpointing, approval interrupt/resume, verification report, and A2A file registry in one local flow before the larger self-bootstrap release gate path in `quickstart/agent-cluster`.
