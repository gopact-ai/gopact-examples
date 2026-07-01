package repositorychecks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureCoverageMatrixDocumentsExpectedCapabilities(t *testing.T) {
	matrix := readText(t, "../../FEATURES.md")
	readme := readText(t, "../../README.md")
	if !strings.Contains(readme, "FEATURES.md") {
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
			capability:  "workflow graph",
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
			capability:  "Plan-Execute workflow",
			path:        "quickstart/plan-exec",
			mockCommand: "go test -count=1 ./quickstart/plan-exec",
		},
		{
			capability:  "agent as tool",
			path:        "quickstart/agent-as-tool",
			mockCommand: "go test -count=1 ./quickstart/agent-as-tool",
		},
		{
			capability:  "A2A local cluster",
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
