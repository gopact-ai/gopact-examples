# Agnes Chat Quickstart

Run a single chat completion through `gopact-ext/models/agnes`.

## Configure

Create `.env` at the repository root:

```dotenv
GOPACT_LLM_BASEURL=https://apihub.agnes-ai.com/v1
GOPACT_LLM_TOKEN=your-agnes-token
GOPACT_LLM_MODEL=agnes-2.0-flash
```

## Run

```bash
go run ./quickstart/agnes-chat
```
