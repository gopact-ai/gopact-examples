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

func TestLoadArkOpenAIConfigDefaultsRegionAndBaseURL(t *testing.T) {
	chdir(t, t.TempDir())
	t.Setenv("GOPACT_LLM_BASEURL", "")
	t.Setenv("GOPACT_LLM_TOKEN", "token")
	t.Setenv("GOPACT_LLM_MODEL", "ep-test")

	cfg, err := LoadArkOpenAIConfig()
	if err != nil {
		t.Fatalf("LoadArkOpenAIConfig() error = %v", err)
	}
	if cfg.BaseURL != ArkDefaultBaseURL {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, ArkDefaultBaseURL)
	}
}

func TestLoadArkOpenAIConfigRequiresModelAndCredentials(t *testing.T) {
	chdir(t, t.TempDir())
	t.Setenv("GOPACT_LLM_BASEURL", "")
	t.Setenv("GOPACT_LLM_TOKEN", "")
	t.Setenv("GOPACT_LLM_MODEL", "")

	if _, err := LoadArkOpenAIConfig(); err == nil {
		t.Fatal("LoadArkOpenAIConfig() error = nil, want missing env error")
	}
}

func TestLoadAgnesConfigUsesAgnesSpecificEnv(t *testing.T) {
	chdir(t, t.TempDir())
	t.Setenv("GOPACT_LLM_BASEURL", "https://ark.example.test/api/v3")
	t.Setenv("GOPACT_LLM_TOKEN", "shared-token")
	t.Setenv("GOPACT_LLM_MODEL", "shared-model")
	t.Setenv("GOPACT_AGNES_API_KEY", "agnes-token")
	t.Setenv("GOPACT_AGNES_MODEL", "agnes-model")

	cfg, err := LoadAgnesConfig()
	if err != nil {
		t.Fatalf("LoadAgnesConfig() error = %v", err)
	}
	if cfg.BaseURL != AgnesDefaultBaseURL {
		t.Fatalf("BaseURL = %q, want Agnes default", cfg.BaseURL)
	}
	if cfg.Token != "agnes-token" || cfg.Model != "agnes-model" {
		t.Fatalf("config = %+v, want Agnes-specific token/model", cfg)
	}
}

func TestLoadAgnesConfigSupportsSharedLLMEnv(t *testing.T) {
	chdir(t, t.TempDir())
	t.Setenv("GOPACT_LLM_BASEURL", "https://agnes.example.test/v1")
	t.Setenv("GOPACT_LLM_TOKEN", "shared-token")
	t.Setenv("GOPACT_LLM_MODEL", "shared-model")

	cfg, err := LoadAgnesConfig()
	if err != nil {
		t.Fatalf("LoadAgnesConfig() error = %v", err)
	}
	if cfg.BaseURL != "https://agnes.example.test/v1" || cfg.Token != "shared-token" || cfg.Model != "shared-model" {
		t.Fatalf("config = %+v, want shared LLM config", cfg)
	}
}

func TestLoadAgnesConfigRequiresCredential(t *testing.T) {
	chdir(t, t.TempDir())
	t.Setenv("GOPACT_AGNES_API_KEY", "")
	t.Setenv("GOPACT_LLM_TOKEN", "")

	if _, err := LoadAgnesConfig(); err == nil {
		t.Fatal("LoadAgnesConfig() error = nil, want missing credential error")
	}
}

func TestLoadArkSDKConfigUsesArkSpecificEnv(t *testing.T) {
	chdir(t, t.TempDir())
	t.Setenv("GOPACT_LLM_TOKEN", "openai-compatible-token")
	t.Setenv("GOPACT_ARK_API_KEY", "ark-sdk-token")
	t.Setenv("GOPACT_ARK_MODEL", "ep-test")

	cfg, err := LoadArkSDKConfig()
	if err != nil {
		t.Fatalf("LoadArkSDKConfig() error = %v", err)
	}
	if cfg.APIKey != "ark-sdk-token" {
		t.Fatalf("APIKey = %q, want ark-specific token", cfg.APIKey)
	}
	if cfg.BaseURL != ArkDefaultBaseURL || cfg.Region != ArkDefaultRegion {
		t.Fatalf("defaults = %q/%q, want %q/%q", cfg.BaseURL, cfg.Region, ArkDefaultBaseURL, ArkDefaultRegion)
	}
}

func TestLoadArkSDKConfigSupportsAkSk(t *testing.T) {
	chdir(t, t.TempDir())
	t.Setenv("GOPACT_ARK_ACCESS_KEY", "ak")
	t.Setenv("GOPACT_ARK_SECRET_KEY", "sk")
	t.Setenv("GOPACT_ARK_MODEL", "ep-test")

	cfg, err := LoadArkSDKConfig()
	if err != nil {
		t.Fatalf("LoadArkSDKConfig() error = %v", err)
	}
	if cfg.AccessKey != "ak" || cfg.SecretKey != "sk" {
		t.Fatalf("ak/sk = %q/%q, want ak/sk", cfg.AccessKey, cfg.SecretKey)
	}
}

func TestLoadArkSDKConfigRequiresModelAndCredentials(t *testing.T) {
	chdir(t, t.TempDir())
	t.Setenv("GOPACT_ARK_API_KEY", "")
	t.Setenv("GOPACT_ARK_ACCESS_KEY", "")
	t.Setenv("GOPACT_ARK_SECRET_KEY", "")
	t.Setenv("GOPACT_ARK_MODEL", "")

	if _, err := LoadArkSDKConfig(); err == nil {
		t.Fatal("LoadArkSDKConfig() error = nil, want missing env error")
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
