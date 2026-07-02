package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-ext/agents/planexec"
)

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	failFirstDraft := true
	agent, err := planexec.New(
		planexec.PlannerFunc(func(_ context.Context, request planexec.PlanRequest) ([]planexec.Step, error) {
			return []planexec.Step{
				{ID: "draft", Instruction: "Draft " + request.Task},
				{ID: "review", Instruction: "Review " + request.Task},
			}, nil
		}),
		planexec.ExecutorFunc(func(_ context.Context, step planexec.Step) (planexec.StepResult, error) {
			if step.ID == "draft" && failFirstDraft {
				failFirstDraft = false
				return planexec.StepResult{}, errors.New("draft failed")
			}
			return planexec.StepResult{StepID: step.ID, Output: "done " + step.ID}, nil
		}),
		planexec.WithReplanner(planexec.ReplannerFunc(func(_ context.Context, request planexec.ReplanRequest) ([]planexec.Step, error) {
			return []planexec.Step{
				{ID: "draft-retry", Instruction: "Retry " + request.Task},
				{ID: "review", Instruction: "Review " + request.Task},
			}, nil
		})),
	)
	if err != nil {
		return err
	}

	state := planexec.State{Task: "a tiny example"}
	var events []string
	for event, err := range agent.Run(ctx, state, gopact.WithRuntimeIDs(gopact.RuntimeIDs{RunID: "plan-exec-demo"})) {
		if err != nil {
			return err
		}
		events = append(events, eventLabel(event))
		if event.Type == gopact.EventNodeCompleted {
			if next, ok := event.StepSnapshot.Output.(planexec.State); ok {
				state = next
			}
		}
	}

	if _, err := fmt.Fprintf(out, "events: %s\n", strings.Join(events, " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "trace: %s\n", strings.Join(state.Trace, " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "results: %s\n", resultsText(state.Results)); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "summary: %s\n", state.Summary)
	return err
}

func resultsText(results []planexec.StepResult) string {
	parts := make([]string, 0, len(results))
	for _, result := range results {
		parts = append(parts, result.StepID+"="+result.Output)
	}
	return strings.Join(parts, ", ")
}

func eventLabel(event gopact.Event) string {
	if event.Node == "" {
		return string(event.Type)
	}
	return fmt.Sprintf("%s(%s)", event.Type, event.Node)
}
