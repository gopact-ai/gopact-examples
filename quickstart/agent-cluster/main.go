package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-ext/devagent/gitdiff"
	"github.com/gopact-ai/gopact/a2a"
	"github.com/gopact-ai/gopact/gopacttest"
	"github.com/gopact-ai/gopact/graph"
)

type localAgent struct {
	card      a2a.AgentCard
	output    string
	artifacts []gopact.ArtifactRef
	events    []a2a.TaskEvent
}

type memoryCheckpointStore struct {
	latest map[string]graph.Checkpoint[clusterState]
}

type clusterState struct {
	Input        string
	Results      map[string]string
	Artifacts    []gopact.ArtifactRef
	ReviewLabels []string
	PolicyLabels []string
	Trace        []string
}

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	registry := a2a.NewRegistry()
	agents := []localAgent{
		{
			card:      a2a.AgentCard{Name: "planner-agent", Capabilities: []string{"planning"}},
			output:    "plan: research -> code -> review",
			artifacts: []gopact.ArtifactRef{{ID: "planner-plan", Name: "plan.md", URI: "memory://planner-plan"}},
		},
		{
			card:      a2a.AgentCard{Name: "research-agent", Capabilities: []string{"research"}},
			output:    "research: graph, a2a, examples",
			artifacts: []gopact.ArtifactRef{{ID: "research-notes", Name: "research.md", URI: "memory://research-notes"}},
		},
		{
			card:      a2a.AgentCard{Name: "code-agent", Capabilities: []string{"code.write"}},
			output:    "code: prepare a small tested patch",
			artifacts: []gopact.ArtifactRef{{ID: "code-patch", Name: "patch.diff", URI: "memory://code-patch"}},
		},
		{
			card:   a2a.AgentCard{Name: "review-agent", Capabilities: []string{"code.review"}},
			output: "review: pass",
			events: []a2a.TaskEvent{
				{Status: a2a.TaskStatusRunning, Message: "reviewing evidence"},
				{Status: a2a.TaskStatusCompleted, Result: &a2a.Result{Output: "review: pass"}},
			},
		},
	}
	servers := make([]*httptest.Server, 0, len(agents))
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()
	cards := make([]a2a.AgentCard, 0, len(agents))
	for i := range agents {
		server := httptest.NewServer(a2a.NewHTTPHandler(agents[i]))
		servers = append(servers, server)
		remote, err := a2a.NewHTTPAgent(server.URL, a2a.WithHTTPClient(server.Client()))
		if err != nil {
			return err
		}
		discovered, err := remote.Discover(ctx, a2a.DiscoveryQuery{URL: server.URL})
		if err != nil {
			return err
		}
		callable, err := a2a.NewHTTPAgent(server.URL,
			a2a.WithHTTPClient(server.Client()),
			a2a.WithHTTPAgentCard(discovered.Card),
		)
		if err != nil {
			return err
		}
		if err := registry.Register(ctx, callable); err != nil {
			return err
		}
		cards = append(cards, discovered.Card)
	}

	if _, err := fmt.Fprintln(out, "gateway: accepted self-bootstrap slice"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "discovery: %d HTTP agent cards\n", len(cards)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "cards: %s\n", cardNames(cards)); err != nil {
		return err
	}

	workflow, err := newAgentClusterWorkflow(registry)
	if err != nil {
		return err
	}
	state := clusterState{Input: "self-bootstrap slice", Results: map[string]string{}}
	ids := gopact.RuntimeIDs{RunID: "agent-cluster-demo", ThreadID: "agent-cluster-thread"}
	checkpoints := &memoryCheckpointStore{latest: map[string]graph.Checkpoint[clusterState]{}}
	recorder := gopact.NewRunRecorder()
	events := []string{}
	for event, err := range workflow.Run(ctx, state,
		graph.WithRuntimeIDs(ids),
		graph.WithCheckpointStore[clusterState](checkpoints),
	) {
		if err != nil {
			return err
		}
		if err := recorder.Record(event); err != nil {
			return err
		}
		events = append(events, eventLabel(event))
		if event.Type == gopact.EventNodeCompleted {
			if next, ok := event.StepSnapshot.Output.(clusterState); ok {
				state = next
			}
		}
	}
	export, err := recorder.Export()
	if err != nil {
		return err
	}
	diffChecks, diffSummary, err := worktreeDiffChecks(ctx, ".")
	if err != nil {
		return err
	}
	releaseGate, err := gopacttest.BuildSelfBootstrapReleaseGateBundle(
		export,
		gopacttest.WithSelfBootstrapAdditionalChecks(diffChecks...),
	)
	if err != nil {
		return err
	}
	if err := requireSelfBootstrapReleaseGate(ctx, releaseGate); err != nil {
		return err
	}
	export = releaseGate.RunExport
	resumeEvents, resumed, err := checkpointResume(ctx, workflow, checkpoints, ids)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(out, "workflow events: %s\n", strings.Join(events, " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "run export: %s events=%d steps=%d verification_reports=%d\n", export.Outcome, len(export.Events), len(export.Steps), len(export.VerificationReports)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "git diff evidence: %s\n", diffSummary); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "release gate: %s checks=%d requirements=%d\n", releaseGate.Report.Status, len(releaseGate.Report.Checks), len(gopacttest.SelfBootstrapReleaseGateRequirements())); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "checkpoint resume: loaded %s step=%d events=%s\n", resumed.Node, resumed.Step, strings.Join(resumeEvents, " -> ")); err != nil {
		return err
	}
	for _, name := range []string{"planner-agent", "research-agent", "code-agent"} {
		if _, err := fmt.Fprintf(out, "%s: %s\n", name, state.Results[name]); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(out, "artifacts: %s\n", artifactLabels(state.Artifacts)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "review stream: %s\n", strings.Join(state.ReviewLabels, " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "policy events: %s\n", strings.Join(state.PolicyLabels, " -> ")); err != nil {
		return err
	}
	failureLine, err := missingAgentFailureAttribution(ctx, registry, ids)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "failure attribution: %s\n", failureLine); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "agent trace: %s\n", strings.Join(state.Trace, " -> ")); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, "summary: local agent cluster completed 4 calls")
	return err
}

