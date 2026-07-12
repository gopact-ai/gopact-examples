package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gopact-ai/gopact"
)

func TestMemoryWorkflowMapsScopesAndBuildsModelRequest(t *testing.T) {
	t.Parallel()

	wantSearch := map[string]string{
		"query":    "Which database does this user prefer?",
		"user_id":  "user-7",
		"agent_id": "support-agent",
		"run_id":   "session-memory-9",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/search" {
			t.Errorf("request = %s %s, want POST /search", r.Method, r.URL.Path)
		}
		var got map[string]string
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode search request: %v", err)
		}
		if !reflect.DeepEqual(got, wantSearch) {
			t.Errorf("search request = %#v, want %#v", got, wantSearch)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"results":[{"memory":"Prefers PostgreSQL for transactional data."},{"memory":"Uses pgvector for semantic search."}]}`)); err != nil {
			t.Errorf("write search response: %v", err)
		}
	}))
	defer server.Close()

	model := &recordingModel{}
	client := mem0Client{baseURL: server.URL, httpClient: server.Client()}
	answer, err := runMemoryWorkflow(t.Context(), client.Search, model, workflowInput{
		Query:       "Which database does this user prefer?",
		UserID:      "user-7",
		AgentID:     "support-agent",
		UserMessage: gopact.UserMessage("Recommend a database for my next service."),
	}, gopact.WithSessionID("session-memory-9"), gopact.WithRunID("run-memory-3"))
	if err != nil {
		t.Fatalf("runMemoryWorkflow() error = %v", err)
	}
	if answer != "PostgreSQL is a good fit." {
		t.Fatalf("response text = %q", answer)
	}

	wantMemory := "Prefers PostgreSQL for transactional data.\nUses pgvector for semantic search."
	if len(model.request.Messages) != 2 {
		t.Fatalf("model messages = %d, want 2", len(model.request.Messages))
	}
	if got := model.request.Messages[0].Parts[0].Text; got != wantMemory {
		t.Errorf("system memory = %q, want %q", got, wantMemory)
	}
	if got := model.request.Messages[1]; !reflect.DeepEqual(got, gopact.UserMessage("Recommend a database for my next service.")) {
		t.Errorf("user message = %#v", got)
	}
	if got := model.request.Metadata["gopact.workflow.run_id"]; got != "run-memory-3" {
		t.Errorf("workflow provenance = %q, want run-memory-3", got)
	}
	if got := model.request.Metadata["provider.request_profile"]; got != "balanced" {
		t.Errorf("provider metadata = %q, want balanced", got)
	}
}

func TestMem0ClientSendsAPIKeyOnlyWhenConfigured(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name       string
		apiKey     string
		wantHeader string
	}{
		{name: "absent", apiKey: ""},
		{name: "configured", apiKey: "secret", wantHeader: "secret"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("X-API-Key"); got != test.wantHeader {
					t.Errorf("X-API-Key = %q, want %q", got, test.wantHeader)
				}
				if _, err := w.Write([]byte(`{"results":[]}`)); err != nil {
					t.Errorf("write search response: %v", err)
				}
			}))
			defer server.Close()

			client := mem0Client{baseURL: server.URL, apiKey: test.apiKey, httpClient: server.Client()}
			if _, err := client.Search(t.Context(), searchRequest{}); err != nil {
				t.Fatalf("Search() error = %v", err)
			}
		})
	}
}

func TestMem0ClientDoesNotFollowRedirects(t *testing.T) {
	t.Parallel()

	var targetVisited atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		targetVisited.Store(true)
		t.Errorf("redirect target received request with X-API-Key %q", r.Header.Get("X-API-Key"))
	}))
	defer target.Close()

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer source.Close()

	client := mem0Client{baseURL: source.URL, apiKey: "secret", httpClient: source.Client()}
	_, err := client.Search(t.Context(), searchRequest{})
	if err == nil || !strings.Contains(err.Error(), "unexpected status 302 Found") {
		t.Fatalf("Search() error = %v, want redirect status error", err)
	}
	if targetVisited.Load() {
		t.Fatal("redirect target was visited")
	}
}

type recordingModel struct {
	request gopact.ModelRequest
}

func (m *recordingModel) NewRequest(messages ...gopact.Message) gopact.ModelRequest {
	return gopact.ModelRequest{
		Messages: append([]gopact.Message(nil), messages...),
		Metadata: map[string]string{"provider.request_profile": "balanced"},
	}
}

func (m *recordingModel) Invoke(_ context.Context, request gopact.ModelRequest, _ ...gopact.ModelCallOption) (gopact.ModelResponse, error) {
	m.request = request
	return gopact.ModelResponse{Message: gopact.Message{
		Role:  "assistant",
		Parts: []gopact.MessagePart{{Type: "text", Text: "PostgreSQL is a good fit."}},
	}}, nil
}
