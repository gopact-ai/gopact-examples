# Feature Coverage

<!-- gopact:doc-language: en -->

Chinese documentation: [FEATURES_zh.md](FEATURES_zh.md)

This matrix is the executable contract for `gopact-examples`. CI uses mocks, local fake servers, and scripted agents only. Real provider checks are local opt-in tests behind the `integration` build tag.

| Capability | Path | Mock command | Integration command |
| --- | --- | --- | --- |
| dotenv configuration | `internal/exampleenv` | `go test -count=1 ./internal/exampleenv` | Not required |
| scripted ReAct loop | `quickstart/react-agent` | `go test -count=1 ./quickstart/react-agent` | Not required |
| workflow graph branch, dynamic fan-out, fan-in, loop, subgraph, and step limit | `quickstart/workflow-graph` | `go test -count=1 ./quickstart/workflow-graph` | Not required |
| checkpoint approval resume | `quickstart/agent-scaffold` | `go test -count=1 ./quickstart/agent-scaffold` | Not required |
| verification bundle | `quickstart/agent-scaffold` | `go test -count=1 ./quickstart/agent-scaffold` | Not required |
| A2A file registry scaffold | `quickstart/agent-scaffold` | `go test -count=1 ./quickstart/agent-scaffold` | Not required |
| core agent init/run scaffold | `quickstart/generated-agent` | `go test -count=1 ./quickstart/generated-agent` | Not required |
| Plan-Execute workflow with replan, approval resume, and cancel | `quickstart/plan-exec` | `go test -count=1 ./quickstart/plan-exec` | Not required |
| Supervisor routing to named Plan-Execute child agents | `quickstart/supervisor` | `go test -count=1 ./quickstart/supervisor` | Not required |
| agent as tool success and failure evidence | `quickstart/agent-as-tool` | `go test -count=1 ./quickstart/agent-as-tool` | Not required |
| A2A local cluster + multi-source discovery + tag route + fallback + cancel | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | Not required |
| A2A env mesh sync with readiness pruning | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | Not required |
| A2A continuous env mesh sync with registry changes | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | Not required |
| A2A local cluster expiry-aware discovery and lease heartbeat evidence | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | Not required |
| A2A local cluster readiness-gated endpoint discovery | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | Not required |
| A2A local cluster run export golden trajectory | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | Not required |
| A2A local cluster policy deny and review | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | Not required |
| A2A local cluster retry evidence | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | Not required |
| Dev Agent test and review evidence | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | Not required |
| Dev Agent replay and command evidence | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | Not required |
| OpenAI-compatible chat | `quickstart/openai-chat` | `go test -count=1 ./quickstart/openai-chat` | Not required |
| OpenAI-compatible streaming | `quickstart/openai-streaming` | `go test -count=1 ./quickstart/openai-streaming` | Not required |
| tool calling | `quickstart/tool-calling` | `go test -count=1 ./quickstart/tool-calling` | Not required |
| structured output | `quickstart/structured-output` | `go test -count=1 ./quickstart/structured-output` | Not required |
| Ark SDK provider | `quickstart/ark-chat` | `go test -count=1 ./quickstart/ark-chat` | Not required |
| Ark OpenAI-compatible streaming | `quickstart/ark-streaming` | `go test -count=1 ./quickstart/ark-streaming` | Not required |
| Agnes provider | `quickstart/agnes-chat` | `go test -count=1 ./quickstart/agnes-chat` | `go test -tags=integration -count=1 ./quickstart/agnes-chat` |

Every quickstart must include `main.go`, `main_test.go`, and `README.md`. Provider examples must keep their default tests deterministic and credential-free.
