# Ark Streaming

Stream an Ark OpenAI-compatible Responses call through `gopact-ext/models/openai`.
The example runs both `thinking=disabled` and `thinking=enabled` with `max_output_tokens=1024`.
Reasoning tokens count against the output budget, so tiny values like `64` can finish with reasoning-only output on thinking models.

```bash
GOPACT_LLM_BASEURL=https://ark.cn-beijing.volces.com/api/v3 \
GOPACT_LLM_TOKEN=your-ark-api-key \
GOPACT_LLM_MODEL=ep-20260624181107-glhd6 \
go run ./quickstart/ark-streaming
```
