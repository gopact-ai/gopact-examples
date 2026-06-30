# OpenAI Streaming

Stream both OpenAI API surfaces through `gopact-ext/models/openai`.

## Configure

Create `.env` at the repository root:

```dotenv
GOPACT_LLM_BASEURL=https://api.openai.com/v1
GOPACT_LLM_TOKEN=your-token
GOPACT_LLM_MODEL=gpt-4o-mini
```

## Run

```bash
go run ./quickstart/openai-streaming
```

## Notes

- `openai.WithChatCompletionsAPI()` streams from `/chat/completions` and sends a `messages` payload. This is the common OpenAI-compatible chat API.
- `openai.WithResponsesAPI()` streams from `/responses` and sends an `input` payload. Use it when the provider supports the Responses API, richer content parts, or reasoning deltas.
