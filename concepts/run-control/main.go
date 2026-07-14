package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact/workflow"
)

const (
	sourceRunID = "failed-run"
	retryRunID  = "retry-run"
	forkRunID   = "fork-run"
)

type exampleResult struct {
	retryOutput      string
	forkOutput       string
	sourceStatus     workflow.CheckpointStatus
	retryRunID       string
	forkRunID        string
	retrySourceRunID string
	forkSourceRunID  string
}

func runExample(ctx context.Context) (exampleResult, error) {
	store := workflow.NewMemoryStore()
	failOnce := true
	wf := workflow.New[string, string]("run-control", workflow.WithStore(store))
	process := wf.Node("process", func(_ context.Context, input string) (string, error) {
		if failOnce {
			failOnce = false
			return "", errors.New("temporary failure")
		}
		return "processed:" + input, nil
	})
	wf.Entry(process)
	wf.Exit(process)

	if _, err := wf.Invoke(ctx, "original", gopact.WithRunID(sourceRunID)); err == nil {
		return exampleResult{}, errors.New("source invoke: expected failure")
	}
	snapshot, err := wf.Snapshot(ctx, workflow.SnapshotRequest{RunID: sourceRunID})
	if err != nil {
		return exampleResult{}, fmt.Errorf("load failed run: %w", err)
	}

	var nodeID string
	var nodeExecutionVersion int64
	var rootSequence int64
	for _, event := range snapshot.Timeline {
		if event.EventType == workflow.EventNodeStarted {
			nodeID = event.NodeID
			nodeExecutionVersion = event.NodeExecutionVersion
		}
	}
	for _, checkpoint := range snapshot.Checkpoints {
		if checkpoint.Root && checkpoint.ReplayStatus == workflow.ReplaySafe {
			rootSequence = checkpoint.EventSeq
			break
		}
	}
	if nodeID == "" || nodeExecutionVersion == 0 || rootSequence == 0 {
		return exampleResult{}, errors.New("failed run has no retry or fork point")
	}

	retryOutput, err := wf.Retry(ctx, workflow.RetryRequest{
		RunID:                sourceRunID,
		NodeID:               nodeID,
		NodeExecutionVersion: nodeExecutionVersion,
	}, gopact.WithRunID(retryRunID))
	if err != nil {
		return exampleResult{}, fmt.Errorf("retry failed run: %w", err)
	}
	forkOutput, err := snapshot.Fork(ctx, wf, workflow.ForkRequest{
		SourceRunID:  sourceRunID,
		FromEventSeq: rootSequence,
		Patch: workflow.ForkPatch{
			WorkflowInput: &workflow.InputPatch{Value: "forked"},
		},
	}, gopact.WithRunID(forkRunID))
	if err != nil {
		return exampleResult{}, fmt.Errorf("fork failed run: %w", err)
	}

	source, err := store.Load(ctx, sourceRunID)
	if err != nil {
		return exampleResult{}, fmt.Errorf("load source: %w", err)
	}
	retry, err := store.Load(ctx, retryRunID)
	if err != nil {
		return exampleResult{}, fmt.Errorf("load retry: %w", err)
	}
	fork, err := store.Load(ctx, forkRunID)
	if err != nil {
		return exampleResult{}, fmt.Errorf("load fork: %w", err)
	}
	return exampleResult{
		retryOutput: retryOutput, forkOutput: forkOutput, sourceStatus: source.Status,
		retryRunID: retry.RunID, forkRunID: fork.RunID,
		retrySourceRunID: retry.SourceRunID, forkSourceRunID: fork.SourceRunID,
	}, nil
}

func main() {
	result, err := runExample(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf(
		"source=%s status=%s retry=%s:%s fork=%s:%s\n",
		sourceRunID,
		result.sourceStatus,
		result.retryRunID,
		result.retryOutput,
		result.forkRunID,
		result.forkOutput,
	)
}
