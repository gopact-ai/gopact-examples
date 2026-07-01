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
		"steps: plan -> draft -> review -> summarize",
		"summary: workflow completed 2 parallel actions",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}
