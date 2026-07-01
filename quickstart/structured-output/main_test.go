package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunRequestsStructuredOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %q, want /chat/completions", r.URL.Path)
		}
		var body struct {
			ResponseFormat *struct {
				Type       string `json:"type"`
				JSONSchema *struct {
					Strict bool           `json:"strict"`
					Schema map[string]any `json:"schema"`
				} `json:"json_schema"`
			} `json:"response_format"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if body.ResponseFormat == nil ||
			body.ResponseFormat.Type != "json_schema" ||
			body.ResponseFormat.JSONSchema == nil ||
			!body.ResponseFormat.JSONSchema.Strict ||
			body.ResponseFormat.JSONSchema.Schema["type"] != "object" {
			t.Fatalf("response_format = %#v, want strict object json schema", body.ResponseFormat)
		}

		_, _ = w.Write([]byte(`{
			"choices": [{
				"message": {
					"role": "assistant",
					"content": "{\"status\":\"ok\",\"summary\":\"structured output ready\"}"
				}
			}]
		}`))
	}))
	defer server.Close()

	t.Setenv("GOPACT_LLM_BASEURL", server.URL)
	t.Setenv("GOPACT_LLM_TOKEN", "test-token")
	t.Setenv("GOPACT_LLM_MODEL", "test-model")

	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if got := out.String(); !strings.Contains(got, "status=ok summary=structured output ready") {
		t.Fatalf("output = %q, want structured fields", got)
	}
}
