# Ark Streaming

Stream an Ark OpenAI-compatible Responses call through `gopact-ext/models/openai`.
The example disables Ark thinking so small `max_output_tokens` budgets produce visible text deltas instead of reasoning-only output.

```bash
GOPACT_LLM_BASEURL=https://ark.cn-beijing.volces.com/api/v3 \
GOPACT_LLM_TOKEN=your-ark-api-key \
GOPACT_LLM_MODEL=ep-20260624181107-glhd6 \
go run ./quickstart/ark-streaming
```
