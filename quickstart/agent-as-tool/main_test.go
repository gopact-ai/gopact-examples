package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunShowsAgentAsTool(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"parent: delegated",
		"child_result: completed 1 steps",
		"child_evidence: a2a_task_completed(planexec-child, child-task-1)",
		"events: run_started -> node_started(call_model) -> model_message(call_model) -> node_completed(call_model) -> node_started(call_tool) -> tool_call(call_tool) -> a2a_task_completed(call_tool) -> tool_result(call_tool) -> node_completed(call_tool) -> node_started(call_model) -> model_message(call_model) -> node_completed(call_model) -> run_completed",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}
