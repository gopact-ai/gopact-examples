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
	cfg, err := exampleenv.LoadArkConfig()
	if err != nil {
		return err
	}

	model, err := openai.New(openai.Options{
		Provider: "ark",
		BaseURL:  cfg.BaseURL,
		APIKey:   cfg.Token,
		API:      openai.APIResponses,
	})
	if err != nil {
		return err
	}

	for event, err := range model.Stream(ctx, gopact.ModelRequest{
		Model: cfg.Model,
		Messages: []gopact.Message{{
			Role:    gopact.RoleUser,
			Content: "Count from one to three, separated by commas.",
		}},
	}) {
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
