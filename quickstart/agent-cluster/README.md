# Agent Cluster

Run a local A2A-style agent cluster without external services or model credentials.

```bash
go run ./quickstart/agent-cluster
```

This example starts four local HTTP agents, discovers their agent cards, registers them in an `a2a.Registry`, and executes a `graph` workflow across planner, research, code, and review agents. The gateway calls agents by name, records graph runtime events into a `RunExport`, builds a self-bootstrap release gate bundle, writes local checkpoints, resumes from the latest checkpoint, prints artifact handoff refs, gates the review stream through policy, records failure attribution for a missing remote agent, and consumes the review agent as a stream.

CI runs this example with local mock agents only. Provider-backed examples use `.env` for local integration testing and are kept separate from this cluster smoke test.
