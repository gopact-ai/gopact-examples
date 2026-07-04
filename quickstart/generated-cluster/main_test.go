package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestRunGeneratesTestedA2AClusterScaffold(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var out bytes.Buffer
	if err := run(ctx, &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	line := strings.TrimSpace(out.String())
	if !strings.Contains(line, "generated generated-cluster at ") {
		t.Fatalf("output = %q, want generated scaffold path", out.String())
	}
	if !strings.Contains(line, "served generated-cluster at http://") {
		t.Fatalf("output = %q, want generated cluster run smoke evidence", out.String())
	}
	if !strings.Contains(line, "verified generated-cluster with gopact agent verify") {
		t.Fatalf("output = %q, want generated cluster verify evidence", out.String())
	}
	dir := strings.TrimPrefix(strings.Split(line, "\n")[0], "generated generated-cluster at ")
	for _, file := range []string{"go.mod", "main.go", "main_test.go", "agents.json", "README.md", ".env.example", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(dir, file)); err != nil {
			t.Fatalf("generated scaffold missing %s: %v", file, err)
		}
	}
	assertGeneratedFileContains(t, filepath.Join(dir, "go.mod"), "github.com/gopact-ai/gopact "+gopactVersion)
	assertGeneratedFileContains(t, filepath.Join(dir, "main.go"), "a2a.NewHTTPRegistryHandler")
	assertGeneratedFileContains(t, filepath.Join(dir, "main.go"), "a2a.NewHTTPRegistry")
	assertGeneratedFileContains(t, filepath.Join(dir, "main.go"), "a2a.NewMesh")
	assertGeneratedFileMatches(t, filepath.Join(dir, "main.go"), regexp.MustCompile(`clusterRegistryURLEnv\s*=\s*"GOPACT_A2A_REGISTRY_URL"`))
	assertGeneratedFileContains(t, filepath.Join(dir, "main.go"), "bootstrapClusterMeshFromEnv")
	assertGeneratedFileStartsWith(t, filepath.Join(dir, "agents.json"), "[")
	assertGeneratedFileContains(t, filepath.Join(dir, "main_test.go"), "TestClusterRegistryBootstrapsMesh")
	assertGeneratedFileContains(t, filepath.Join(dir, "main_test.go"), "TestClusterBootstrapsMeshFromEnvRegistryURL")
	assertGeneratedFileContains(t, filepath.Join(dir, "main_test.go"), "TestClusterRoutesStreamingTasks")
	assertGeneratedFileContains(t, filepath.Join(dir, "README.md"), "gopact agent verify .")
	assertGeneratedFileContains(t, filepath.Join(dir, "README.md"), "GOPACT_A2A_REGISTRY_URL")
	assertGeneratedFileContains(t, filepath.Join(dir, ".env.example"), "GOPACT_CLUSTER_URL=http://localhost:8080")
	assertGeneratedFileContains(t, filepath.Join(dir, ".env.example"), "GOPACT_A2A_REGISTRY_URL=http://localhost:8080/agents.json")
	assertGeneratedFileContains(t, filepath.Join(dir, ".gitignore"), ".env")
}

func TestReadmeMentionsCoreAgentInitCluster(t *testing.T) {
	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	for _, want := range []string{
		"gopact agent init-cluster",
		"/agents.json",
		"go run ./quickstart/generated-cluster",
	} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("README.md missing %q", want)
		}
	}
}

func TestQuickstartUsesClusterScaffoldDefaults(t *testing.T) {
	raw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	if strings.Contains(string(raw), "-sdk-version") {
		t.Fatalf("quickstart should exercise gopact agent init-cluster default SDK version, not pass -sdk-version")
	}
	if strings.Contains(string(raw), "-module") {
		t.Fatalf("quickstart should exercise gopact agent init-cluster default module path, not pass -module")
	}
}

func assertGeneratedFileContains(t *testing.T, path, want string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(raw), want) {
		t.Fatalf("%s missing %q:\n%s", path, want, raw)
	}
}

func assertGeneratedFileMatches(t *testing.T, path string, want *regexp.Regexp) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !want.Match(raw) {
		t.Fatalf("%s missing pattern %q:\n%s", path, want, raw)
	}
}

func assertGeneratedFileStartsWith(t *testing.T, path, want string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(raw)), want) {
		t.Fatalf("%s does not start with %q:\n%s", path, want, raw)
	}
}
