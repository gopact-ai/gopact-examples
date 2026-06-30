package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-examples/internal/exampleenv"
	"github.com/gopact-ai/gopact-ext/models/openai"
)

const (
	arkStreamingMaxOutputTokens = 1024
	arkStreamingPrompt          = "Count from one to three, separated by commas."
)

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	cfg, err := exampleenv.LoadArkOpenAIConfig()
	if err != nil {
		return err
	}

	model, err := openai.NewClient(
		openai.ProviderArk,
		cfg.BaseURL,
		cfg.Token,
		openai.WithResponsesAPI(),
		gopact.WithModel(cfg.Model),
		gopact.EnableStreaming(),
	)
	if err != nil {
		return err
	}

	if err := stream(ctx, out, model, "thinking=disabled", openai.DisableThinking()); err != nil {
		return err
	}
	return stream(ctx, out, model, "thinking=enabled", openai.EnableThinking())
}

func stream(ctx context.Context, out io.Writer, model gopact.StreamingModel, label string, thinking gopact.ModelRequestOption) error {
	if _, err := fmt.Fprintf(out, "%s: ", label); err != nil {
		return err
	}
	for event, err := range model.Stream(ctx, gopact.NewModelRequest(
		gopact.WithMessages(gopact.UserMessage(arkStreamingPrompt)),
		gopact.WithMaxOutputTokens(arkStreamingMaxOutputTokens),
		gopact.WithTemperature(0.2),
		thinking,
	)) {
		if err != nil {
			return err
		}
		if event.Message != nil && event.Message.Text() != "" {
			if _, err := io.WriteString(out, event.Message.Text()); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintln(out)
	return err
}
