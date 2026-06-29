package main

import (
	"context"
	"fmt"
	"io"
	"os"

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
			Name:     cfg.Model,
			Provider: "llm",
		}},
	})
	if err != nil {
		return err
	}

	response, err := model.Generate(ctx, gopact.ModelRequest{
		Model: cfg.Model,
		Messages: []gopact.Message{
			{Role: gopact.RoleSystem, Content: "You are a concise assistant."},
			{Role: gopact.RoleUser, Content: "Say hello from gopact in one sentence."},
		},
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(out, response.Message.Text())
	return err
}
