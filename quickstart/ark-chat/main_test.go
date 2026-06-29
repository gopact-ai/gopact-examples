package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunUsesConfiguredArk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("path = %q, want /responses", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		_, _ = w.Write([]byte(`{"output_text": "hello from fake ark"}`))
	}))
	defer server.Close()

	t.Setenv("GOPACT_LLM_BASEURL", server.URL)
	t.Setenv("GOPACT_LLM_TOKEN", "test-token")
	t.Setenv("GOPACT_LLM_MODEL", "ep-test")

	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(out.String(), "hello from fake ark") {
		t.Fatalf("output = %q, want fake response", out.String())
	}
}
