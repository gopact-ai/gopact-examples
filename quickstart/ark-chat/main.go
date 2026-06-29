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

	model, err := openai.NewClient(
		openai.ProviderArk,
		cfg.BaseURL,
		cfg.Token,
		openai.WithResponsesAPI(),
		gopact.WithModel(cfg.Model),
	)
	if err != nil {
		return err
	}

	response, err := model.Generate(ctx, gopact.NewModelRequest(
		gopact.WithMessages(
			gopact.Message{Role: gopact.RoleSystem, Content: "You are a concise assistant."},
			gopact.Message{Role: gopact.RoleUser, Content: "Say hello from gopact and Ark in one sentence."},
		),
	))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(out, response.Message.Text())
	return err
}
