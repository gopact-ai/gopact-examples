package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-examples/internal/exampleenv"
	"github.com/gopact-ai/gopact-ext/models/openai"
	"github.com/gopact-ai/gopact/provider"
)

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	cfg, err := exampleenv.LoadConfig()
	if err != nil {
		return err
	}

	model, err := openai.New(openai.Options{
		Provider: "llm",
		BaseURL:  cfg.BaseURL,
		APIKey:   cfg.Token,
		Models: []provider.ModelInfo{{
			Name:         cfg.Model,
			Provider:     "llm",
			Capabilities: []provider.Capability{provider.CapabilityToolCalling},
		}},
	})
	if err != nil {
		return err
	}

	messages := []gopact.Message{
		{Role: gopact.RoleSystem, Content: "Use tools when they are available, then answer briefly."},
		{Role: gopact.RoleUser, Content: "Use the uppercase tool on the text gopact."},
	}
	first, err := model.Generate(ctx, gopact.ModelRequest{
		Model:    cfg.Model,
		Messages: messages,
		Tools: []gopact.ToolSpec{{
			Name:        "uppercase",
			Description: "Uppercase a text string.",
			InputSchema: gopact.JSONSchema{
				"type":     "object",
				"required": []any{"text"},
				"properties": map[string]any{
					"text": map[string]any{"type": "string"},
				},
			},
		}},
	})
	if err != nil {
		return err
	}
	if len(first.Message.ToolCalls) == 0 {
		_, err = fmt.Fprintln(out, first.Message.Text())
		return err
	}

	messages = append(messages, first.Message)
	for _, call := range first.Message.ToolCalls {
		result, err := invokeTool(call)
		if err != nil {
			return err
		}
		messages = append(messages, gopact.Message{
			Role:       gopact.RoleTool,
			ToolCallID: call.ID,
			Content:    result,
		})
	}

	final, err := model.Generate(ctx, gopact.ModelRequest{
		Model:    cfg.Model,
		Messages: messages,
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, final.Message.Text())
	return err
}

func invokeTool(call gopact.ToolCall) (string, error) {
	switch call.Name {
	case "uppercase":
		var args struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			return "", err
		}
		return strings.ToUpper(args.Text), nil
	default:
		return "", fmt.Errorf("unknown tool %q", call.Name)
	}
}
