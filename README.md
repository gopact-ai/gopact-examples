# gopact-examples

<!-- gopact:doc-language: en -->

Executable examples for the redesigned `gopact` API.

> **Go 1.27+ only.** This project is built around generic methods and celebrates what we see as one of Go's most consequential language changes of the past decade. Until Go 1.27 is officially released, it requires a development toolchain and should be treated as a preview, not a stable release.

Default example runs are intentionally offline:

- `quickstart/workflow-basic`
- `quickstart/model-basic`
- `quickstart/react-basic`
- [`concepts/session-correlation`](./concepts/session-correlation) — correlate independent Runs with a Session, then inspect and resume one selected Run
- [`integrations/otel`](./integrations/otel) — map Workflow identity onto a caller-owned OpenTelemetry span
- [`integrations/mem0`](./integrations/mem0) — retrieve semantic Memory in an explicit node and build the Agent Context in application code

The session query lists related Runs. Snapshot and resume select a mandatory `RunID`; there is no Session snapshot. The shared `workflow.MemoryStore` holds process-lifetime execution checkpoints and journal records, not semantic Memory.

## OpenTelemetry integration

Use `integrations/otel` when an application already owns OpenTelemetry setup and wants Workflow events on the active span. The example maps `SessionID` to `gen_ai.conversation.id`, `RunID` to `gopact.run.id`, and the Workflow definition ID to `gopact.workflow.name`.

This keeps telemetry identity out of the domain Event and storage schema, leaves the core free of an OpenTelemetry dependency, and works with any SDK exporter. The adapter only enriches a valid span carried by the invocation `context.Context`; without one, the OpenTelemetry API is a no-op.

## Mem0 integration

Use `integrations/mem0` when an Agent needs semantic Memory from Mem0 or a compatible HTTP service. It solves retrieval and scope mapping with an explicit typed topology:

```text
load-memory (HTTP I/O) -> build-model-request (pure) -> model
```

The application constructs the Agent Context that determines what the model sees; Memory is one input to that Context, not a framework-owned container or provider interface.

The retrieval node reads SessionID and Workflow RunID from `workflow.RunInfoFromContext`; business Context does not duplicate execution metadata. The caller supplies identity through RunOptions, and the framework propagates the resulting identity to the node.

| Application identity | Mem0 / model mapping |
| --- | --- |
| UserID | `user_id` |
| Agent identity | `agent_id` |
| SessionID | Mem0 `run_id` |
| Workflow RunID | `gopact.workflow.run_id` in `ModelRequest.Metadata` for provenance |

Advantages: the I/O boundary is visible in the Workflow, provider policy stays in application code, and no Mem0 dependency enters core or ext. Limitations: the application owns result selection, prompt construction, HTTP compatibility, and failure policy; the minimal client demonstrates one `POST /search` contract rather than a complete Mem0 SDK. To prevent API-key disclosure, it rejects every redirect, including same-origin redirects; configure the final endpoint URL directly.

The deterministic example uses an offline response. To run the bounded external smoke test, optionally load the repository-local `.env` first:

```bash
set -a; [ ! -f .env ] || . ./.env; set +a
MEM0_INTEGRATION=1 go test -tags=integration ./integrations/mem0 -run TestMem0Smoke -count=1 -v
```

`MEM0_BASE_URL` defaults to `http://localhost:8888`; `MEM0_API_KEY` is optional.

Run all examples:

```bash
go test ./...
```
