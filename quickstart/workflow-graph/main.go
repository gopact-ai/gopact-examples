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
	Task        string
	Plan        []string
	Done        []string
	Refinements int
	Trace       []string
	Summary     string
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
	nestedEvents := []string{}
	for event, err := range run.Run(ctx, state, graph.WithRuntimeIDs(gopact.RuntimeIDs{RunID: "workflow-demo"})) {
		if err != nil {
			return err
		}
		events = append(events, eventLabel(event))
		if event.Metadata[graph.EventMetadataParentNode] == "polish" && event.Node != "" {
			nestedEvents = append(nestedEvents, eventLabel(event))
		}
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
	if _, err := fmt.Fprintf(out, "nested events: %s\n", strings.Join(nestedEvents, " -> ")); err != nil {
		return err
	}
	stepLimit, err := stepLimitGuard(ctx)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "step limit: %s\n", stepLimit); err != nil {
		return err
	}
	completedResume, interruptedResume, err := stepResumeDemos(ctx)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "step export resume: %s\n", completedResume); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "interrupt resume: %s\n", interruptedResume); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "summary: %s\n", state.Summary)
	return err
}

func newWorkflow() (*graph.Runnable[workflowState], error) {
	polish, err := newPolishWorkflow()
	if err != nil {
		return nil, err
	}
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
	g.AddRunnableNode("polish", polish)
	g.AddNode("refine", func(_ context.Context, state workflowState) (workflowState, error) {
		state.Refinements++
		state.Trace = append(state.Trace, fmt.Sprintf("refine-%d", state.Refinements))
		return state, nil
	})
	g.AddNode("summarize", func(_ context.Context, state workflowState) (workflowState, error) {
		state.Summary = fmt.Sprintf("workflow completed %d parallel actions after %d refinements", len(state.Done), state.Refinements)
		state.Trace = append(state.Trace, "summarize")
		return state, nil
	})
	g.AddEdge(graph.Start, "plan")
	g.AddBranch("plan", func(_ context.Context, state workflowState) ([]string, error) {
		return append([]string(nil), state.Plan...), nil
	})
	g.AddEdge("draft", "polish")
	g.AddEdge("review", "polish")
	g.AddEdge("polish", "refine")
	g.AddBranch("refine", func(_ context.Context, state workflowState) ([]string, error) {
		if state.Refinements < 2 {
			return []string{"refine"}, nil
		}
		return []string{"summarize"}, nil
	})
	g.AddEdge("summarize", graph.End)
	return g.Compile()
}

func newPolishWorkflow() (*graph.Runnable[workflowState], error) {
	g := graph.New[workflowState]()
	g.AddNode("polish-start", func(_ context.Context, state workflowState) (workflowState, error) {
		state.Trace = append(state.Trace, "polish-start")
		return state, nil
	})
	g.AddNode("polish-finish", func(_ context.Context, state workflowState) (workflowState, error) {
		state.Trace = append(state.Trace, "polish-finish")
		return state, nil
	})
	g.AddEdge(graph.Start, "polish-start")
	g.AddEdge("polish-start", "polish-finish")
	g.AddEdge("polish-finish", graph.End)
	return g.Compile()
}

func stepLimitGuard(ctx context.Context) (string, error) {
	g := graph.New[workflowState]()
	g.AddNode("loop", func(_ context.Context, state workflowState) (workflowState, error) {
		state.Refinements++
		return state, nil
	})
	g.AddEdge(graph.Start, "loop")
	g.AddEdge("loop", "loop")
	run, err := g.Compile()
	if err != nil {
		return "", err
	}
	_, err = run.Invoke(ctx, workflowState{}, graph.WithMaxSteps(2))
	if err == nil {
		return "", fmt.Errorf("step limit guard did not fail")
	}
	return err.Error(), nil
}

func stepResumeDemos(ctx context.Context) (string, string, error) {
	completed, err := completedStepExportResume(ctx)
	if err != nil {
		return "", "", err
	}
	interrupted, err := interruptedStepExportResume(ctx)
	if err != nil {
		return "", "", err
	}
	return completed, interrupted, nil
}

func completedStepExportResume(ctx context.Context) (string, error) {
	ids := gopact.RuntimeIDs{RunID: "workflow-step-export"}
	g := graph.New[workflowState]()
	g.AddNode("first", func(_ context.Context, state workflowState) (workflowState, error) {
		return state, fmt.Errorf("completed exported step reran")
	})
	g.AddNode("next", func(_ context.Context, state workflowState) (workflowState, error) {
		state.Trace = append(state.Trace, "next")
		return state, nil
	})
	g.AddEdge(graph.Start, "first")
	g.AddEdge("first", "next")
	g.AddEdge("next", graph.End)
	run, err := g.Compile()
	if err != nil {
		return "", err
	}
	return eventTypes(run.Run(ctx, workflowState{}, graph.WithStepExport(gopact.StepExport{
		Version: gopact.RunExportVersion,
		Step: gopact.StepSnapshot{
			ID:     "workflow-step-export:1",
			Step:   1,
			Node:   "first",
			Phase:  gopact.StepCompleted,
			IDs:    ids,
			Output: workflowState{Trace: []string{"first"}},
			Queue:  []string{"next"},
		},
	})))
}

func interruptedStepExportResume(ctx context.Context) (string, error) {
	ids := gopact.RuntimeIDs{RunID: "workflow-step-interrupt"}
	g := graph.New[workflowState]()
	g.AddNode("ask", func(_ context.Context, state workflowState) (workflowState, error) {
		return state, fmt.Errorf("interrupted exported step reran")
	})
	g.AddNode("answer", func(_ context.Context, state workflowState) (workflowState, error) {
		state.Trace = append(state.Trace, "answer")
		return state, nil
	})
	g.AddEdge(graph.Start, "ask")
	g.AddEdge("ask", "answer")
	g.AddEdge("answer", graph.End)
	run, err := g.Compile()
	if err != nil {
		return "", err
	}
	return eventTypes(run.Run(ctx, workflowState{},
		graph.WithStepExport(gopact.StepExport{
			Version: gopact.RunExportVersion,
			Step: gopact.StepSnapshot{
				ID:     "workflow-step-interrupt:1",
				Step:   1,
				Node:   "ask",
				Phase:  gopact.StepInterrupted,
				IDs:    ids,
				Output: workflowState{Trace: []string{"ask"}},
				Queue:  []string{"answer"},
				Pending: &gopact.InterruptRecord{
					ID:     "interrupt-ask",
					Type:   gopact.InterruptInput,
					Reason: "need input",
				},
			},
		}),
		graph.WithResumeRequest(gopact.ResumeRequest{
			StepID:      "workflow-step-interrupt:1",
			InterruptID: "interrupt-ask",
			Payload:     "continue",
		}),
	))
}

func eventTypes(seq func(func(gopact.Event, error) bool)) (string, error) {
	events := []string{}
	for event, err := range seq {
		if err != nil {
			return "", err
		}
		if event.Type == gopact.EventRunStarted {
			continue
		}
		events = append(events, string(event.Type))
	}
	return strings.Join(events, " -> "), nil
}

func eventLabel(event gopact.Event) string {
	if event.Node == "" {
		return string(event.Type)
	}
	return fmt.Sprintf("%s(%s)", event.Type, event.Node)
}
