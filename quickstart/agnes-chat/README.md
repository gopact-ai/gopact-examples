# Agnes Chat Quickstart

Run a single chat completion through `gopact-ext/models/agnes`.

## Configure

Create `.env` at the repository root:

```dotenv
GOPACT_LLM_BASEURL=https://apihub.agnes-ai.com/v1
GOPACT_LLM_TOKEN=your-agnes-token
GOPACT_LLM_MODEL=agnes-2.0-flash
```

You can also use Agnes-specific credentials:

```dotenv
GOPACT_AGNES_API_KEY=your-agnes-token
GOPACT_AGNES_SK=your-agnes-token
GOPACT_AGNES_MODEL=agnes-2.0-flash
```

## Run

```bash
go run ./quickstart/agnes-chat
```

## Local Integration

```bash
go test -tags=integration -count=1 ./quickstart/agnes-chat
```
