package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact/workflow"
)

const exampleRunID = "durable-run-42"

type exampleResult struct {
	output       string
	resumedRunID string
}

func runExample(ctx context.Context) (exampleResult, error) {
	store := workflow.NewMemoryStore()
	wf := workflow.New[string, string]("durable-resume", workflow.WithStore(store))
	process := wf.Node("process", func(ctx context.Context, input string) (string, error) {
		return "processed:" + input, nil
	})
	process.Guard(workflow.BeforeRun("approval", workflow.GuardFunc[string, string](
		func(context.Context, workflow.GuardContext[string, string]) (workflow.GuardDecision[string, string], error) {
			return workflow.GuardInterrupt[string, string]{Request: workflow.InterruptRequest{
				ID:                  "approval-1",
				Subject:             "approve processing",
				ResolutionSchemaRef: "schema://approval",
			}}, nil
		},
	)))
	wf.Entry(process)
	wf.Exit(process)

	_, err := wf.Invoke(ctx, "order-42", gopact.WithRunID(exampleRunID))
	var interrupt workflow.InterruptError
	if !errors.As(err, &interrupt) {
		return exampleResult{}, fmt.Errorf("initial invoke: got %v, want interrupt", err)
	}

	output, err := wf.Invoke(ctx, "", workflow.WithResume(workflow.ResumeRequest{
		RunID:        exampleRunID,
		CheckpointID: interrupt.CheckpointID,
		Resolutions: []workflow.InterruptResolution{{
			InterruptID: "approval-1",
			PayloadRef:  "artifact://approved",
		}},
	}))
	if err != nil {
		return exampleResult{}, fmt.Errorf("resume run: %w", err)
	}
	snapshot, err := wf.Snapshot(ctx, workflow.SnapshotRequest{RunID: exampleRunID})
	if err != nil {
		return exampleResult{}, fmt.Errorf("load completed run: %w", err)
	}
	return exampleResult{output: output, resumedRunID: snapshot.RunMeta.RunID}, nil
}

func main() {
	result, err := runExample(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("run=%s output=%s\n", result.resumedRunID, result.output)
}
