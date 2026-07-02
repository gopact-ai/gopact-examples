package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-ext/agents/planexec"
	"github.com/gopact-ai/gopact-ext/agents/supervisor"
)

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	writer, err := planExecChild("writer", []planexec.Step{
		{ID: "outline", Instruction: "Outline the release note"},
		{ID: "polish", Instruction: "Polish the release note"},
	})
	if err != nil {
		return err
	}
	reviewer, err := planExecChild("reviewer", []planexec.Step{
		{ID: "check", Instruction: "Check the release note"},
	})
	if err != nil {
		return err
	}

	agent, err := supervisor.New(
		supervisor.RouterFunc(func(_ context.Context, request supervisor.Request) (supervisor.Route, error) {
			agentName := "reviewer"
			if strings.Contains(strings.ToLower(request.Task), "draft") {
				agentName = "writer"
			}
			return supervisor.Route{Agent: agentName, Input: request.Task}, nil
		}),
		supervisor.Child{Name: "writer", Runnable: writer},
		supervisor.Child{Name: "reviewer", Runnable: reviewer},
	)
	if err != nil {
		return err
	}

	var (
		selected     string
		childAgent   string
		childSummary string
		childTrace   []string
		childResults []planexec.StepResult
		events       []string
	)
	for event, err := range agent.Run(ctx,
		supervisor.State{Task: "draft a tiny release note"},
		gopact.WithRuntimeIDs(gopact.RuntimeIDs{RunID: "supervisor-demo", ThreadID: "quickstart-supervisor"}),
	) {
		if err != nil {
			return err
		}
		events = append(events, eventLabel(event))
		if event.StepSnapshot == nil {
			continue
		}
		if state, ok := event.StepSnapshot.Output.(supervisor.State); ok && state.SelectedAgent != "" {
			selected = state.SelectedAgent
		}
		if state, ok := event.StepSnapshot.Output.(planexec.State); ok && state.Summary != "" {
			childAgent = event.IDs.AgentID
			childSummary = state.Summary
			childTrace = append([]string(nil), state.Trace...)
			childResults = append([]planexec.StepResult(nil), state.Results...)
		}
	}

	if _, err := fmt.Fprintf(out, "selected: %s\n", selected); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "child_agent: %s\n", childAgent); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "child_trace: %s\n", strings.Join(childTrace, " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "child_results: %s\n", resultsText(childResults)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "child_summary: %s\n", childSummary); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "events: %s\n", strings.Join(events, " -> "))
	return err
}

func planExecChild(name string, steps []planexec.Step) (*planexec.Agent, error) {
	return planexec.New(
		planexec.PlannerFunc(func(context.Context, planexec.PlanRequest) ([]planexec.Step, error) {
			return append([]planexec.Step(nil), steps...), nil
		}),
		planexec.ExecutorFunc(func(_ context.Context, step planexec.Step) (planexec.StepResult, error) {
			return planexec.StepResult{StepID: step.ID, Output: name + " done " + step.ID}, nil
		}),
	)
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
