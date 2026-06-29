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

	model, err := openai.NewClient(
		openai.ProviderOpenAI,
		cfg.BaseURL,
		cfg.Token,
		gopact.WithModel(cfg.Model),
		gopact.EnableToolCalling(),
	)
	if err != nil {
		return err
	}

	messages := []gopact.Message{
		gopact.SystemMessage("Use tools when they are available, then answer briefly."),
		gopact.UserMessage("Use the uppercase tool on the text gopact."),
	}
	first, err := model.Generate(ctx, gopact.NewModelRequest(
		gopact.WithMessages(messages...),
		gopact.WithTools(gopact.ObjectToolSpec(
			"uppercase",
			"Uppercase a text string.",
			gopact.RequiredStringField("text", "Text to uppercase."),
		)),
	))
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
		messages = append(messages, gopact.ToolMessage(call.ID, result))
	}

	final, err := model.Generate(ctx, gopact.NewModelRequest(gopact.WithMessages(messages...)))
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
