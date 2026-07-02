# Security Policy

<!-- gopact:doc-language: zh -->

[英文文档](./SECURITY.md)

## 中文

`gopact-examples` 展示 provider、tool、A2A agent 和工程证据的完整调用路径，因此安全要求重点放在凭据隔离和可公开日志上。默认示例必须能在没有真实 provider credential 的 CI 中运行。

## Supported Versions

`gopact-examples` 跟随 `main` 分支以及 `go.mod` 中声明的最新 `gopact` / `gopact-ext` 版本。仓库进入稳定版本线后，本节会改为明确的支持版本表。

## Reporting a Vulnerability

不要为疑似漏洞创建公开 issue。请通过 `gopact-ai` 组织维护者私有渠道报告，直到仓库启用 GitHub Security Advisory 流程。

报告时请包含：

- 受影响的 quickstart 或 internal package。
- 最小复现步骤。
- 影响边界：provider token、prompt、tool args/result、artifact、A2A event、本地文件或用户数据。
- 是否已在 fork、CI log、issue、PR 评论或 commit message 中暴露敏感信息。

处理要求：

- `.env` 必须保持本地文件，`.env.example` 只能包含占位值。
- CI 不读取 `.env`，不要求真实 provider credential。
- public readiness check 必须扫描 tracked file 和 commit message 中的高置信敏感模式。
- 示例输出不得打印真实 token、原始密钥、完整私有 prompt 或客户数据。
