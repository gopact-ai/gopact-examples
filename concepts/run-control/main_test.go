package main

import (
	"context"
	"testing"

	"github.com/gopact-ai/gopact/workflow"
)

func TestRunExampleCreatesNewRunsFromFailedSource(t *testing.T) {
	result, err := runExample(context.Background())
	if err != nil {
		t.Fatalf("runExample() error = %v", err)
	}
	if result.retryOutput != "processed:original" || result.forkOutput != "processed:forked" {
		t.Fatalf("retry/fork output = %q/%q", result.retryOutput, result.forkOutput)
	}
	if result.sourceStatus != workflow.CheckpointFailed {
		t.Fatalf("source status = %q, want failed", result.sourceStatus)
	}
	if result.retrySourceRunID != sourceRunID || result.forkSourceRunID != sourceRunID {
		t.Fatalf(
			"retry/fork source = %q/%q, want %q",
			result.retrySourceRunID,
			result.forkSourceRunID,
			sourceRunID,
		)
	}
}
