# Contributing to gopact-examples

<!-- gopact:doc-language: zh,en -->

## 中文

`gopact-examples` 的每个示例都必须能从仓库根目录运行，并且必须有测试固化文档中的命令路径。示例代码可以简洁，但不能依赖隐藏前置条件。

## Development Setup

前置工具：

- Go 1.25.11
- Git
- `golangci-lint` v2.8.0
- `govulncheck` v1.1.4

克隆后先执行：

```bash
git clone git@github.com:gopact-ai/gopact-examples.git
cd gopact-examples
go test -count=1 ./...
```

修改规则：

- CI 保持 mock-only。真实 provider 调用必须放在 `integration` build tag 下。
- provider 行为优先用本地 fake server 覆盖，确保无凭据也能验证 request/response 形态。
- 新增 quickstart 必须同时提供 `main.go`、`main_test.go`、`README.md`，并在根 README 与 [FEATURES.md](FEATURES.md) 登记。
- 新增环境变量必须同步更新 `.env.example`、根 README 和对应 quickstart README。
- 不提交 `.env`、真实 token、真实 endpoint ID、私有 prompt、原始模型响应或客户数据。

## Verification

提交 PR 前运行：

```bash
git diff --check
./scripts/public-readiness-check.sh
go mod tidy
git diff --exit-code
go test -count=1 ./...
go test -race -count=1 ./...
go vet ./...
golangci-lint run ./...
go test -coverprofile=coverage.out ./...
govulncheck ./...
```

## Pull Request Checklist

- 示例能从仓库根目录执行。
- README 中的 `go run ./quickstart/...` 命令和 `main_test.go` 覆盖一致。
- provider-backed 行为默认有 fake server/mock 测试，真实服务测试使用 integration tag。
- 新增或修改的环境变量已写入 `.env.example` 和相关 README。
- PR 不包含真实密钥、真实 endpoint ID、原始模型输出、私有 prompt 或用户数据。

## English

Every example in `gopact-examples` must run from the repository root and have tests that lock the documented command path. Keep CI mock-only, use fake servers for provider-shaped behavior, and reserve real provider checks for explicit integration tests.
