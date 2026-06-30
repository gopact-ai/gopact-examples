# Ark Chat

Run a single Ark Chat Completions call through `gopact-ext/models/ark` and the official Ark SDK.

```bash
GOPACT_ARK_API_KEY=your-ark-api-key \
GOPACT_ARK_MODEL=your-ark-endpoint-id \
go run ./quickstart/ark-chat
```

Optional: set `GOPACT_ARK_BASEURL`, `GOPACT_ARK_REGION`, or use `GOPACT_ARK_ACCESS_KEY` + `GOPACT_ARK_SECRET_KEY` instead of `GOPACT_ARK_API_KEY`.
