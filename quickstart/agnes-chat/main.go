package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-examples/internal/exampleenv"
	"github.com/gopact-ai/gopact-ext/models/agnes"
)

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	cfg, err := exampleenv.LoadAgnesConfig()
	if err != nil {
		return err
	}

	model, err := agnes.NewClient(
		cfg.BaseURL,
		cfg.Token,
		gopact.WithModel(cfg.Model),
		agnes.DisableThinking(),
	)
	if err != nil {
		return err
	}

	response, err := model.Generate(ctx, gopact.NewModelRequest(
		gopact.WithMessages(
			gopact.SystemMessage("You are a concise assistant."),
			gopact.UserMessage("Say hello from Agnes through gopact in one sentence."),
		),
		gopact.WithMaxOutputTokens(512),
		gopact.WithTemperature(0.2),
		agnes.DisableThinking(),
	))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(out, response.Message.Text())
	return err
}
