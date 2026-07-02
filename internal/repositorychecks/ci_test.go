package repositorychecks

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestExamplesCIMockGateIsDocumented(t *testing.T) {
	workflow := readText(t, "../../.github/workflows/ci.yml")
	readme := readText(t, "../../README.md")

	for _, command := range []string{
		"git diff --check",
		"go mod tidy",
		"git diff --exit-code",
		"go test -count=1 ./...",
		"go test -race -count=1 ./...",
		"go vet ./...",
		"golangci-lint run ./...",
		"go test -coverprofile=coverage.out ./...",
		"govulncheck ./...",
	} {
		if !strings.Contains(workflow, command) {
			t.Fatalf("workflow missing mock CI command %q", command)
		}
		if !strings.Contains(readme, command) {
			t.Fatalf("README missing mock CI command %q", command)
		}
	}
	for _, action := range []string{"actions/checkout@v7", "actions/setup-go@v6"} {
		if !strings.Contains(workflow, action) {
			t.Fatalf("workflow missing current GitHub Action %q", action)
		}
	}

	for _, forbidden := range []string{"-tags=integration", ".env"} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("workflow contains %q; examples CI must stay mock-only", forbidden)
		}
	}
}

func TestQuickstartsAreDocumentedAndTested(t *testing.T) {
	entries, err := os.ReadDir("../../quickstart")
	if err != nil {
		t.Fatalf("read quickstart: %v", err)
	}

	wantCommands := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, file := range []string{"README.md", "main.go", "main_test.go"} {
			path := filepath.Join("../../quickstart", name, file)
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("quickstart/%s missing %s: %v", name, file, err)
			}
		}
		wantCommands = append(wantCommands, "go run ./quickstart/"+name)
	}
	slices.Sort(wantCommands)

	gotCommands := quickstartCommands(readText(t, "../../README.md"))
	if !slices.Equal(gotCommands, wantCommands) {
		t.Fatalf("README quickstart commands = %v, want %v", gotCommands, wantCommands)
	}
}

func TestAgnesLocalIntegrationCommandIsDocumented(t *testing.T) {
	readme := readText(t, "../../README.md")
	envExample := readText(t, "../../.env.example")

	command := "go test -tags=integration -count=1 ./quickstart/agnes-chat"
	if !strings.Contains(readme, command) {
		t.Fatalf("README missing Agnes integration command %q", command)
	}
	for _, key := range []string{"GOPACT_AGNES_API_KEY", "GOPACT_AGNES_SK", "GOPACT_LLM_TOKEN"} {
		if !strings.Contains(readme, key) {
			t.Fatalf("README missing Agnes credential variable %q", key)
		}
		if !strings.Contains(envExample, key) {
			t.Fatalf(".env.example missing Agnes credential variable %q", key)
		}
	}
}

func TestAgentClusterDiscoveryEnvIsDocumented(t *testing.T) {
	for _, path := range []string{"../../README.md", "../../.env.example", "../../quickstart/agent-cluster/README.md"} {
		text := readText(t, path)
		for _, key := range []string{"GOPACT_A2A_REGISTRY_FILE", "GOPACT_A2A_REGISTRY_URL", "GOPACT_A2A_ENDPOINTS"} {
			if !strings.Contains(text, key) {
				t.Fatalf("%s missing %s", path, key)
			}
		}
	}
}

func TestExamplesUseCurrentReleasedModules(t *testing.T) {
	goMod := readText(t, "../../go.mod")
	generatedAgent := readText(t, "../../quickstart/generated-agent/main.go")

	for _, requirement := range []string{
		"github.com/gopact-ai/gopact v0.0.30",
		"github.com/gopact-ai/gopact-ext/agents/agenttool v0.1.13",
		"github.com/gopact-ai/gopact-ext/agents/planexec v0.2.13",
		"github.com/gopact-ai/gopact-ext/agents/react v0.2.12",
		"github.com/gopact-ai/gopact-ext/devagent/filesnapshot v0.1.11",
		"github.com/gopact-ai/gopact-ext/devagent/gitdiff v0.1.11",
		"github.com/gopact-ai/gopact-ext/models/agnes v0.1.14",
		"github.com/gopact-ai/gopact-ext/models/ark v0.2.12",
		"github.com/gopact-ai/gopact-ext/models/openai v0.5.14",
	} {
		if !strings.Contains(goMod, requirement) {
			t.Fatalf("go.mod missing current released module %q", requirement)
		}
	}
	if !strings.Contains(generatedAgent, `gopactVersion = "v0.0.30"`) {
		t.Fatal("quickstart/generated-agent must exercise gopact agent init at current core SDK v0.0.30")
	}
}

func TestScaffoldPathUsesCredentialFreeQuickstarts(t *testing.T) {
	readme := readText(t, "../../README.md")
	for _, phrase := range []string{
		"## Scaffold Path",
		"Start without credentials:",
		"go run ./quickstart/react-agent",
		"go run ./quickstart/plan-exec",
		"go run ./quickstart/agent-as-tool",
		"go run ./quickstart/agent-cluster",
		"Use provider quickstarts after `.env` is configured.",
	} {
		if !strings.Contains(readme, phrase) {
			t.Fatalf("README missing scaffold path phrase %q", phrase)
		}
	}
}

func readText(t *testing.T, path string) string {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func quickstartCommands(readme string) []string {
	seen := map[string]struct{}{}
	var commands []string
	for _, line := range strings.Split(readme, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go run ./quickstart/") {
			if _, ok := seen[line]; ok {
				continue
			}
			seen[line] = struct{}{}
			commands = append(commands, line)
		}
	}
	slices.Sort(commands)
	return commands
}
