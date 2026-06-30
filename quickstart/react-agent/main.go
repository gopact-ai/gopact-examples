package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gopact-ai/gopact"
	react "github.com/gopact-ai/gopact-ext/agents/react"
	"github.com/gopact-ai/gopact/tools"
)

type scriptedModel struct {
	responses []gopact.Message
}

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	registry := tools.NewRegistry()
	if err := registry.Register(ctx, uppercaseTool(), tools.RegisterOptions{
		Namespace:  "local",
		Visibility: tools.VisibleTool,
	}); err != nil {
		return err
	}

	agent, err := react.New(&scriptedModel{
		responses: []gopact.Message{
			{
				Role: gopact.RoleAssistant,
				ToolCalls: []gopact.ToolCall{{
					ID:        "call-1",
					Name:      "local.uppercase",
					Arguments: []byte(`{"text":"gopact"}`),
				}},
			},
			gopact.AssistantMessage("final answer: GOPACT"),
		},
	}, registry)
	if err != nil {
		return err
	}

	var events []string
	var toolResult string
	var final string
	for event, err := range agent.Run(ctx, react.State{
		Messages: []gopact.Message{gopact.UserMessage("Use the uppercase tool on gopact.")},
	}, gopact.WithRuntimeIDs(gopact.RuntimeIDs{RunID: "react-demo"})) {
		if err != nil {
			return err
		}
		events = append(events, eventLabel(event))
		if event.Result != nil && event.Result.Content != "" {
			toolResult = event.Result.Content
		}
		if event.Message != nil && event.Message.Text() != "" {
			final = event.Message.Text()
		}
	}

	if _, err := fmt.Fprintf(out, "events: %s\n", strings.Join(events, " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "tool_result: %s\n", toolResult); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "final: %s\n", final)
	return err
}

func (m *scriptedModel) Generate(ctx context.Context, request gopact.ModelRequest) (gopact.Message, error) {
	if err := ctx.Err(); err != nil {
		return gopact.Message{}, err
	}
	if len(m.responses) == 0 {
		return gopact.AssistantMessage("done"), nil
	}
	response := m.responses[0]
	m.responses = m.responses[1:]
	_ = request
	return response, nil
}

func uppercaseTool() gopact.Tool {
	return gopact.ToolFunc{
		SpecValue: gopact.ObjectToolSpec(
			"uppercase",
			"Uppercase a text string.",
			gopact.RequiredStringField("text", "Text to uppercase."),
		),
		InvokeFunc: func(_ context.Context, args json.RawMessage) (gopact.ToolResult, error) {
			var input struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return gopact.ToolResult{}, err
			}
			return gopact.ToolResult{Content: strings.ToUpper(input.Text)}, nil
		},
	}
}

func eventLabel(event gopact.Event) string {
	if event.Node == "" {
		return string(event.Type)
	}
	return fmt.Sprintf("%s(%s)", event.Type, event.Node)
}
