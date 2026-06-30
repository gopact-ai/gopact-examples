# Agent as Tool

Run a parent ReAct agent that delegates one task to a Plan-Exec child agent exposed as a normal tool.

```bash
go run ./quickstart/agent-as-tool
```

This example uses only scripted local models. It demonstrates `agenttool.New`, `a2a.NewRunnableAgent`, child A2A completion evidence, and runtime identity propagation without external credentials.