func worktreeDiffChecks(ctx context.Context, dir string) ([]gopact.VerificationCheck, string, error) {
	snapshot, err := gitdiff.ScanWorktree(ctx, dir)
	if err != nil {
		return nil, "", err
	}
	if snapshot.Skipped {
		if snapshot.Summary == "" {
			return nil, "no diff", nil
		}
		return nil, snapshot.Summary, nil
	}

	recorder := gopact.NewVerificationRecorder()
	if err := gopacttest.RecordDiffCheck(recorder, snapshot); err != nil {
		return nil, "", err
	}
	checks := recorder.Checks()
	if len(checks) != 1 {
		return nil, "", fmt.Errorf("git diff checks=%d, want 1", len(checks))
	}
	return checks, diffCheckSummary(checks[0]), nil
}

func diffCheckSummary(check gopact.VerificationCheck) string {
	if len(check.Evidence) > 0 && check.Evidence[0].Summary != "" {
		return check.Evidence[0].Summary
	}
	return check.Summary
}

func requireSelfBootstrapReleaseGate(ctx context.Context, gate gopacttest.SelfBootstrapReleaseGateBundle) error {
	for _, result := range gopacttest.CheckSelfBootstrapReleaseGate(ctx, gate.RunExport, gate.Report) {
		if !result.Passed {
			return fmt.Errorf("self-bootstrap release gate %s: %w", result.Case, result.Err)
		}
	}
	return nil
}

