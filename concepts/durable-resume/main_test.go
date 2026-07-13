package main

import (
	"context"
	"strings"
	"testing"
)

func TestRunExampleDeduplicatesRecoveredSideEffect(t *testing.T) {
	result, err := runExample(context.Background())
	if err != nil {
		t.Fatalf("runExample() error = %v", err)
	}
	if result.output != "charged:order-42" {
		t.Fatalf("output = %q, want charged:order-42", result.output)
	}
	if result.nodeRuns != 2 || result.effectAttempts != 2 {
		t.Fatalf(
			"node runs/effect attempts = %d/%d, want 2/2 after recovery",
			result.nodeRuns,
			result.effectAttempts,
		)
	}
	if result.appliedEffects != 1 {
		t.Fatalf("applied effects = %d, want 1", result.appliedEffects)
	}
	if len(result.idempotencyKeys) != 2 || result.idempotencyKeys[0] != result.idempotencyKeys[1] {
		t.Fatalf("idempotency keys = %v, want one stable key across both attempts", result.idempotencyKeys)
	}
	if !strings.HasPrefix(result.idempotencyKeys[0], exampleRunID+"/") ||
		strings.TrimPrefix(result.idempotencyKeys[0], exampleRunID+"/") == "" {
		t.Fatalf(
			"idempotency key = %q, want RunInfo.RunID/RunInfo.ActivationID",
			result.idempotencyKeys[0],
		)
	}
}
