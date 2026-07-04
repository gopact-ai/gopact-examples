package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunShowsHumanReviewResumePaths(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"first_events: run_started -> node_started(draft) -> node_completed(draft) -> node_started(review) -> interrupted(review) -> run_interrupted(review)",
		"pending: review:publish required_by=release template=humanreview",
		"step_export_resume: run_started -> step_imported(review) -> resume_received(review) -> node_resumed(publish) -> node_completed(publish) -> run_completed",
		"checkpoint_resume: run_started -> checkpoint_loaded(review) -> resume_received(review) -> node_resumed(publish) -> node_completed(publish) -> run_completed checkpoint=human-review-checkpoint-first:2",
		"step_trace: draft -> publish",
		"checkpoint_trace: draft -> publish",
		"summary: published draft for release notes",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}