func newAgentClusterWorkflow(registry *a2a.Registry) (*graph.Runnable[clusterState], error) {
	g := graph.New[clusterState]()
	for _, name := range []string{"planner-agent", "research-agent", "code-agent"} {
		g.AddNode(name, agentCallNode(registry, name))
	}
	g.AddNode("review-agent", func(ctx context.Context, state clusterState) (clusterState, error) {
		policy := gopact.PolicyFunc(func(context.Context, gopact.PolicyRequest) (gopact.PolicyDecision, error) {
			return gopact.PolicyDecision{Action: gopact.PolicyAllow, Reason: "local demo allow"}, nil
		})
		policyLabels, err := authorizeA2AStream(ctx, policy, "review-agent", "review-task")
		if err != nil {
			return state, err
		}
		labels, err := collectReviewStream(ctx, registry)
		if err != nil {
			return state, err
		}
		state.ReviewLabels = labels
		state.PolicyLabels = policyLabels
		state.Trace = append(state.Trace, "review-agent")
		return state, nil
	})
	g.AddEdge(graph.Start, "planner-agent")
	g.AddEdge("planner-agent", "research-agent")
	g.AddEdge("research-agent", "code-agent")
	g.AddEdge("code-agent", "review-agent")
	g.AddEdge("review-agent", graph.End)
	return g.Compile()
}

func checkpointResume(ctx context.Context, workflow *graph.Runnable[clusterState], store *memoryCheckpointStore, ids gopact.RuntimeIDs) ([]string, gopact.StepSnapshot, error) {
	events := []string{}
	var loaded gopact.StepSnapshot
	for event, err := range workflow.Run(ctx, clusterState{}, graph.WithRuntimeIDs(ids), graph.WithCheckpointStore[clusterState](store)) {
		if err != nil {
			return nil, gopact.StepSnapshot{}, err
		}
		events = append(events, eventLabel(event))
		if event.Type == gopact.EventCheckpointLoaded && event.StepSnapshot != nil {
			loaded = *event.StepSnapshot
		}
	}
	return events, loaded, nil
}

func missingAgentFailureAttribution(ctx context.Context, registry *a2a.Registry, ids gopact.RuntimeIDs) (string, error) {
	task := a2a.Task{ID: "missing-agent-task", IDs: ids, Input: "route to missing agent"}
	_, err := registry.Send(ctx, "missing-agent", task)
	if !errors.Is(err, a2a.ErrAgentNotFound) {
		if err == nil {
			return "", fmt.Errorf("missing-agent unexpectedly resolved")
		}
		return "", err
	}

	recorder := gopact.NewVerificationRecorder()
	attribution := gopact.FailureAttribution{
		ID:      "missing-agent",
		Kind:    gopact.FailureExternal,
		IDs:     ids,
		Node:    "missing-agent",
		Summary: "remote agent unavailable",
		Error:   err.Error(),
		Evidence: []gopact.VerificationEvidence{{
			Type:    "a2a_task",
			Ref:     task.ID,
			Summary: "registry did not resolve missing-agent",
		}},
		Metadata: map[string]any{"agent_name": "missing-agent"},
	}
	if recordErr := gopact.RecordFailureAttributionCheck(recorder, attribution); !errors.Is(recordErr, gopact.ErrFailureAttributionFailed) {
		return "", recordErr
	}
	checks := recorder.Checks()
	if len(checks) != 1 {
		return "", fmt.Errorf("failure attribution checks=%d, want 1", len(checks))
	}
	return fmt.Sprintf("%s %s check=%s", attribution.Kind, attribution.Node, checks[0].ID), nil
}

func agentCallNode(registry *a2a.Registry, name string) graph.NodeFunc[clusterState] {
	return func(ctx context.Context, state clusterState) (clusterState, error) {
		result, err := registry.Send(ctx, name, a2a.Task{ID: name + "-task", Input: state.Input})
		if err != nil {
			return state, err
		}
		if state.Results == nil {
			state.Results = map[string]string{}
		}
		state.Results[name] = result.Output
		state.Artifacts = append(state.Artifacts, result.Artifacts...)
		state.Trace = append(state.Trace, name)
		return state, nil
	}
}

