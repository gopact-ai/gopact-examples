package exampleenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnvSearchesParentsAndDoesNotOverrideEnv(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "quickstart", "openai-chat")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(`
GOPACT_LLM_BASEURL=https://example.test/v1
GOPACT_LLM_TOKEN=from-file
GOPACT_LLM_MODEL='test-model'
`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("GOPACT_LLM_TOKEN", "from-env")

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Chdir() cleanup error = %v", err)
		}
	})
	if err := os.Chdir(child); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	if err := LoadDotEnv(); err != nil {
		t.Fatalf("LoadDotEnv() error = %v", err)
	}

	if got := os.Getenv("GOPACT_LLM_BASEURL"); got != "https://example.test/v1" {
		t.Fatalf("GOPACT_LLM_BASEURL = %q, want file value", got)
	}
	if got := os.Getenv("GOPACT_LLM_TOKEN"); got != "from-env" {
		t.Fatalf("GOPACT_LLM_TOKEN = %q, want existing env value", got)
	}
	if got := os.Getenv("GOPACT_LLM_MODEL"); got != "test-model" {
		t.Fatalf("GOPACT_LLM_MODEL = %q, want unquoted file value", got)
	}
}

func TestLoadConfigRequiresLLMEnv(t *testing.T) {
	chdir(t, t.TempDir())
	t.Setenv("GOPACT_LLM_BASEURL", "")
	t.Setenv("GOPACT_LLM_TOKEN", "")
	t.Setenv("GOPACT_LLM_MODEL", "")

	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() error = nil, want missing env error")
	}
}

func TestLoadArkConfigDefaultsRegionAndBaseURL(t *testing.T) {
	chdir(t, t.TempDir())
	t.Setenv("GOPACT_LLM_BASEURL", "")
	t.Setenv("GOPACT_LLM_TOKEN", "token")
	t.Setenv("GOPACT_LLM_MODEL", "ep-test")

	cfg, err := LoadArkConfig()
	if err != nil {
		t.Fatalf("LoadArkConfig() error = %v", err)
	}
	if cfg.BaseURL != ArkDefaultBaseURL {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, ArkDefaultBaseURL)
	}
}

func TestLoadArkConfigRequiresModelAndCredentials(t *testing.T) {
	chdir(t, t.TempDir())
	t.Setenv("GOPACT_LLM_BASEURL", "")
	t.Setenv("GOPACT_LLM_TOKEN", "")
	t.Setenv("GOPACT_LLM_MODEL", "")

	if _, err := LoadArkConfig(); err == nil {
		t.Fatal("LoadArkConfig() error = nil, want missing env error")
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Chdir() cleanup error = %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
}
