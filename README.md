# gopact-examples

Runnable examples for `github.com/gopact-ai/gopact` and official extensions.

## Configuration

Examples read these environment variables:

- `GOPACT_LLM_BASEURL`: OpenAI-shaped `/v1` API base URL.
- `GOPACT_LLM_TOKEN`: API token.
- `GOPACT_LLM_MODEL`: model name.

`quickstart/ark-chat` uses Ark SDK variables instead: `GOPACT_ARK_API_KEY` or `GOPACT_ARK_ACCESS_KEY` + `GOPACT_ARK_SECRET_KEY`, plus `GOPACT_ARK_MODEL`.

By default examples load a `.env` file from the current directory or a parent directory. `.env` is ignored by git.

```bash
cp .env.example .env
```

## Examples

```bash
go run ./quickstart/agent-as-tool
go run ./quickstart/agent-cluster
go run ./quickstart/agnes-chat
go run ./quickstart/ark-chat
go run ./quickstart/ark-streaming
go run ./quickstart/openai-chat
go run ./quickstart/openai-streaming
go run ./quickstart/plan-exec
go run ./quickstart/react-agent
go run ./quickstart/tool-calling
go run ./quickstart/workflow-graph
```

Tests use local fake LLM servers, so CI does not need real credentials.

## Local Integration

CI stays mock-only. To verify the provider-backed Agnes quickstart locally, put one of `GOPACT_AGNES_API_KEY`, `GOPACT_AGNES_SK`, or `GOPACT_LLM_TOKEN` in `.env`, then run:

```bash
go test -tags=integration -count=1 ./quickstart/agnes-chat
```

## Development

```bash
git diff --check
go mod tidy
git diff --exit-code
go test -count=1 ./...
go test -race -count=1 ./...
go vet ./...
golangci-lint run ./...
go test -coverprofile=coverage.out ./...
govulncheck ./...
```
