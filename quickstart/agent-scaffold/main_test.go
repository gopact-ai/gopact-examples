package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunShowsScaffoldApprovalResume(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"first_events: run_started -> node_started(plan) -> node_completed(plan) -> node_started(write) -> node_completed(write) -> node_started(approval) -> interrupted(approval) -> run_interrupted(approval)",
		"pending: approval checkpoint=scaffold-first:3",
		"resume_events: run_started -> checkpoint_loaded(approval) -> resume_received(approval) -> node_resumed(summary) -> node_completed(summary) -> run_completed",
		"trace: plan -> write -> approval -> summary",
		"summary: published draft for add a README example",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}
