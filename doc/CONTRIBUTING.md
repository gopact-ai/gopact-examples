# Contributing to gopact-examples

<!-- gopact:doc-language: en -->

Chinese documentation: [CONTRIBUTING_zh.md](CONTRIBUTING_zh.md)

Every example in `gopact-examples` must run from the repository root and have tests that lock the documented command path. Keep CI mock-only, use fake servers for provider-shaped behavior, and reserve real provider checks for explicit integration tests.

## Development Setup

Install Go 1.25 or newer, clone the repository, and work on a pull-request branch:

```bash
git clone https://github.com/gopact-ai/gopact-examples.git
cd gopact-examples
git switch -c your-change
```

Copy `.env.example` to `.env` only when running local provider checks. `.env` is ignored and must never be committed.

## Verification

Run the repository gates before opening a pull request:

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

Real provider checks are opt-in:

```bash
go test -tags=integration -count=1 ./quickstart/agnes-chat
```

## Pull Request Checklist

- Keep each quickstart runnable from the repository root.
- Keep default tests deterministic and credential-free.
- Update `README.md`, the quickstart README, and `doc/FEATURES.md` when adding or changing an example.
- Document all environment variables that a provider or A2A discovery path reads.
- Keep generated files, provider tokens, endpoint IDs, and local `.env` values out of git history.
