# Agent Cluster

Run a local A2A-style agent cluster without external services or model credentials.

```bash
go run ./quickstart/agent-cluster
```

This example starts four local HTTP agents, bootstraps their agent cards through `a2a.Mesh`, and executes a `graph` workflow across planner, research, code, and review agents. The gateway calls agents by name, records graph runtime events into a `RunExport`, captures local git diff, file snapshot, and A2A task evidence, builds a self-bootstrap release gate bundle, writes local checkpoints, resumes from the latest checkpoint, prints artifact handoff refs, gates the review stream through policy, records tag-route fallback evidence, records cancel evidence, records failure attribution for a missing remote agent, and consumes the review agent as a stream.

Set `GOPACT_A2A_REGISTRY_FILE=./agents.json` to bootstrap from an existing agent-card file instead of the demo's temporary local registry.
Set `GOPACT_A2A_REGISTRY_URL=http://localhost:8080/agents.json` to bootstrap from an HTTP agent-card registry document.
Set `GOPACT_A2A_ENDPOINTS=http://localhost:8080,http://localhost:8081` to bootstrap by fetching each endpoint's well-known agent card. If multiple discovery variables are set, the example bootstraps all configured sources in file, registry URL, then endpoint order.

CI runs this example with local mock agents only and fixes the cluster `RunExport` event sequence with a golden trajectory. Provider-backed examples use `.env` for local integration testing and are kept separate from this cluster smoke test.
