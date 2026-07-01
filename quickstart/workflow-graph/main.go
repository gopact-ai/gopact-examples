package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact/graph"
)

type workflowState struct {
	Task    string
	Plan    []string
	Done    []string
	Trace   []string
	Summary string
}

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	run, err := newWorkflow()
	if err != nil {
		return err
	}

	state := workflowState{Task: "ship a tiny workflow example"}
	events := []string{}
	for event, err := range run.Run(ctx, state, graph.WithRuntimeIDs(gopact.RuntimeIDs{RunID: "workflow-demo"})) {
		if err != nil {
			return err
		}
		events = append(events, eventLabel(event))
		if event.Type == gopact.EventNodeCompleted {
			if next, ok := event.StepSnapshot.Output.(workflowState); ok {
				state = next
			}
		}
	}

	if _, err := fmt.Fprintf(out, "events: %s\n", strings.Join(events, " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "steps: %s\n", strings.Join(state.Trace, " -> ")); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "summary: %s\n", state.Summary)
	return err
}

func newWorkflow() (*graph.Runnable[workflowState], error) {
	g := graph.New[workflowState]()
	g.AddNode("plan", func(_ context.Context, state workflowState) (workflowState, error) {
		state.Plan = []string{"draft", "review"}
		state.Trace = append(state.Trace, "plan")
		return state, nil
	})
	g.AddNode("draft", func(_ context.Context, state workflowState) (workflowState, error) {
		state.Done = append(state.Done, "draft")
		state.Trace = append(state.Trace, "draft")
		return state, nil
	})
	g.AddNode("review", func(_ context.Context, state workflowState) (workflowState, error) {
		state.Done = append(state.Done, "review")
		state.Trace = append(state.Trace, "review")
		return state, nil
	})
	g.AddNode("summarize", func(_ context.Context, state workflowState) (workflowState, error) {
		state.Summary = fmt.Sprintf("workflow completed %d parallel actions", len(state.Done))
		state.Trace = append(state.Trace, "summarize")
		return state, nil
	})
	g.AddEdge(graph.Start, "plan")
	g.AddBranch("plan", func(_ context.Context, state workflowState) ([]string, error) {
		return append([]string(nil), state.Plan...), nil
	})
	g.AddEdge("draft", "summarize")
	g.AddEdge("review", "summarize")
	g.AddEdge("summarize", graph.End)
	return g.Compile()
}

func eventLabel(event gopact.Event) string {
	if event.Node == "" {
		return string(event.Type)
	}
	return fmt.Sprintf("%s(%s)", event.Type, event.Node)
}
