package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-ext/agents/planexec"
	"github.com/gopact-ai/gopact-ext/agents/supervisor"
	"github.com/gopact-ai/gopact/gopacttest"
)

func TestRunShowsSupervisorRoutingFlow(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"selected: writer",
		"child_agent: writer",
		"child_trace: plan -> execute -> summarize",
		"child_results: outline=writer done outline, polish=writer done polish",
		"child_summary: completed 2 steps",
		"events: run_started -> node_started(route) -> node_completed(route) -> run_started -> node_started(plan) -> node_completed(plan) -> node_started(execute) -> node_completed(execute) -> node_started(summarize) -> node_completed(summarize) -> run_completed -> run_completed",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}

func TestSupervisorPassesChildAgentRuntimeID(t *testing.T) {
	child, err := planExecChild("reviewer", []planexec.Step{{ID: "check", Instruction: "Check the example"}})
	if err != nil {
		t.Fatalf("planExecChild() error = %v", err)
	}
	agent, err := supervisor.New(
		supervisor.RouterFunc(func(context.Context, supervisor.Request) (supervisor.Route, error) {
			return supervisor.Route{Agent: "reviewer", Input: "review the example"}, nil
		}),
		supervisor.Child{Name: "reviewer", Runnable: child},
	)
	if err != nil {
		t.Fatalf("supervisor.New() error = %v", err)
	}

	events, err := gopacttest.CollectEvents(agent.Run(context.Background(), "review the example",
		gopact.WithRuntimeIDs(gopact.RuntimeIDs{RunID: "supervisor-test", ThreadID: "thread-1"}),
	))
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	for _, event := range events {
		if event.IDs.AgentID == "reviewer" {
			return
		}
	}
	t.Fatalf("events = %+v, want child events with reviewer agent id", events)
}

func TestSupervisorFailsOnUnknownChild(t *testing.T) {
	agent, err := supervisor.New(
		supervisor.RouterFunc(func(context.Context, supervisor.Request) (supervisor.Route, error) {
			return supervisor.Route{Agent: "missing"}, nil
		}),
		supervisor.Child{Name: "reviewer", Runnable: reviewerChild},
	)
	if err != nil {
		t.Fatalf("supervisor.New() error = %v", err)
	}

	events, err := gopacttest.CollectEvents(agent.Run(context.Background(), "review the example"))
	if !errors.Is(err, supervisor.ErrRouteAgentUnknown) {
		t.Fatalf("Run() error = %v, want unknown child", err)
	}
	gopacttest.RequireEventTypes(t, events,
		gopact.EventRunStarted,
		gopact.EventNodeStarted,
		gopact.EventRunFailed,
	)
}

var reviewerChild = func() *planexec.Agent {
	child, err := planExecChild("reviewer", []planexec.Step{{ID: "check", Instruction: "Check the example"}})
	if err != nil {
		panic(err)
	}
	return child
}()
