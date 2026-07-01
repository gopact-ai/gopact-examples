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
	"github.com/gopact-ai/gopact/checkpoint"
	"github.com/gopact-ai/gopact/graph"
)

type agentState struct {
	Task    string
	Plan    []string
	Draft   string
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
	workflow, err := newAgentWorkflow()
	if err != nil {
		return err
	}
	store := checkpoint.NewMemory[agentState]()
	threadID := "scaffold-thread"

	firstEvents, _, interrupted, err := collectRun(workflow.Run(ctx,
		agentState{Task: "add a README example"},
		graph.WithRuntimeIDs(gopact.RuntimeIDs{RunID: "scaffold-first", ThreadID: threadID}),
		graph.WithCheckpointStore(store),
	))
	if err != nil {
		return err
	}
	if !interrupted {
		return errors.New("agent scaffold: first run should wait for approval")
	}

	checkpoints := store.List(ctx, threadID)
	if len(checkpoints) == 0 || checkpoints[len(checkpoints)-1].Pending == nil {
		return errors.New("agent scaffold: missing approval checkpoint")
	}
	pending := checkpoints[len(checkpoints)-1]

	resumeEvents, final, interrupted, err := collectRun(workflow.Run(ctx,
		agentState{},
		graph.WithRuntimeIDs(gopact.RuntimeIDs{RunID: "scaffold-resume", ThreadID: threadID}),
		graph.WithCheckpointStore(store),
		graph.WithResumeRequest(gopact.ResumeRequest{
			CheckpointID: pending.ID,
			InterruptID:  pending.Pending.ID,
			Payload:      map[string]any{"approved": true},
		}),
	))
	if err != nil {
		return err
	}
	if interrupted {
		return errors.New("agent scaffold: resume should complete")
	}

	if _, err := fmt.Fprintf(out, "first_events: %s\n", strings.Join(firstEvents, " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "pending: %s checkpoint=%s\n", pending.Pending.Type, pending.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "resume_events: %s\n", strings.Join(resumeEvents, " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "trace: %s\n", strings.Join(final.Trace, " -> ")); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "summary: %s\n", final.Summary)
	return err
}

func newAgentWorkflow() (*graph.Runnable[agentState], error) {
	g := graph.New[agentState]()
	g.AddNode("plan", func(_ context.Context, state agentState) (agentState, error) {
		state.Plan = []string{"draft", "approval"}
		state.Trace = append(state.Trace, "plan")
		return state, nil
	})
	g.AddNode("write", func(_ context.Context, state agentState) (agentState, error) {
		state.Draft = "draft for " + state.Task
		state.Trace = append(state.Trace, "write")
		return state, nil
	})
	g.AddNode("approval", func(_ context.Context, state agentState) (agentState, error) {
		state.Trace = append(state.Trace, "approval")
		return state, gopact.Interrupt(gopact.InterruptRecord{
			ID:     "approval-1",
			Type:   gopact.InterruptApproval,
			Reason: "review draft before publishing",
		})
	})
	g.AddNode("summary", func(_ context.Context, state agentState) (agentState, error) {
		state.Summary = "published " + state.Draft
		state.Trace = append(state.Trace, "summary")
		return state, nil
	})
	g.AddEdge(graph.Start, "plan")
	g.AddEdge("plan", "write")
	g.AddEdge("write", "approval")
	g.AddEdge("approval", "summary")
	g.AddEdge("summary", graph.End)
	return g.Compile()
}

func collectRun(events iter.Seq2[gopact.Event, error]) ([]string, agentState, bool, error) {
	var labels []string
	var state agentState
	for event, err := range events {
		labels = append(labels, eventLabel(event))
		if event.StepSnapshot != nil {
			if next, ok := event.StepSnapshot.Output.(agentState); ok {
				state = next
			}
		}
		if err != nil {
			if errors.Is(err, gopact.ErrInterrupted) {
				return labels, state, true, nil
			}
			return labels, state, false, err
		}
	}
	return labels, state, false, nil
}

func eventLabel(event gopact.Event) string {
	if event.Node == "" {
		return string(event.Type)
	}
	return fmt.Sprintf("%s(%s)", event.Type, event.Node)
}
