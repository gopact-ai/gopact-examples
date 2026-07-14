package main

import (
	"context"
	"testing"
)

func TestRunExampleResumesInterruptedRun(t *testing.T) {
	result, err := runExample(context.Background())
	if err != nil {
		t.Fatalf("runExample() error = %v", err)
	}
	if result.output != "processed:order-42" {
		t.Fatalf("output = %q, want processed:order-42", result.output)
	}
	if result.resumedRunID != exampleRunID {
		t.Fatalf("resumed RunID = %q, want %q", result.resumedRunID, exampleRunID)
	}
}
