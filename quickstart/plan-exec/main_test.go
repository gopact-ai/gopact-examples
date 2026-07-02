package main

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-ext/agents/planexec"
	"github.com/gopact-ai/gopact/gopacttest"
)

func TestRunShowsPlanExecuteFlow(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"events: run_started -> node_started(plan) -> node_completed(plan) -> node_started(execute) -> node_completed(execute) -> node_started(summarize) -> node_completed(summarize) -> run_completed",
		"trace: plan -> replan -> execute -> summarize",
		"results: draft-retry=done draft-retry, review=done review",
		"summary: completed 2 steps",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}

func TestPlanExecuteApprovalResume(t *testing.T) {
	executions := 0
	agent, err := planexec.New(
		planexec.PlannerFunc(func(context.Context, planexec.PlanRequest) ([]planexec.Step, error) {
			return []planexec.Step{{ID: "draft", Instruction: "draft example"}}, nil
		}),
		planexec.ExecutorFunc(func(_ context.Context, step planexec.Step) (planexec.StepResult, error) {
			executions++
			return planexec.StepResult{StepID: step.ID, Output: "done " + step.ID}, nil
		}),
		planexec.WithApprovalPolicy(gopact.PolicyFunc(func(context.Context, gopact.PolicyRequest) (gopact.PolicyDecision, error) {
			return gopact.PolicyDecision{Action: gopact.PolicyReview, Reason: "needs approval"}, nil
		})),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	events, err := gopacttest.CollectEvents(agent.Run(context.Background(), "ship example"))
	if !errors.Is(err, gopact.ErrInterrupted) {
		t.Fatalf("Run() error = %v, want ErrInterrupted", err)
	}
	if executions != 0 {
		t.Fatalf("executions before approval = %d, want 0", executions)
	}
	gopacttest.RequireEventTypes(t, events,
		gopact.EventRunStarted,
		gopact.EventNodeStarted,
		gopact.EventNodeCompleted,
		gopact.EventNodeStarted,
		gopact.EventInterrupted,
		gopact.EventRunInterrupted,
	)
	interrupted := events[4].StepSnapshot
	if interrupted == nil || interrupted.Pending == nil {
		t.Fatalf("interrupted step = %+v, want pending approval", interrupted)
	}

	resumed, err := gopacttest.CollectEvents(agent.Run(context.Background(), planexec.State{},
		gopact.WithStepExport(gopact.StepExport{Version: 1, Step: *interrupted}),
		gopact.WithResumeRequest(gopact.ResumeRequest{
			StepID:      interrupted.ID,
			InterruptID: interrupted.Pending.ID,
			Payload:     map[string]any{"approved": true},
		}),
	))
	if err != nil {
		t.Fatalf("resumed Run() error = %v", err)
	}
	gopacttest.RequireEventTypes(t, resumed,
		gopact.EventRunStarted,
		gopact.EventStepImported,
		gopact.EventResumeReceived,
		gopact.EventNodeResumed,
		gopact.EventNodeCompleted,
		gopact.EventNodeStarted,
		gopact.EventNodeCompleted,
		gopact.EventRunCompleted,
	)
	if executions != 1 {
		t.Fatalf("executions after approval = %d, want 1", executions)
	}
}

func TestPlanExecuteCancelStopsBeforeSummary(t *testing.T) {
	agent, err := planexec.New(
		planexec.PlannerFunc(func(context.Context, planexec.PlanRequest) ([]planexec.Step, error) {
			return []planexec.Step{{ID: "draft", Instruction: "draft example"}}, nil
		}),
		planexec.ExecutorFunc(func(context.Context, planexec.Step) (planexec.StepResult, error) {
			return planexec.StepResult{}, context.Canceled
		}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	events, err := gopacttest.CollectEvents(agent.Run(context.Background(), "ship example"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context canceled", err)
	}
	gopacttest.RequireEventTypes(t, events,
		gopact.EventRunStarted,
		gopact.EventNodeStarted,
		gopact.EventNodeCompleted,
		gopact.EventNodeStarted,
		gopact.EventRunCanceled,
	)
	canceled := events[4].StepSnapshot
	if canceled == nil || canceled.Node != "execute" || canceled.Phase != gopact.StepCanceled {
		t.Fatalf("canceled step = %+v, want execute step_canceled", canceled)
	}
	output, ok := canceled.Output.(planexec.State)
	if !ok {
		t.Fatalf("canceled output type = %T, want State", canceled.Output)
	}
	if output.Summary != "" || !reflect.DeepEqual(output.Trace, []string{"plan"}) {
		t.Fatalf("canceled output = %+v, want no summary after plan", output)
	}
}
