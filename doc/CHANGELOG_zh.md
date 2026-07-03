# Changelog

<!-- gopact:doc-language: zh -->

[英文文档](./CHANGELOG.md)

## 中文

本文件记录 `gopact-examples` 对用户可见的变更。示例仓库跟随 `gopact` 和 `gopact-ext` 的当前发布版本，主要变更应体现在新增 quickstart、环境变量、测试覆盖和文档入口上。

## Unreleased

- 增加 `quickstart/background-scheduler`，无凭据演示带 lease 的后台任务、retry、dead-letter、drain 和 schedule verification evidence。
- workflow graph quickstart 增加 completed step export/import resume 和 interrupted resume 的无凭据演示。
- 示例仓库跟进 `gopact` core `v0.0.45` 和当前 `gopact-ext` 发布 tag。
- agent cluster quickstart 演示通过 mesh-level HTTP options 配置环境变量驱动的 A2A discovery。
- 增加 `quickstart/agent-node`，演示将 A2A 子 agent 挂成 typed graph node，并保留嵌套 evidence。
- agent cluster quickstart 覆盖 `Mesh.SyncEnv` 与 `Mesh.SyncEnvEvery`，演示环境变量驱动的 A2A discovery、HTTP agent 注册、readiness pruning 和 registry changes。
- agent cluster quickstart 演示 A2A lease heartbeat evidence、replay evidence 和 command evidence。
- 增加 `quickstart/supervisor`，无凭据演示 supervisor 路由到具名 Plan-Execute 子 agent。
- 重写根 README、quickstart README 和 `doc/` 文档，补齐示例定位、运行路径、环境变量、mock/integration 测试边界、安全和治理说明。
- 保持 CI mock-only，并继续通过 `go test -tags=integration -count=1 ./quickstart/agnes-chat` 支持本地 Agnes provider 验证。
- 明确无凭据 scaffold path：`react-agent`、`plan-exec`、`supervisor`、`agent-as-tool`、`agent-cluster`。

## 2026-07-02

- 增加 public readiness 检查，扫描 tracked file 和 commit message 中的高置信敏感信息模式。
- 增加 PR governance workflows：admin-authored PR 在必需门禁通过后自动 squash merge；non-admin-authored PR 需要至少一名 admin 审批。
- 将 CI 拆成 hygiene、unit、race、static、coverage、security 等独立 job，并保留聚合的 `ci/test` required check。
- 固化 quickstart 覆盖矩阵：dotenv、workflow graph、agent scaffold、generated agent、Plan-Execute、agent-as-tool、A2A cluster、OpenAI-compatible、Ark、Agnes、tool calling、structured output。
