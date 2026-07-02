# gopact-examples

#### Runnable examples for gopact workflows, providers, A2A discovery, and agent templates.

[![CI](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/gopact-ai/gopact-examples/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/gopact-ai/gopact-examples.svg)](https://pkg.go.dev/github.com/gopact-ai/gopact-examples)
[![License](https://img.shields.io/github/license/gopact-ai/gopact-examples)](LICENSE)

<!-- gopact:doc-language: en -->

Chinese documentation: [README_zh.md](README_zh.md)

`gopact-examples` contains executable examples for [`gopact`](https://github.com/gopact-ai/gopact) and [`gopact-ext`](https://github.com/gopact-ai/gopact-ext). The repository favors complete local flows over isolated snippets: workflow graphs, agent templates, provider adapters, A2A discovery, verification, and development-agent evidence are all covered by tests.

CI uses fake LLM servers, scripted models, and local A2A agents. Real provider checks are local opt-in tests driven by `.env`.

## Scaffold Path

Start without credentials:

```bash
go run ./quickstart/react-agent
go run ./quickstart/plan-exec
go run ./quickstart/supervisor
go run ./quickstart/agent-as-tool
go run ./quickstart/agent-cluster
```

This path starts with a scripted ReAct agent, then adds Plan-Execute, supervisor routing, agent-as-tool delegation, and a local A2A agent cluster. Use provider quickstarts after `.env` is configured.

## Quickstarts

Run examples from the repository root:

```bash
go run ./quickstart/agent-as-tool
go run ./quickstart/agent-cluster
go run ./quickstart/agent-scaffold
go run ./quickstart/agnes-chat
go run ./quickstart/ark-chat
go run ./quickstart/ark-streaming
go run ./quickstart/generated-agent
go run ./quickstart/openai-chat
go run ./quickstart/openai-streaming
go run ./quickstart/plan-exec
go run ./quickstart/react-agent
go run ./quickstart/structured-output
go run ./quickstart/supervisor
go run ./quickstart/tool-calling
go run ./quickstart/workflow-graph
```

| Example | Demonstrates | Credentials |
| --- | --- | --- |
| `quickstart/react-agent` | Scripted ReAct loop and local tool calling. | No |
| `quickstart/workflow-graph` | Typed graph, branch fan-out and fan-in, subgraph, loop, and step limit. | No |
| `quickstart/agent-scaffold` | Checkpoint approval resume, verification bundle, and A2A file registry scaffold. | No |
| `quickstart/generated-agent` | Core agent init/run scaffold generated through `gopact agent init`. | No |
| `quickstart/plan-exec` | Plan-Execute workflow with replan, approval resume, and cancel. | No |
| `quickstart/supervisor` | Supervisor routing to named Plan-Execute child agents. | No |
| `quickstart/agent-as-tool` | Agent as tool success and failure evidence. | No |
| `quickstart/agent-cluster` | A2A local cluster, `Mesh.SyncEnv`/`Mesh.SyncEnvEvery` discovery, tag route, fallback, policy, retry, cancel, and Dev Agent test, review, replay, and command evidence. | No |
| `quickstart/openai-chat` | OpenAI-compatible chat. | Yes |
| `quickstart/openai-streaming` | OpenAI-compatible streaming. | Yes |
| `quickstart/tool-calling` | Tool calling through an OpenAI-compatible provider. | Yes |
| `quickstart/structured-output` | Structured output through JSON schema. | Yes |
| `quickstart/ark-chat` | Ark SDK provider. | Yes |
| `quickstart/ark-streaming` | Ark OpenAI-compatible streaming. | Yes |
| `quickstart/agnes-chat` | Agnes provider. | Yes |

## Configuration

Examples load `.env` from the current directory or a parent directory. `.env` is ignored; only `.env.example` is committed.

```bash
cp .env.example .env
```

OpenAI-shaped provider examples read:

- `GOPACT_LLM_BASEURL`
- `GOPACT_LLM_TOKEN`
- `GOPACT_LLM_MODEL`

Agnes examples also support:

- `GOPACT_AGNES_API_KEY`
- `GOPACT_AGNES_SK`
- `GOPACT_AGNES_MODEL`

Ark SDK examples read:

- `GOPACT_ARK_API_KEY`
- `GOPACT_ARK_ACCESS_KEY`
- `GOPACT_ARK_SECRET_KEY`
- `GOPACT_ARK_MODEL`
- `GOPACT_ARK_REGION`

A2A cluster discovery reads:

- `GOPACT_A2A_REGISTRY_FILE`
- `GOPACT_A2A_REGISTRY_URL`
- `GOPACT_A2A_ENDPOINTS`

The agent-cluster quickstart uses `Mesh.SyncEnv` to import env-configured cards, register callable HTTP agents, and prune unready endpoints before routing tasks. Its tests use `Mesh.SyncEnvEvery` to cover continuous registry refresh.

## Integration Tests

CI is mock-only. Run real provider tests explicitly from a local machine:

```bash
./scripts/local-agnes-integration.sh
go test -tags=integration -count=1 ./quickstart/agnes-chat
```

## Verification

Run the same gates before opening a pull request:

```bash
git diff --check
./scripts/public-readiness-check.sh
go mod tidy
git diff --exit-code
go test -count=1 ./...
go test -race -count=1 ./...
go vet ./...
golangci-lint run ./...
go test -coverprofile=coverage.out ./...
govulncheck ./...
```

## Documentation

- [doc/README.md](doc/README.md): documentation index.
- [doc/FEATURES.md](doc/FEATURES.md): executable capability matrix.
- [doc/CONTRIBUTING.md](doc/CONTRIBUTING.md): development setup, local checks, and pull request rules.
- [doc/SECURITY.md](doc/SECURITY.md): security policy and vulnerability reporting.
- [doc/CHANGELOG.md](doc/CHANGELOG.md): user-visible changes.
- [doc/maintainers/repository-governance.md](doc/maintainers/repository-governance.md): PR-only flow, CI gates, admin auto-merge, and public repository governance.

## Contributing

Keep examples runnable from the repository root, covered by mock tests, and documented with the exact command users should run. Provider credentials belong only in local `.env` files.
