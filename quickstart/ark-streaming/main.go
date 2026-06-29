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

	for event, err := range model.Stream(ctx, gopact.NewModelRequest(
		gopact.WithMessages(gopact.UserMessage("Count from one to three, separated by commas.")),
		gopact.WithMaxOutputTokens(64),
		gopact.WithTemperature(0.2),
	)) {
		if err != nil {
			return err
		}
		if event.Message != nil {
			_, _ = io.WriteString(out, event.Message.Text())
		}
	}
	_, err = fmt.Fprintln(out)
	return err
}
