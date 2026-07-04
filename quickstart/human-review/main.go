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
	"github.com/gopact-ai/gopact-ext/agents/humanreview"
	"github.com/gopact-ai/gopact/checkpoint"
	"github.com/gopact-ai/gopact/graph"
)

type reviewState struct {
	Task    string
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
	workflow, err := newReviewWorkflow()
	if err != nil {
		return err
	}

	firstEvents, stepEvents, stepFinal, pending, err := stepExportResume(ctx, workflow)
	if err != nil {
		return err
	}
	checkpointEvents, checkpointFinal, checkpointID, err := checkpointResume(ctx, workflow)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(out, "first_events: %s\n", strings.Join(firstEvents, " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "pending: %s required_by=%s template=%v\n", pending.ID, pending.RequiredBy, pending.Metadata["template"]); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "step_export_resume: %s\n", strings.Join(stepEvents, " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "checkpoint_resume: %s checkpoint=%s\n", strings.Join(checkpointEvents, " -> "), checkpointID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "step_trace: %s\n", strings.Join(stepFinal.Trace, " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "checkpoint_trace: %s\n", strings.Join(checkpointFinal.Trace, " -> ")); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "summary: %s\n", stepFinal.Summary)
	return err
}

func newReviewWorkflow() (*graph.Runnable[reviewState], error) {
	review, err := humanreview.New(func(_ context.Context, state reviewState) (humanreview.Request, error) {
		return humanreview.Request{
			ID:         "review:publish",
			Reason:     "publish approval required",
			RequiredBy: "release",
			Prompt:     gopact.UserMessage("Approve publishing " + state.Draft + "?"),
			Metadata:   map[string]any{"risk": "medium"},
		}, nil
	})
	if err != nil {
		return nil, err
	}

	g := graph.New[reviewState]()
	g.AddNode("draft", func(_ context.Context, state reviewState) (reviewState, error) {
		state.Draft = "draft for " + state.Task
		state.Trace = append(state.Trace, "draft")
		return state, nil
	})
	g.AddNode("review", review)
	g.AddNode("publish", func(_ context.Context, state reviewState) (reviewState, error) {
		state.Summary = "published " + state.Draft
		state.Trace = append(state.Trace, "publish")
		return state, nil
	})
	g.AddEdge(graph.Start, "draft")
	g.AddEdge("draft", "review")
	g.AddEdge("review", "publish")
	g.AddEdge("publish", graph.End)
	return g.Compile()
}

func stepExportResume(ctx context.Context, workflow *graph.Runnable[reviewState]) ([]string, []string, reviewState, gopact.InterruptRecord, error) {
	ids := gopact.RuntimeIDs{RunID: "human-review-step", ThreadID: "human-review-step-thread"}
	firstEvents, _, interrupted, err := collectReviewRun(workflow.Run(ctx,
		reviewState{Task: "release notes"},
		graph.WithRuntimeIDs(ids),
	))
	if !errors.Is(err, gopact.ErrInterrupted) {
		return nil, nil, reviewState{}, gopact.InterruptRecord{}, fmt.Errorf("step export first run error = %v, want %v", err, gopact.ErrInterrupted)
	}
	if interrupted == nil || interrupted.Pending == nil {
		return nil, nil, reviewState{}, gopact.InterruptRecord{}, errors.New("step export first run missing pending review")
	}

	resumed, final, _, err := collectReviewRun(workflow.Run(ctx,
		reviewState{},
		graph.WithRuntimeIDs(ids),
		graph.WithStepExport(gopact.StepExport{Version: gopact.RunExportVersion, Step: *interrupted}),
		graph.WithResumeRequest(gopact.ResumeRequest{
			StepID:      interrupted.ID,
			InterruptID: interrupted.Pending.ID,
			Payload:     map[string]any{"approved": true},
		}),
	))
	if err != nil {
		return nil, nil, reviewState{}, gopact.InterruptRecord{}, err
	}
	return firstEvents, resumed, final, *interrupted.Pending, nil
}

func checkpointResume(ctx context.Context, workflow *graph.Runnable[reviewState]) ([]string, reviewState, string, error) {
	store := checkpoint.NewMemory[reviewState]()
	ids := gopact.RuntimeIDs{RunID: "human-review-checkpoint-first", ThreadID: "human-review-checkpoint-thread"}
	_, _, _, err := collectReviewRun(workflow.Run(ctx,
		reviewState{Task: "release notes"},
		graph.WithRuntimeIDs(ids),
		graph.WithCheckpointStore(store),
	))
	if !errors.Is(err, gopact.ErrInterrupted) {
		return nil, reviewState{}, "", fmt.Errorf("checkpoint first run error = %v, want %v", err, gopact.ErrInterrupted)
	}
	latest, ok, err := store.Latest(ctx, ids.ThreadID)
	if err != nil {
		return nil, reviewState{}, "", err
	}
	if !ok || latest.Pending == nil {
		return nil, reviewState{}, "", errors.New("checkpoint first run missing pending review")
	}

	resumed, final, _, err := collectReviewRun(workflow.Run(ctx,
		reviewState{},
		graph.WithRuntimeIDs(gopact.RuntimeIDs{RunID: "human-review-checkpoint-resume", ThreadID: ids.ThreadID}),
		graph.WithCheckpointStore(store),
		graph.WithResumeRequest(gopact.ResumeRequest{
			CheckpointID: latest.ID,
			InterruptID:  latest.Pending.ID,
			Payload:      map[string]any{"approved": true},
		}),
	))
	if err != nil {
		return nil, reviewState{}, "", err
	}
	return resumed, final, latest.ID, nil
}

func collectReviewRun(events iter.Seq2[gopact.Event, error]) ([]string, reviewState, *gopact.StepSnapshot, error) {
	var labels []string
	var state reviewState
	var interrupted *gopact.StepSnapshot
	for event, err := range events {
		labels = append(labels, eventLabel(event))
		if event.StepSnapshot != nil {
			if next, ok := event.StepSnapshot.Output.(reviewState); ok {
				state = next
			}
			if event.StepSnapshot.Pending != nil {
				step := *event.StepSnapshot
				interrupted = &step
			}
		}
		if err != nil {
			return labels, state, interrupted, err
		}
	}
	return labels, state, interrupted, nil
}

func eventLabel(event gopact.Event) string {
	if event.Node == "" {
		return string(event.Type)
	}
	return fmt.Sprintf("%s(%s)", event.Type, event.Node)
}
