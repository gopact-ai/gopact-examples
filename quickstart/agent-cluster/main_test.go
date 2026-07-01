package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunShowsLocalAgentCluster(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"gateway: accepted self-bootstrap slice",
		"discovery: 4 HTTP agent cards",
		"cards: planner-agent, research-agent, code-agent, review-agent",
		"workflow events: run_started -> node_started(planner-agent) -> node_completed(planner-agent) -> node_started(research-agent) -> node_completed(research-agent) -> node_started(code-agent) -> node_completed(code-agent) -> node_started(review-agent) -> node_completed(review-agent) -> run_completed",
		"run export: completed events=10 steps=4 verification_reports=1",
		"git diff evidence:",
		"release gate: passed checks=",
		"requirements=12",
		"checkpoint resume: loaded review-agent step=4 events=run_started -> checkpoint_loaded(review-agent) -> run_completed",
		"planner-agent: plan: research -> code -> review",
		"research-agent: research: graph, a2a, examples",
		"artifacts: plan.md(memory://planner-plan) -> research.md(memory://research-notes) -> patch.diff(memory://code-patch)",
		"code-agent: code: prepare a small tested patch",
		"review stream: running(reviewing evidence) -> completed(review: pass)",
		"policy events: policy_requested -> policy_decided",
		"failure attribution: external missing-agent check=failure-attribution:missing-agent",
		"agent trace: planner-agent -> research-agent -> code-agent -> review-agent",
		"summary: local agent cluster completed 4 calls",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}
