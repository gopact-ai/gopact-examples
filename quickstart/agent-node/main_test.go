package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunShowsAgentNodeDelegation(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"plan: plan: research -> code -> review",
		"child_evidence: a2a_task_completed(planner-agent, planner-task)",
		"events: run_started -> node_started(delegate) -> a2a_task_status_updated -> a2a_task_completed -> node_completed(delegate) -> run_completed",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}
