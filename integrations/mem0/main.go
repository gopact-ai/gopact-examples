package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strings"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-ext/models/fake"
	"github.com/gopact-ai/gopact/workflow"
)

type mem0Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type searchRequest struct {
	Query   string `json:"query"`
	UserID  string `json:"user_id"`
	AgentID string `json:"agent_id"`
	RunID   string `json:"run_id"`
}

type searchResponse struct {
	Results []memoryResult `json:"results"`
}

type memoryResult struct {
	Memory string `json:"memory"`
}

type workflowInput struct {
	Query       string
	UserID      string
	AgentID     string
	UserMessage gopact.Message
}

type contextInput struct {
	Model         gopact.Model
	RunID         string
	MemorySummary string
	UserMessage   gopact.Message
}

type memoryContext struct {
	RunID         string
	MemorySummary string
	UserMessage   gopact.Message
}

type memorySearchFunc func(context.Context, searchRequest) (string, error)

func (c mem0Client) Search(ctx context.Context, search searchRequest) (string, error) {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(search); err != nil {
		return "", fmt.Errorf("encode mem0 search: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.baseURL, "/")+"/search", &body)
	if err != nil {
		return "", fmt.Errorf("create mem0 search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	client := c.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	safeClient := *client
	safeClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := safeClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("search mem0: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("search mem0: unexpected status %s", resp.Status)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode mem0 search: %w", err)
	}
	memories := make([]string, 0, len(result.Results))
	for _, item := range result.Results {
		memories = append(memories, item.Memory)
	}
	return strings.Join(memories, "\n"), nil
}

func buildModelRequest(input contextInput) gopact.ModelRequest {
	request := input.Model.NewRequest(
		gopact.Message{
			Role:  "system",
			Parts: []gopact.MessagePart{{Type: "text", Text: input.MemorySummary}},
		},
		input.UserMessage,
	)
	request.Metadata = maps.Clone(request.Metadata)
	if request.Metadata == nil {
		request.Metadata = map[string]string{}
	}
	request.Metadata["gopact.workflow.run_id"] = input.RunID
	return request
}

func runMemoryWorkflow(ctx context.Context, search memorySearchFunc, model gopact.Model, input workflowInput, opts ...gopact.RunOption) (string, error) {
	wf := workflow.New[workflowInput, string]("memory-context")
	loadMemory := wf.Node("load-memory", func(ctx context.Context, input workflowInput) (memoryContext, error) {
		info := workflow.RunInfoFromContext(ctx)
		if info.SessionID == "" || info.RunID == "" {
			return memoryContext{}, fmt.Errorf("memory context: workflow run identity is missing")
		}
		memory, err := search(ctx, searchRequest{
			Query:   input.Query,
			UserID:  input.UserID,
			AgentID: input.AgentID,
			RunID:   info.SessionID,
		})
		if err != nil {
			return memoryContext{}, err
		}
		return memoryContext{
			RunID:         info.RunID,
			MemorySummary: memory,
			UserMessage:   input.UserMessage,
		}, nil
	})
	buildRequest := wf.Node("build-model-request", func(_ context.Context, input memoryContext) (gopact.ModelRequest, error) {
		return buildModelRequest(contextInput{
			Model:         model,
			RunID:         input.RunID,
			MemorySummary: input.MemorySummary,
			UserMessage:   input.UserMessage,
		}), nil
	})
	invokeModel := wf.Node("model", func(ctx context.Context, request gopact.ModelRequest) (string, error) {
		response, err := model.Invoke(ctx, request)
		if err != nil {
			return "", err
		}
		for _, part := range response.Message.Parts {
			if part.Type == "text" {
				return part.Text, nil
			}
		}
		return "", fmt.Errorf("model response has no text part")
	})
	wf.Entry(loadMemory)
	wf.Edge(loadMemory, buildRequest)
	wf.Edge(buildRequest, invokeModel)
	wf.Exit(invokeModel)

	return wf.Invoke(ctx, input, opts...)
}

func main() {
	const (
		sessionID = "session-memory-9"
		runID     = "run-memory-3"
	)
	input := workflowInput{
		Query:       "Which database does this user prefer?",
		UserID:      "user-7",
		AgentID:     "support-agent",
		UserMessage: gopact.UserMessage("Recommend a database for my next service."),
	}
	answer, err := runMemoryWorkflow(
		context.Background(),
		func(context.Context, searchRequest) (string, error) {
			return "Prefers PostgreSQL for transactional data.\nUses pgvector for semantic search.", nil
		},
		fake.New(fake.WithResponse("PostgreSQL is a good fit.")),
		input,
		gopact.WithSessionID(sessionID),
		gopact.WithRunID(runID),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf(
		"session=%s run=%s answer=%s\n",
		sessionID,
		runID,
		answer,
	)
}
