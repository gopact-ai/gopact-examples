package repositorychecks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureCoverageMatrixDocumentsExpectedCapabilities(t *testing.T) {
	matrix := readText(t, "../../doc/FEATURES.md")
	readme := readText(t, "../../README.md")
	if !strings.Contains(readme, "doc/FEATURES.md") {
		t.Fatal("README must link to FEATURES.md")
	}

	tests := []struct {
		capability         string
		path               string
		mockCommand        string
		integrationCommand string
	}{
		{
			capability:  "dotenv configuration",
			path:        "internal/exampleenv",
			mockCommand: "go test -count=1 ./internal/exampleenv",
		},
		{
			capability:  "scripted ReAct loop",
			path:        "quickstart/react-agent",
			mockCommand: "go test -count=1 ./quickstart/react-agent",
		},
		{
			capability: "workflow graph branch, dynamic fan-out, fan-in, loop, subgraph, " +
				"step limit, step export/import, and interrupted resume",
			path:        "quickstart/workflow-graph",
			mockCommand: "go test -count=1 ./quickstart/workflow-graph",
		},
		{
			capability:  "checkpoint approval resume",
			path:        "quickstart/agent-scaffold",
			mockCommand: "go test -count=1 ./quickstart/agent-scaffold",
		},
		{
			capability:  "verification bundle",
			path:        "quickstart/agent-scaffold",
			mockCommand: "go test -count=1 ./quickstart/agent-scaffold",
		},
		{
			capability:  "A2A file registry scaffold",
			path:        "quickstart/agent-scaffold",
			mockCommand: "go test -count=1 ./quickstart/agent-scaffold",
		},
		{
			capability:  "core agent init/verify/run scaffold with default module path",
			path:        "quickstart/generated-agent",
			mockCommand: "go test -count=1 ./quickstart/generated-agent",
		},
		{
			capability:  "core agent init-cluster/verify/run scaffold with default module path and env registry bootstrap",
			path:        "quickstart/generated-cluster",
			mockCommand: "go test -count=1 ./quickstart/generated-cluster",
		},
		{
			capability:  "Plan-Execute workflow with replan, approval resume, and cancel",
			path:        "quickstart/plan-exec",
			mockCommand: "go test -count=1 ./quickstart/plan-exec",
		},
		{
			capability:  "Supervisor routing to named Plan-Execute child agents",
			path:        "quickstart/supervisor",
			mockCommand: "go test -count=1 ./quickstart/supervisor",
		},
		{
			capability:  "agent as tool success and failure evidence",
			path:        "quickstart/agent-as-tool",
			mockCommand: "go test -count=1 ./quickstart/agent-as-tool",
		},
		{
			capability:  "leased background scheduler with retry, dead-letter, drain, lease release, and schedule evidence",
			path:        "quickstart/background-scheduler",
			mockCommand: "go test -count=1 ./quickstart/background-scheduler",
		},
		{
			capability:  "Dev Agent self-bootstrap workflow with policy-approved plan patch apply, quickstart release requirements, diff, file snapshot, command, CI gate, run export, failure attribution, and verification report evidence",
			path:        "quickstart/self-bootstrap",
			mockCommand: "go test -count=1 ./quickstart/self-bootstrap",
		},
		{
			capability:  "A2A child agent as typed graph node with nested evidence",
			path:        "quickstart/agent-node",
			mockCommand: "go test -count=1 ./quickstart/agent-node",
		},
		{
			capability:  "A2A local cluster + multi-source discovery + tag route + fallback + cancel",
			path:        "quickstart/agent-cluster",
			mockCommand: "go test -count=1 ./quickstart/agent-cluster",
		},
		{
			capability:  "A2A env mesh sync with mesh-level HTTP options and readiness pruning",
			path:        "quickstart/agent-cluster",
			mockCommand: "go test -count=1 ./quickstart/agent-cluster",
		},
		{
			capability:  "A2A continuous env mesh sync with registry changes",
			path:        "quickstart/agent-cluster",
			mockCommand: "go test -count=1 ./quickstart/agent-cluster",
		},
		{
			capability:  "A2A local cluster expiry-aware discovery and lease heartbeat evidence",
			path:        "quickstart/agent-cluster",
			mockCommand: "go test -count=1 ./quickstart/agent-cluster",
		},
		{
			capability:  "A2A local cluster readiness-gated endpoint discovery",
			path:        "quickstart/agent-cluster",
			mockCommand: "go test -count=1 ./quickstart/agent-cluster",
		},
		{
			capability:  "A2A local cluster run export golden trajectory",
			path:        "quickstart/agent-cluster",
			mockCommand: "go test -count=1 ./quickstart/agent-cluster",
		},
		{
			capability:  "A2A local cluster policy deny and review",
			path:        "quickstart/agent-cluster",
			mockCommand: "go test -count=1 ./quickstart/agent-cluster",
		},
		{
			capability:  "A2A local cluster retry evidence",
			path:        "quickstart/agent-cluster",
			mockCommand: "go test -count=1 ./quickstart/agent-cluster",
		},
		{
			capability:  "Dev Agent test and review evidence",
			path:        "quickstart/agent-cluster",
			mockCommand: "go test -count=1 ./quickstart/agent-cluster",
		},
		{
			capability:  "Dev Agent replay and command evidence",
			path:        "quickstart/agent-cluster",
			mockCommand: "go test -count=1 ./quickstart/agent-cluster",
		},
		{
			capability:  "OpenAI-compatible chat",
			path:        "quickstart/openai-chat",
			mockCommand: "go test -count=1 ./quickstart/openai-chat",
		},
		{
			capability:  "OpenAI-compatible streaming",
			path:        "quickstart/openai-streaming",
			mockCommand: "go test -count=1 ./quickstart/openai-streaming",
		},
		{
			capability:  "tool calling",
			path:        "quickstart/tool-calling",
			mockCommand: "go test -count=1 ./quickstart/tool-calling",
		},
		{
			capability:  "structured output",
			path:        "quickstart/structured-output",
			mockCommand: "go test -count=1 ./quickstart/structured-output",
		},
		{
			capability:  "Ark SDK provider",
			path:        "quickstart/ark-chat",
			mockCommand: "go test -count=1 ./quickstart/ark-chat",
		},
		{
			capability:  "Ark OpenAI-compatible streaming",
			path:        "quickstart/ark-streaming",
			mockCommand: "go test -count=1 ./quickstart/ark-streaming",
		},
		{
			capability:         "Agnes provider",
			path:               "quickstart/agnes-chat",
			mockCommand:        "go test -count=1 ./quickstart/agnes-chat",
			integrationCommand: "go test -tags=integration -count=1 ./quickstart/agnes-chat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.capability, func(t *testing.T) {
			for _, want := range []string{tt.capability, tt.path, tt.mockCommand} {
				if !strings.Contains(matrix, want) {
					t.Fatalf("FEATURES.md missing %q", want)
				}
			}
			if tt.integrationCommand != "" && !strings.Contains(matrix, tt.integrationCommand) {
				t.Fatalf("FEATURES.md missing integration command %q", tt.integrationCommand)
			}
			assertFeaturePathTested(t, tt.path)
		})
	}
}

func assertFeaturePathTested(t *testing.T, path string) {
	t.Helper()

	testFile := filepath.Join("../..", path, "main_test.go")
	if strings.HasPrefix(path, "internal/") {
		entries, err := os.ReadDir(filepath.Join("../..", path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
				return
			}
		}
		t.Fatalf("%s missing tested package file", path)
	}
	if _, err := os.Stat(testFile); err != nil {
		t.Fatalf("%s missing tested entrypoint: %v", path, err)
	}
}
