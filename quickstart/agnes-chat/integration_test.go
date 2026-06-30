//go:build integration

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gopact-ai/gopact-examples/internal/exampleenv"
)

func TestAgnesChatIntegrationUsesDotEnvProvider(t *testing.T) {
	if _, err := exampleenv.LoadAgnesConfig(); err != nil {
		if strings.Contains(err.Error(), "missing required environment variables") {
			t.Skip(err)
		}
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var out bytes.Buffer
	if err := run(ctx, &out); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Fatal("Agnes response is empty")
	}
}
