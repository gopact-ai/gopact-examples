package exampleenv

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	BaseURLEnv        = "GOPACT_LLM_BASEURL"
	TokenEnv          = "GOPACT_LLM_TOKEN"
	ModelEnv          = "GOPACT_LLM_MODEL"
	ArkDefaultBaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	ArkDefaultRegion  = "cn-beijing"

	ArkBaseURLEnv   = "GOPACT_ARK_BASEURL"
	ArkRegionEnv    = "GOPACT_ARK_REGION"
	ArkAPIKeyEnv    = "GOPACT_ARK_API_KEY"
	ArkAccessKeyEnv = "GOPACT_ARK_ACCESS_KEY"
	ArkSecretKeyEnv = "GOPACT_ARK_SECRET_KEY"
	ArkModelEnv     = "GOPACT_ARK_MODEL"
)

type Config struct {
	BaseURL string
	Token   string
	Model   string
}

type ArkSDKConfig struct {
	BaseURL   string
	Region    string
	APIKey    string
	AccessKey string
	SecretKey string
	Model     string
}

func LoadConfig() (Config, error) {
	if err := LoadDotEnv(); err != nil {
		return Config{}, err
	}

	cfg := Config{
		BaseURL: strings.TrimSpace(os.Getenv(BaseURLEnv)),
		Token:   strings.TrimSpace(os.Getenv(TokenEnv)),
		Model:   strings.TrimSpace(os.Getenv(ModelEnv)),
	}
	var missing []string
	if cfg.BaseURL == "" {
		missing = append(missing, BaseURLEnv)
	}
	if cfg.Token == "" {
		missing = append(missing, TokenEnv)
	}
	if cfg.Model == "" {
		missing = append(missing, ModelEnv)
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

func LoadArkOpenAIConfig() (Config, error) {
	if err := LoadDotEnv(); err != nil {
		return Config{}, err
	}

	cfg := Config{
		BaseURL: strings.TrimSpace(os.Getenv(BaseURLEnv)),
		Token:   strings.TrimSpace(os.Getenv(TokenEnv)),
		Model:   strings.TrimSpace(os.Getenv(ModelEnv)),
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = ArkDefaultBaseURL
	}

	var missing []string
	if cfg.Token == "" {
		missing = append(missing, TokenEnv)
	}
	if cfg.Model == "" {
		missing = append(missing, ModelEnv)
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

func LoadArkSDKConfig() (ArkSDKConfig, error) {
	if err := LoadDotEnv(); err != nil {
		return ArkSDKConfig{}, err
	}

	cfg := ArkSDKConfig{
		BaseURL:   strings.TrimSpace(os.Getenv(ArkBaseURLEnv)),
		Region:    strings.TrimSpace(os.Getenv(ArkRegionEnv)),
		APIKey:    strings.TrimSpace(os.Getenv(ArkAPIKeyEnv)),
		AccessKey: strings.TrimSpace(os.Getenv(ArkAccessKeyEnv)),
		SecretKey: strings.TrimSpace(os.Getenv(ArkSecretKeyEnv)),
		Model:     strings.TrimSpace(os.Getenv(ArkModelEnv)),
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = ArkDefaultBaseURL
	}
	if cfg.Region == "" {
		cfg.Region = ArkDefaultRegion
	}

	var missing []string
	if cfg.Model == "" {
		missing = append(missing, ArkModelEnv)
	}
	if cfg.APIKey == "" && (cfg.AccessKey == "" || cfg.SecretKey == "") {
		missing = append(missing, ArkAPIKeyEnv+" or "+ArkAccessKeyEnv+"+"+ArkSecretKeyEnv)
	}
	if len(missing) > 0 {
		return ArkSDKConfig{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

func LoadDotEnv() error {
	path, ok, err := findDotEnv()
	if err != nil || !ok {
		return err
	}
	return loadDotEnvFile(path)
}

func findDotEnv() (string, bool, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false, err
	}
	for {
		path := filepath.Join(dir, ".env")
		if _, err := os.Stat(path); err == nil {
			return path, true, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", false, err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false, nil
		}
		dir = parent
	}
}

func loadDotEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return scanner.Err()
}