func collectReviewStream(ctx context.Context, registry *a2a.Registry) ([]string, error) {
	labels := []string{}
	for event, err := range registry.Stream(ctx, "review-agent", a2a.Task{ID: "review-task", Input: "review the slice"}) {
		if err != nil {
			return nil, err
		}
		label := string(event.Status)
		if event.Message != "" {
			label += "(" + event.Message + ")"
		}
		if event.Result != nil && event.Result.Output != "" {
			label += "(" + event.Result.Output + ")"
		}
		labels = append(labels, label)
	}
	return labels, nil
}

func authorizeA2AStream(ctx context.Context, policy gopact.Policy, agentName string, taskID string) ([]string, error) {
	if policy == nil {
		return nil, fmt.Errorf("policy is required")
	}
	req := gopact.PolicyRequest{
		Boundary: gopact.PolicyBoundaryA2A,
		Action:   gopact.PolicyActionStream,
		Input:    map[string]string{"agent_name": agentName, "task_id": taskID},
	}
	events := []gopact.Event{gopact.NewPolicyRequestedEvent(req)}
	decision, err := policy.Decide(ctx, req)
	if err != nil {
		return eventLabels(events), err
	}
	events = append(events, gopact.NewPolicyDecidedEvent(req, decision))
	if !decision.Allowed() {
		return eventLabels(events), fmt.Errorf("policy denied: %s", decision.Reason)
	}
	return eventLabels(events), nil
}

func cardNames(cards []a2a.AgentCard) string {
	names := make([]string, 0, len(cards))
	for _, card := range cards {
		names = append(names, card.Name)
	}
	return strings.Join(names, ", ")
}

func artifactLabels(artifacts []gopact.ArtifactRef) string {
	labels := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		labels = append(labels, fmt.Sprintf("%s(%s)", artifact.Name, artifact.URI))
	}
	return strings.Join(labels, " -> ")
}

func eventLabels(events []gopact.Event) []string {
	labels := make([]string, 0, len(events))
	for _, event := range events {
		labels = append(labels, string(event.Type))
	}
	return labels
}

func eventLabel(event gopact.Event) string {
	if event.Node == "" {
		return string(event.Type)
	}
	return fmt.Sprintf("%s(%s)", event.Type, event.Node)
}

func (a localAgent) Card() a2a.AgentCard {
	return a.card
}

func (a localAgent) Send(ctx context.Context, task a2a.Task) (a2a.Result, error) {
	if err := ctx.Err(); err != nil {
		return a2a.Result{}, err
	}
	return a2a.Result{TaskID: task.ID, Output: a.output, Artifacts: a.artifacts}, nil
}

func (a localAgent) Stream(ctx context.Context, task a2a.Task) iter.Seq2[a2a.TaskEvent, error] {
	return func(yield func(a2a.TaskEvent, error) bool) {
		if err := ctx.Err(); err != nil {
			yield(a2a.TaskEvent{TaskID: task.ID, Status: a2a.TaskStatusFailed, Err: err}, err)
			return
		}
		for _, event := range a.events {
			if event.TaskID == "" {
				event.TaskID = task.ID
			}
			if !yield(event, nil) {
				return
			}
		}
	}
}

func (a localAgent) Cancel(ctx context.Context, taskID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if taskID == "" {
		return a2a.ErrTaskIDRequired
	}
	return nil
}

func (s *memoryCheckpointStore) Put(ctx context.Context, checkpoint graph.Checkpoint[clusterState]) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.latest == nil {
		s.latest = map[string]graph.Checkpoint[clusterState]{}
	}
	s.latest[checkpoint.ThreadID] = checkpoint
	return nil
}

func (s *memoryCheckpointStore) Latest(ctx context.Context, threadID string) (graph.Checkpoint[clusterState], bool, error) {
	if err := ctx.Err(); err != nil {
		return graph.Checkpoint[clusterState]{}, false, err
	}
	checkpoint, ok := s.latest[threadID]
	return checkpoint, ok, nil
}
