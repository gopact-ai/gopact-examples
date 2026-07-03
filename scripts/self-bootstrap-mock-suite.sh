#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

go test -count=1 ./quickstart/react-agent
go test -count=1 ./quickstart/workflow-graph
go test -count=1 ./quickstart/agent-scaffold
go test -count=1 ./quickstart/generated-agent
go test -count=1 ./quickstart/plan-exec
go test -count=1 ./quickstart/supervisor
go test -count=1 ./quickstart/agent-as-tool
go test -count=1 ./quickstart/agent-node
go test -count=1 ./quickstart/agent-cluster
