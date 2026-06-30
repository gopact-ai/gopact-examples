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

const openAIStreamingMaxOutputTokens = 512

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

	if _, err := io.WriteString(out, "chat_completions: "); err != nil {
		return err
	}
	if err := stream(ctx, out, cfg, openai.WithChatCompletionsAPI()); err != nil {
		return err
	}
	if _, err := io.WriteString(out, "\nresponses: "); err != nil {
		return err
	}
	if err := stream(ctx, out, cfg, openai.WithResponsesAPI()); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out)
	return err
}

func stream(ctx context.Context, out io.Writer, cfg exampleenv.Config, api openai.Option) error {
	model, err := openai.NewClient(
		openai.ProviderOpenAI,
		cfg.BaseURL,
		cfg.Token,
		api,
		gopact.WithModel(cfg.Model),
		gopact.EnableStreaming(),
	)
	if err != nil {
		return err
	}

	for event, err := range model.Stream(ctx, gopact.NewModelRequest(
		gopact.WithMessages(gopact.UserMessage("Count from one to three, separated by commas.")),
		gopact.WithMaxOutputTokens(openAIStreamingMaxOutputTokens),
		gopact.WithTemperature(0.2),
	)) {
		if err != nil {
			return err
		}
		if event.Message != nil {
			if _, err := io.WriteString(out, event.Message.Text()); err != nil {
				return err
			}
		}
	}
	return nil
}
