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
		gopact.EnableStructuredOutput(),
	)
	if err != nil {
		return err
	}

	schema := outputSchema()
	response, err := model.Generate(ctx, gopact.NewModelRequest(
		gopact.WithMessages(
			gopact.SystemMessage("Return only JSON that satisfies the requested schema."),
			gopact.UserMessage("Return status ok and a short summary for gopact structured output."),
		),
		gopact.WithResponseSchema(schema),
		gopact.WithMaxOutputTokens(512),
		gopact.WithTemperature(0.1),
		gopact.EnableStructuredOutput(),
	))
	if err != nil {
		return err
	}

	var payload struct {
		Status  string `json:"status"`
		Summary string `json:"summary"`
	}
	raw := strings.TrimSpace(response.Message.Text())
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return fmt.Errorf("structured output json: %w", err)
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return fmt.Errorf("structured output value: %w", err)
	}
	if err := gopact.ValidateJSONSchemaValue(schema, value); err != nil {
		return fmt.Errorf("structured output schema: %w", err)
	}

	_, err = fmt.Fprintf(out, "status=%s summary=%s\n", payload.Status, payload.Summary)
	return err
}

func outputSchema() gopact.JSONSchema {
	return gopact.JSONSchema{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []any{"status", "summary"},
		"properties": map[string]any{
			"status":  map[string]any{"type": "string", "const": "ok"},
			"summary": map[string]any{"type": "string", "minLength": 1},
		},
	}
}
