# Feature Coverage

This matrix is the examples repository contract for expected runnable capabilities. CI uses local mocks for these commands; provider-backed checks stay local unless explicitly run with integration tags.

| Capability | Path | Mock test | Local integration |
| --- | --- | --- | --- |
| dotenv configuration | `internal/exampleenv` | `go test -count=1 ./internal/exampleenv` | - |
| scripted ReAct loop | `quickstart/react-agent` | `go test -count=1 ./quickstart/react-agent` | - |
| workflow graph branch, fan-in, loop, subgraph, and step limit | `quickstart/workflow-graph` | `go test -count=1 ./quickstart/workflow-graph` | - |
| checkpoint approval resume | `quickstart/agent-scaffold` | `go test -count=1 ./quickstart/agent-scaffold` | - |
| verification bundle | `quickstart/agent-scaffold` | `go test -count=1 ./quickstart/agent-scaffold` | - |
| A2A file registry scaffold | `quickstart/agent-scaffold` | `go test -count=1 ./quickstart/agent-scaffold` | - |
| core agent init/run scaffold | `quickstart/generated-agent` | `go test -count=1 ./quickstart/generated-agent` | - |
| Plan-Execute workflow with approval resume and cancel | `quickstart/plan-exec` | `go test -count=1 ./quickstart/plan-exec` | - |
| agent as tool success and failure evidence | `quickstart/agent-as-tool` | `go test -count=1 ./quickstart/agent-as-tool` | - |
| A2A local cluster + multi-source discovery + tag route + fallback + cancel | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| A2A local cluster run export golden trajectory | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| A2A local cluster policy deny and review | `quickstart/agent-cluster` | `go test -count=1 ./quickstart/agent-cluster` | - |
| OpenAI-compatible chat | `quickstart/openai-chat` | `go test -count=1 ./quickstart/openai-chat` | - |
| OpenAI-compatible streaming | `quickstart/openai-streaming` | `go test -count=1 ./quickstart/openai-streaming` | - |
| tool calling | `quickstart/tool-calling` | `go test -count=1 ./quickstart/tool-calling` | - |
| structured output | `quickstart/structured-output` | `go test -count=1 ./quickstart/structured-output` | - |
| Ark SDK provider | `quickstart/ark-chat` | `go test -count=1 ./quickstart/ark-chat` | - |
| Ark OpenAI-compatible streaming | `quickstart/ark-streaming` | `go test -count=1 ./quickstart/ark-streaming` | - |
| Agnes provider | `quickstart/agnes-chat` | `go test -count=1 ./quickstart/agnes-chat` | `go test -tags=integration -count=1 ./quickstart/agnes-chat` |
