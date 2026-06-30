package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-ext/agents/agenttool"
	"github.com/gopact-ai/gopact-ext/agents/planexec"
	"github.com/gopact-ai/gopact-ext/agents/react"
	"github.com/gopact-ai/gopact/a2a"
)

type scriptedModel struct {
	responses []gopact.ModelResponse
}

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	child, err := planexec.New(
		planexec.PlannerFunc(func(_ context.Context, request planexec.PlanRequest) ([]planexec.Step, error) {
			return []planexec.Step{{ID: "draft", Instruction: "Draft " + request.Task}}, nil
		}),
		planexec.ExecutorFunc(func(_ context.Context, step planexec.Step) (planexec.StepResult, error) {
			return planexec.StepResult{StepID: step.ID, Output: "done " + step.ID}, nil
		}),
	)
	if err != nil {
		return err
	}

	childA2A, err := a2a.NewRunnableAgent(
		a2a.AgentCard{Name: "planexec-child", Description: "Plans a delegated task."},
		child,
		a2a.WithRunnableInputMapper(func(_ context.Context, task a2a.Task) (any, error) {
			return task.Input, nil
		}),
		a2a.WithRunnableResultMapper(func(_ context.Context, task a2a.Task, events []gopact.Event) (a2a.Result, error) {
			for i := len(events) - 1; i >= 0; i-- {
				if events[i].StepSnapshot == nil {
					continue
				}
				state, ok := events[i].StepSnapshot.Output.(planexec.State)
				if ok && state.Summary != "" {
					return a2a.Result{TaskID: task.ID, Output: state.Summary}, nil
				}
			}
			return a2a.Result{TaskID: task.ID}, nil
		}),
	)
	if err != nil {
		return err
	}
	tool, err := agenttool.New(childA2A, agenttool.WithName("delegate_plan"))
	if err != nil {
		return err
	}

	parent, err := react.NewModelAgent(
		&scriptedModel{responses: []gopact.ModelResponse{
			{Message: gopact.Message{
				Role: gopact.RoleAssistant,
				ToolCalls: []gopact.ToolCall{{
					ID:        "call-delegate",
					Name:      "local.delegate_plan",
					Arguments: []byte(`{"input":"ship a small example","task_id":"child-task-1"}`),
				}},
			}},
			{Message: gopact.AssistantMessage("delegated")},
		}},
		react.WithTools(ctx, tool),
	)
	if err != nil {
		return err
	}

	ids := gopact.RuntimeIDs{RunID: "agent-as-tool-demo", CallID: "parent-call"}
	events := []string{}
	childResult := ""
	childEvidence := ""
	parentText := ""
	for event, err := range parent.Run(ctx, "delegate planning", gopact.WithRuntimeIDs(ids)) {
		if err != nil {
			return err
		}
		events = append(events, eventLabel(event))
		if event.Type == gopact.EventA2ATaskCompleted && event.Result != nil {
			childResult = event.Result.Content
			childEvidence = fmt.Sprintf("%s(%s, %v)", event.Type, event.Metadata["agent_name"], event.Metadata["a2a_task_id"])
		}
		if event.Type == gopact.EventModelMessage && event.Message != nil && event.Message.Text() != "" {
			parentText = event.Message.Text()
		}
	}

	if _, err := fmt.Fprintf(out, "parent: %s\n", parentText); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "child_result: %s\n", childResult); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "child_evidence: %s\n", childEvidence); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "events: %s\n", strings.Join(events, " -> "))
	return err
}

func (m *scriptedModel) Generate(ctx context.Context, request gopact.ModelRequest) (gopact.ModelResponse, error) {
	if err := ctx.Err(); err != nil {
		return gopact.ModelResponse{}, err
	}
	if len(m.responses) == 0 {
		return gopact.ModelResponse{}, fmt.Errorf("missing scripted response")
	}
	response := m.responses[0]
	m.responses = m.responses[1:]
	_ = request
	return response, nil
}

func eventLabel(event gopact.Event) string {
	if event.Node == "" {
		return string(event.Type)
	}
	return fmt.Sprintf("%s(%s)", event.Type, event.Node)
}
