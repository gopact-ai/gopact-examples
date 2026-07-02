# Contributing to gopact-examples

`gopact-examples` contains runnable examples for `gopact` and official
extensions. Examples should be small, reproducible, and safe to run without
real provider credentials unless explicitly marked as integration tests.

## Development Setup

Prerequisites:

- Go 1.25.11
- Git
- `golangci-lint` v2.8.0
- `govulncheck` v1.1.4

Clone and verify the repository:

```bash
git clone git@github.com:gopact-ai/gopact-examples.git
cd gopact-examples
go test -count=1 ./...
```

## Change Guidelines

- Keep CI mock-only. Real provider calls must stay behind the `integration`
  build tag.
- Prefer fake local model servers for examples that demonstrate provider-shaped
  behavior.
- Document required environment variables in the example README and
  `.env.example`.
- Do not commit `.env`, real API keys, model endpoint IDs, prompts, or raw
  provider responses.

## Verification

Before opening a pull request, run:

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

- The example can be run from the repository root.
- Tests cover the documented command path.
- Provider-backed checks are opt-in with integration tags.
- No generated noise, local `.env`, raw prompts, API keys, or endpoint IDs are
  tracked.
