# ReAct Agent

Run a ReAct-style model/tool loop through `gopact-ext/agents/react`.

This example uses a scripted local model so it can run in CI without credentials. Real applications can inject any `gopact.ChatModel`, including the OpenAI adapter from `gopact-ext/models/openai`.

```bash
go run ./quickstart/react-agent
```
