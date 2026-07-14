# 🧪 gopact-examples

<!-- gopact:doc-language: en -->

Chinese documentation: [README_zh.md](README_zh.md)

Executable examples for the redesigned `gopact` API.

> **Go 1.27+ only.** This project is built around generic methods and celebrates what we see as one of Go's most consequential language changes of the past decade. Until Go 1.27 is officially released, it requires a development toolchain and should be treated as a preview, not a stable release.

Before the coordinated RC modules are published, the manual source E2E workflow requires
reviewed 40-character core and ext commit SHAs, checks out those exact commits, prints all
three SHAs, and joins them with a temporary Go workspace. Ordinary CI consumes the immutable
module versions declared by the examples module with `GOWORK=off`.

The release order is core → the two ext modules → examples. After the approved immutable dependency tags exist, this module must pin those exact versions and pass with `GOWORK=off`. That post-tag gate has not passed yet; RCs remain production-evaluation candidates until Go 1.27 stable validation and burn-in complete.

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
| [`concepts/durable-resume`](./concepts/durable-resume) | Resume one interrupted Run from its checkpoint |
| [`concepts/run-control`](./concepts/run-control) | Retry or fork a failed Run into a new Run with source lineage |
| [`concepts/session-correlation`](./concepts/session-correlation) | Correlate independent Runs with a Session, then inspect and resume one selected Run |

The durable-resume example keeps only the public interrupt/resume path. Fresh-process decoding, fencing, crash windows, and side-effect idempotency belong to the core and Store integration suites rather than a quick conceptual example. `MemoryStore` keeps this example offline; replace it with a durable Store before relying on process recovery.

The run-control example leaves the failed source Run immutable. `Retry` replays one failed node activation into a new Run, while `Fork` starts another new Run from a replay-safe root with a patched workflow input. Both new Runs retain `SourceRunID` lineage.

The session query lists related Runs. Snapshot and resume select a mandatory `RunID`; there is no Session snapshot. The shared `workflow.MemoryStore` holds process-lifetime execution checkpoints and journal records, not semantic Memory, and is only for tests or short-lived processes. Use SQLite on one machine or from processes that safely share one local database file; use a distributed database Store with atomic Claim and fencing for multiple hosts.

## Integrations

| Example | What it teaches |
| --- | --- |
| [`integrations/otel`](./integrations/otel) | Map Workflow identity and events onto a caller-owned OpenTelemetry span |
| [`integrations/mem0`](./integrations/mem0) | Retrieve semantic Memory explicitly and build Agent Context in application code |

## OpenTelemetry integration

Use `integrations/otel` when an application already owns OpenTelemetry setup. The example projects Workflow domain Events onto the run span, mapping `SessionID` to `gen_ai.conversation.id`, `RunID` to `gopact.run.id`, and the Workflow definition ID to `gopact.workflow.name`. It separately wraps an application adapter with an infrastructure span; that span does not manufacture a Workflow Event.

This keeps telemetry identity out of the domain Event and storage schema, leaves the core free of an OpenTelemetry dependency, and works with any SDK exporter. Both projections use the invocation `context.Context`; with the OpenTelemetry no-op provider they add no runtime telemetry.

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

From a published checkout:

```bash
GOWORK=off go mod download
GOWORK=off go test -count=1 ./...
```

The pre-tag source E2E workflow instead creates a temporary workspace over the three
coordinated source checkouts and runs:

```bash
go test ./...
```
