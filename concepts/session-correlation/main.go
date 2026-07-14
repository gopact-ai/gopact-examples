package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact/agent"
	"github.com/gopact-ai/gopact/workflow"
)

const (
	sessionID   = "customer-case-42"
	intakeRunID = "run-intake"
	reviewRunID = "run-review"
)

type exampleResult struct {
	sessionID  string
	beforeRuns []workflow.RunSummary
	afterRuns  []workflow.RunSummary
	snapshot   workflow.Snapshot
	output     string
}

func runExample(ctx context.Context) (exampleResult, error) {
	store := workflow.NewMemoryStore()
	storeOptions := []workflow.BuildOption{
		workflow.WithStore(store),
	}

	intake := workflow.New[agent.Request, agent.Response]("intake", storeOptions...)
	respond := intake.Node("respond", func(_ context.Context, request agent.Request) (agent.Response, error) {
		return agent.Response{Message: request.Messages[0]}, nil
	})
	intake.Entry(respond)
	intake.Exit(respond)
	intakeAgent, err := agent.NewWorkflowAgent(agent.Identity{
		Name:        "intake",
		Description: "records a customer request",
		Version:     "v1",
	}, intake)
	if err != nil {
		return exampleResult{}, fmt.Errorf("create intake agent: %w", err)
	}
	if _, err := intakeAgent.Invoke(
		ctx,
		agent.Request{Messages: []gopact.Message{gopact.UserMessage("request")}},
		gopact.WithSessionID(sessionID),
		gopact.WithRunID(intakeRunID),
	); err != nil {
		return exampleResult{}, fmt.Errorf("invoke intake: %w", err)
	}

	review := workflow.New[string, string]("review", storeOptions...)
	approve := review.Node("approve", func(_ context.Context, input string) (string, error) {
		return "approved:" + input, nil
	})
	approve.Guard(workflow.BeforeRun("approval", workflow.GuardFunc[string, string](
		func(context.Context, workflow.GuardContext[string, string]) (workflow.GuardDecision[string, string], error) {
			return workflow.GuardInterrupt[string, string]{Request: workflow.InterruptRequest{
				ID:                  "approval-1",
				Subject:             "customer review",
				ResolutionSchemaRef: "schema://approval",
			}}, nil
		},
	)))
	review.Entry(approve)
	review.Exit(approve)

	_, err = review.Invoke(
		ctx,
		"request",
		gopact.WithSessionID(sessionID),
		gopact.WithRunID(reviewRunID),
	)
	if err == nil {
		return exampleResult{}, errors.New("invoke review: expected approval interrupt")
	}
	var interrupt workflow.InterruptError
	if !errors.As(err, &interrupt) {
		return exampleResult{}, fmt.Errorf("invoke review: %w", err)
	}

	beforeRuns, err := workflow.ListSessionRuns(ctx, store, workflow.SessionRunsRequest{SessionID: sessionID})
	if err != nil {
		return exampleResult{}, fmt.Errorf("list session runs before resume: %w", err)
	}
	snapshot, err := review.Snapshot(ctx, workflow.SnapshotRequest{RunID: reviewRunID})
	if err != nil {
		return exampleResult{}, fmt.Errorf("snapshot review: %w", err)
	}
	output, err := review.Invoke(ctx, "", workflow.WithResume(workflow.ResumeRequest{
		RunID:        reviewRunID,
		CheckpointID: interrupt.CheckpointID,
		Resolutions: []workflow.InterruptResolution{{
			InterruptID: "approval-1",
			PayloadRef:  "artifact://approved",
		}},
	}))
	if err != nil {
		return exampleResult{}, fmt.Errorf("resume review: %w", err)
	}
	afterRuns, err := workflow.ListSessionRuns(ctx, store, workflow.SessionRunsRequest{SessionID: sessionID})
	if err != nil {
		return exampleResult{}, fmt.Errorf("list session runs after resume: %w", err)
	}

	return exampleResult{
		sessionID:  sessionID,
		beforeRuns: beforeRuns,
		afterRuns:  afterRuns,
		snapshot:   snapshot,
		output:     output,
	}, nil
}

func main() {
	result, err := runExample(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf(
		"session=%s runs=%d selected=%s status=%s output=%s\n",
		result.sessionID,
		len(result.afterRuns),
		result.snapshot.RunMeta.RunID,
		result.afterRuns[1].Status,
		result.output,
	)
}
