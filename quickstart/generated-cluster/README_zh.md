# Generated Cluster

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)

<!-- gopact:doc-language: zh -->

[英文文档](./README.md)

这个示例调用 core CLI 的 `gopact agent init-cluster`，在临时目录生成本地 A2A HTTP cluster，然后通过 `gopact agent verify` 校验 scaffold，并通过 `gopact agent run` 拉起服务做 smoke test。

```bash
go run ./quickstart/generated-cluster
```

生成项目会暴露 `/agents.json`、planner/worker/reviewer A2A endpoint、health/readiness endpoint，并包含 mesh bootstrap、`GOPACT_A2A_REGISTRY_URL` registry override、routing、streaming、cancel 和 shutdown 测试。
