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
go run ./quickstart/workflow-graph
go run ./quickstart/react-agent
go run ./quickstart/plan-exec
go run ./quickstart/openai-chat
go run ./quickstart/openai-streaming
go run ./quickstart/ark-chat
go run ./quickstart/ark-streaming
go run ./quickstart/tool-calling
```

Tests use local fake LLM servers, so CI does not need real credentials.
