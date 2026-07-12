//go:build integration

package main

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestMem0Smoke verifies that a configured Mem0 endpoint accepts a scoped search request.
func TestMem0Smoke(t *testing.T) {
	if os.Getenv("MEM0_INTEGRATION") != "1" {
		t.Skip("set MEM0_INTEGRATION=1 to run the Mem0 smoke test")
	}
	baseURL := os.Getenv("MEM0_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8888"
	}
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()

	client := mem0Client{
		baseURL: baseURL,
		apiKey:  os.Getenv("MEM0_API_KEY"),
	}
	_, err := client.Search(ctx, searchRequest{
		Query:   "Which database does this user prefer?",
		UserID:  "user-7",
		AgentID: "support-agent",
		RunID:   "session-memory-9",
	})
	if err != nil {
		t.Fatalf("Mem0 search smoke test: %v", err)
	}
}
