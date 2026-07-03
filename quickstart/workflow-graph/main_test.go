package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunShowsWorkflowEventsAndSummary(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"steps: plan -> draft -> review -> polish-start -> polish-finish -> refine-1 -> refine-2 -> summarize",
		"nested events: node_started(polish-start) -> node_completed(polish-start) -> node_started(polish-finish) -> node_completed(polish-finish)",
		"step limit: graph: exceeded max steps 2",
		"step export resume: step_imported -> node_resumed -> node_completed -> run_completed",
		"interrupt resume: step_imported -> resume_received -> node_resumed -> node_completed -> run_completed",
		"summary: workflow completed 2 parallel actions after 2 refinements",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}
