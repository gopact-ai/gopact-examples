package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunShowsPlanExecuteFlow(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"events: run_started -> node_started(plan) -> node_completed(plan) -> node_started(execute) -> node_completed(execute) -> node_started(summarize) -> node_completed(summarize) -> run_completed",
		"results: draft=done draft, review=done review",
		"summary: completed 2 steps",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}
