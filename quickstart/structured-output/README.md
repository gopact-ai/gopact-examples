# Structured Output Quickstart

Run an OpenAI-compatible chat completion with a JSON schema response contract.

## Configure

Create `.env` at the repository root:

```dotenv
GOPACT_LLM_BASEURL=https://api.openai.com/v1
GOPACT_LLM_TOKEN=your-token
GOPACT_LLM_MODEL=gpt-4o-mini
```

## Run

```bash
go run ./quickstart/structured-output
```
