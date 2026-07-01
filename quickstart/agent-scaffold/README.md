# Agent Scaffold

Minimal no-credential agent skeleton with graph nodes, checkpointing, approval interrupt, resume, and a verification report.

```bash
go run ./quickstart/agent-scaffold
```

This is the smallest local scaffold path before adding a real provider. It records a `RunExport`, builds a verification report, and embeds the report into a bundle; use `quickstart/agent-cluster` when the same scaffold needs A2A task evidence and a self-bootstrap release gate bundle. Use `quickstart/react-agent`, `quickstart/plan-exec`, and provider quickstarts to expand it.
