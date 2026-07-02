package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"strings"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-ext/agents/agentnode"
	"github.com/gopact-ai/gopact/a2a"
	"github.com/gopact-ai/gopact/graph"
)

type workflowState struct {
	Input string
	Plan  string
}

type plannerAgent struct{}

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	workflow, err := newWorkflow()
	if err != nil {
		return err
	}

	state := workflowState{Input: "ship a self-bootstrap quickstart"}
	events := []string{}
	childEvidence := ""
	for event, err := range workflow.Run(ctx, state, graph.WithRuntimeIDs(gopact.RuntimeIDs{
		RunID:    "agent-node-demo",
		ThreadID: "quickstart-agent-node",
	})) {
		if err != nil {
			return err
		}
		events = append(events, eventLabel(event))
		if event.Type == gopact.EventA2ATaskCompleted {
			childEvidence = fmt.Sprintf(
				"%s(%v, %v)",
				event.Type,
				event.Metadata["agent_name"],
				event.Metadata["a2a_task_id"],
			)
		}
		if event.Type == gopact.EventNodeCompleted && event.StepSnapshot != nil {
			if next, ok := event.StepSnapshot.Output.(workflowState); ok {
				state = next
			}
		}
	}

	if _, err := fmt.Fprintf(out, "plan: %s\n", state.Plan); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "child_evidence: %s\n", childEvidence); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "events: %s\n", strings.Join(events, " -> "))
	return err
}

func newWorkflow() (*graph.Runnable[workflowState], error) {
	node, err := agentnode.New[workflowState](
		plannerAgent{},
		func(ctx context.Context, state workflowState) (a2a.Task, error) {
			ids, _ := gopact.RuntimeIDsFromContext(ctx)
			return a2a.Task{ID: "planner-task", IDs: ids, Input: state.Input}, nil
		},
		func(_ context.Context, state workflowState, result a2a.Result) (workflowState, error) {
			state.Plan = result.Output
			return state, nil
		},
	)
	if err != nil {
		return nil, err
	}

	g := graph.New[workflowState]()
	g.AddNode("delegate", node)
	g.AddEdge(graph.Start, "delegate")
	g.AddEdge("delegate", graph.End)
	return g.Compile()
}

func (plannerAgent) Card() a2a.AgentCard {
	return a2a.AgentCard{
		Name:        "planner-agent",
		Description: "Builds a small execution plan.",
		URL:         "memory://planner-agent",
		Streaming:   true,
		Skills: []a2a.AgentSkill{{
			Name:        "plan",
			Description: "Create a short plan from a task.",
		}},
	}
}

func (plannerAgent) Send(_ context.Context, task a2a.Task) (a2a.Result, error) {
	return a2a.Result{TaskID: task.ID, Output: "plan: research -> code -> review"}, nil
}

func (plannerAgent) Cancel(context.Context, string) error {
	return nil
}

func (plannerAgent) Stream(ctx context.Context, task a2a.Task) iter.Seq2[a2a.TaskEvent, error] {
	return func(yield func(a2a.TaskEvent, error) bool) {
		if ctx == nil {
			ctx = context.TODO()
		}
		if err := ctx.Err(); err != nil {
			yield(a2a.TaskEvent{TaskID: task.ID, IDs: task.IDs, Status: a2a.TaskStatusFailed, Err: err}, err)
			return
		}
		if task.ID == "" {
			err := errors.New("planner task id is required")
			yield(a2a.TaskEvent{IDs: task.IDs, Status: a2a.TaskStatusFailed, Err: err}, err)
			return
		}
		if !yield(a2a.TaskEvent{
			TaskID:  task.ID,
			IDs:     task.IDs,
			Status:  a2a.TaskStatusRunning,
			Message: "planning",
		}, nil) {
			return
		}
		yield(a2a.TaskEvent{
			TaskID: task.ID,
			IDs:    task.IDs,
			Status: a2a.TaskStatusCompleted,
			Result: &a2a.Result{TaskID: task.ID, Output: "plan: research -> code -> review"},
		}, nil)
	}
}

func eventLabel(event gopact.Event) string {
	if event.Node == "" {
		return string(event.Type)
	}
	return fmt.Sprintf("%s(%s)", event.Type, event.Node)
}
