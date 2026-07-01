package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-ext/agents/agenttool"
	"github.com/gopact-ai/gopact-ext/agents/planexec"
	"github.com/gopact-ai/gopact-ext/agents/react"
	"github.com/gopact-ai/gopact/a2a"
	"github.com/gopact-ai/gopact/gopacttest"
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

func TestAgentAsToolFailurePreservesChildEvidence(t *testing.T) {
	parentIDs := gopact.RuntimeIDs{
		RunID:    "agent-as-tool-failure-demo",
		ThreadID: "agent-as-tool-thread",
		CallID:   "parent-call",
	}
	childIDs := parentIDs
	childIDs.ParentCallID = parentIDs.CallID
	childIDs.CallID = "call-delegate"
	childErr := errors.New("child executor failed")

	child, err := planexec.New(
		planexec.PlannerFunc(func(context.Context, planexec.PlanRequest) ([]planexec.Step, error) {
			return []planexec.Step{{ID: "draft", Instruction: "draft example"}}, nil
		}),
		planexec.ExecutorFunc(func(context.Context, planexec.Step) (planexec.StepResult, error) {
			return planexec.StepResult{}, childErr
		}),
	)
	if err != nil {
		t.Fatalf("planexec.New() error = %v", err)
	}
	childA2A, err := a2a.NewRunnableAgent(
		a2a.AgentCard{Name: "planexec-child", Description: "Plans a delegated task."},
		child,
		a2a.WithRunnableInputMapper(func(_ context.Context, task a2a.Task) (any, error) {
			return task.Input, nil
		}),
	)
	if err != nil {
		t.Fatalf("a2a.NewRunnableAgent() error = %v", err)
	}
	tool, err := agenttool.New(childA2A, agenttool.WithName("delegate_plan"))
	if err != nil {
		t.Fatalf("agenttool.New() error = %v", err)
	}
	parent, err := react.NewModelAgent(
		&scriptedModel{responses: []gopact.ModelResponse{
			{Message: gopact.Message{
				Role: gopact.RoleAssistant,
				ToolCalls: []gopact.ToolCall{{
					ID:        childIDs.CallID,
					Name:      "local.delegate_plan",
					Arguments: []byte(`{"input":"ship a small example","task_id":"child-task-1"}`),
				}},
			}},
			{Message: gopact.AssistantMessage("should not run")},
		}},
		react.WithTools(context.Background(), tool),
	)
	if err != nil {
		t.Fatalf("react.NewModelAgent() error = %v", err)
	}

	events, err := gopacttest.CollectEvents(parent.Run(context.Background(), "delegate planning", gopact.WithRuntimeIDs(parentIDs)))
	if !errors.Is(err, childErr) {
		t.Fatalf("Run() error = %v, want child executor failure", err)
	}
	gopacttest.RequireEventTypes(t, events,
		gopact.EventRunStarted,
		gopact.EventNodeStarted,
		gopact.EventModelMessage,
		gopact.EventNodeCompleted,
		gopact.EventNodeStarted,
		gopact.EventToolCall,
		gopact.EventA2ATaskFailed,
		gopact.EventNodeFailed,
		gopact.EventRunFailed,
	)
	if events[6].Metadata["agent_name"] != "planexec-child" ||
		events[6].Metadata["a2a_task_id"] != "child-task-1" ||
		events[6].Metadata["a2a_status"] != string(a2a.TaskStatusFailed) {
		t.Fatalf("child failure metadata = %+v, want agent-as-tool failure evidence", events[6].Metadata)
	}
}
