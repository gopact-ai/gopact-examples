package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunGeneratesTestedA2AAgentScaffold(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var out bytes.Buffer
	if err := run(ctx, &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	line := strings.TrimSpace(out.String())
	if !strings.HasPrefix(line, "generated generated-agent at ") {
		t.Fatalf("output = %q, want generated scaffold path", out.String())
	}
	dir := strings.TrimPrefix(line, "generated generated-agent at ")
	for _, file := range []string{"go.mod", "main.go", "main_test.go", "agents.json", "README.md", ".env.example", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(dir, file)); err != nil {
			t.Fatalf("generated scaffold missing %s: %v", file, err)
		}
	}
	assertGeneratedFileContains(t, filepath.Join(dir, "go.mod"), "github.com/gopact-ai/gopact "+gopactVersion)
	assertGeneratedFileContains(t, filepath.Join(dir, "main.go"), "a2a.NewHTTPHandler(agent)")
	assertGeneratedFileContains(t, filepath.Join(dir, "main.go"), "signal.NotifyContext")
	assertGeneratedFileContains(t, filepath.Join(dir, "main.go"), "server.Shutdown")
	assertGeneratedFileContains(t, filepath.Join(dir, "main.go"), "a2a.NewHTTPRegistryHandler")
	assertGeneratedFileContains(t, filepath.Join(dir, "main_test.go"), "a2a.NewHTTPRegistry")
	assertGeneratedFileContains(t, filepath.Join(dir, "main_test.go"), "TestScaffoldAgentServesHealthEndpoints")
	assertGeneratedFileContains(t, filepath.Join(dir, "main_test.go"), "TestScaffoldServerStopsOnContextCancel")
	assertGeneratedFileContains(t, filepath.Join(dir, "README.md"), "gopact agent run .")
	assertGeneratedFileContains(t, filepath.Join(dir, ".env.example"), "GOPACT_AGENT_URL=http://localhost:8080")
	assertGeneratedFileContains(t, filepath.Join(dir, ".gitignore"), ".env")
}

func TestReadmeMentionsCoreAgentInit(t *testing.T) {
	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	for _, want := range []string{
		"gopact agent init",
		"/agents.json",
		"go run ./quickstart/generated-agent",
	} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("README.md missing %q", want)
		}
	}
}

func TestQuickstartUsesScaffoldDefaultSDKVersion(t *testing.T) {
	raw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	if strings.Contains(string(raw), "-sdk-version") {
		t.Fatalf("quickstart should exercise gopact agent init default SDK version, not pass -sdk-version")
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
