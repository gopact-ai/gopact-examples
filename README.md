# 🧪 gopact-examples

<!-- gopact:doc-language: en -->

Chinese documentation: [README_zh.md](README_zh.md)

Executable examples for the redesigned `gopact` API.

> **Go 1.27+ only.** This project is built around generic methods and celebrates what we see as one of Go's most consequential language changes of the past decade. Until Go 1.27 is officially released, it requires a development toolchain and should be treated as a preview, not a stable release.

Default example runs are intentionally offline.

## Quickstarts

| Example | What it teaches |
| --- | --- |
| [`quickstart/model-basic`](./quickstart/model-basic) | Implement and invoke the minimal `gopact.Model` contract |
| [`quickstart/workflow-basic`](./quickstart/workflow-basic) | Build and run a typed Workflow with observable events |
| [`quickstart/react-basic`](./quickstart/react-basic) | Connect a model and tool through the ReAct Agent |

## Concepts

| Example | What it teaches |
| --- | --- |
| [`concepts/durable-resume`](./concepts/durable-resume) | Deduplicate an external side effect when recovery reruns the same activation |
| [`concepts/session-correlation`](./concepts/session-correlation) | Correlate independent Runs with a Session, then inspect and resume one selected Run |

The durable-resume example derives `RunInfo.RunID + "/" + RunInfo.ActivationID`, then proves that a node may run twice after simulated process loss while the side effect is applied once. Its in-memory idempotent API is a deterministic demonstration, not a production outbox. In production, pass the stable key to an external API that natively deduplicates it, or write a uniquely constrained dedup/outbox record in the same transaction as the business data. An explicit business retry intended to create a new side effect needs a new operation key.

The session query lists related Runs. Snapshot and resume select a mandatory `RunID`; there is no Session snapshot. The shared `workflow.MemoryStore` holds process-lifetime execution checkpoints and journal records, not semantic Memory, and is only for tests or short-lived processes. Use SQLite on one machine or from processes that safely share one local database file; use a distributed database Store with atomic Claim and fencing for multiple hosts.

## Integrations

| Example | What it teaches |
| --- | --- |
| [`integrations/otel`](./integrations/otel) | Map Workflow identity and events onto a caller-owned OpenTelemetry span |
| [`integrations/mem0`](./integrations/mem0) | Retrieve semantic Memory explicitly and build Agent Context in application code |

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

## Run all examples

```bash
go test ./...
```
