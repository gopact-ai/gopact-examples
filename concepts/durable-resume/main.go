package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact/workflow"
)

const (
	exampleRunID            = "durable-run-42"
	sideEffectReturnedEvent = "example.side_effect_returned"
)

var errSimulatedProcessLoss = errors.New("simulated process loss after side effect")

type exampleResult struct {
	output          string
	nodeRuns        int
	effectAttempts  int
	appliedEffects  int
	idempotencyKeys []string
}

// idempotentChargeAPI is an in-memory stand-in for an external API that
// natively deduplicates requests by idempotency key.
type idempotentChargeAPI struct {
	attempts int
	receipts map[string]string
}

func newIdempotentChargeAPI() *idempotentChargeAPI {
	return &idempotentChargeAPI{receipts: map[string]string{}}
}

func (api *idempotentChargeAPI) charge(key, orderID string) string {
	api.attempts++
	if receipt, exists := api.receipts[key]; exists {
		return receipt
	}
	receipt := "charged:" + orderID
	api.receipts[key] = receipt
	return receipt
}

func runExample(ctx context.Context) (exampleResult, error) {
	store := workflow.NewMemoryStore()
	charges := newIdempotentChargeAPI()
	nodeRuns := 0
	idempotencyKeys := []string{}

	wf := workflow.New[string, string](
		"durable-resume",
		workflow.WithStore(store),
	)
	charge := wf.Node("charge", func(ctx context.Context, orderID string) (string, error) {
		info := workflow.RunInfoFromContext(ctx)
		if info.RunID == "" || info.ActivationID == "" {
			return "", errors.New("durable-resume: workflow run identity is unavailable")
		}
		// This key represents recovery of the same logical activation. A business
		// retry that intentionally charges again must use a new operation key.
		idempotencyKey := info.RunID + "/" + info.ActivationID
		nodeRuns++
		idempotencyKeys = append(idempotencyKeys, idempotencyKey)
		receipt := charges.charge(idempotencyKey, orderID)

		// A strict sink failure here deterministically models losing the process
		// after the external side effect succeeds but before node completion is durable.
		if err := workflow.Emit(ctx, gopact.Event{
			Type:    sideEffectReturnedEvent,
			Summary: "external side effect returned successfully",
		}); err != nil {
			return "", err
		}
		return receipt, nil
	})
	wf.Entry(charge)
	wf.Exit(charge)

	_, err := wf.Invoke(
		ctx,
		"order-42",
		gopact.WithRunID(exampleRunID),
		gopact.WithStrictEventHandler(func(_ context.Context, event gopact.Event) error {
			if event.Type == sideEffectReturnedEvent {
				return errSimulatedProcessLoss
			}
			return nil
		}),
	)
	if !errors.Is(err, errSimulatedProcessLoss) {
		return exampleResult{}, fmt.Errorf("initial invoke: got %v, want simulated process loss", err)
	}

	output, err := wf.Invoke(
		ctx,
		"",
		workflow.WithResume(workflow.ResumeRequest{RunID: exampleRunID}),
	)
	if err != nil {
		return exampleResult{}, fmt.Errorf("resume run: %w", err)
	}
	return exampleResult{
		output:          output,
		nodeRuns:        nodeRuns,
		effectAttempts:  charges.attempts,
		appliedEffects:  len(charges.receipts),
		idempotencyKeys: idempotencyKeys,
	}, nil
}

func main() {
	result, err := runExample(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf(
		"node_runs=%d effect_attempts=%d applied_effects=%d key=%s output=%s\n",
		result.nodeRuns,
		result.effectAttempts,
		result.appliedEffects,
		result.idempotencyKeys[0],
		result.output,
	)
}
