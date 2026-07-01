package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-examples/internal/exampleenv"
	"github.com/gopact-ai/gopact-ext/devagent/filesnapshot"
	"github.com/gopact-ai/gopact-ext/devagent/gitdiff"
	"github.com/gopact-ai/gopact/a2a"
	"github.com/gopact-ai/gopact/gopacttest"
	"github.com/gopact-ai/gopact/graph"
)

const (
	a2aRegistryFileEnv = "GOPACT_A2A_REGISTRY_FILE"
	a2aRegistryURLEnv  = "GOPACT_A2A_REGISTRY_URL"
	a2aEndpointsEnv    = "GOPACT_A2A_ENDPOINTS"
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
	A2AChecks    []gopact.VerificationCheck
	Trace        []string
}

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	if err := exampleenv.LoadDotEnv(); err != nil {
		return err
	}
	mesh, err := a2a.NewMesh()
	if err != nil {
		return err
	}
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
			card:      a2a.AgentCard{Name: "code-agent", Capabilities: []string{"code.write"}, Tags: []string{"code", "local"}},
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
	cards, discoverySource, cleanup, err := bootstrapAgentDiscovery(ctx, mesh, agents)
	defer cleanup()
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "gateway: accepted self-bootstrap slice"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "bootstrap discovery: %d %s agent cards\n", len(cards), discoverySource); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "cards: %s\n", cardNames(cards)); err != nil {
		return err
	}

	workflow, err := newAgentClusterWorkflow(mesh)
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
	fileChecks, fileSummary, err := fileSnapshotChecks(ctx, "go.mod")
	if err != nil {
		return err
	}
	featureChecks, featureSummary, err := fileSnapshotChecks(ctx, "FEATURES.md")
	if err != nil {
		return err
	}
	releaseGate, err := gopacttest.BuildSelfBootstrapReleaseGateBundle(
		export,
		gopacttest.WithSelfBootstrapAdditionalChecks(append(append(append(diffChecks, fileChecks...), featureChecks...), state.A2AChecks...)...),
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
	if _, err := fmt.Fprintf(out, "file snapshot evidence: %s\n", fileSummary); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "feature coverage evidence: %s\n", featureSummary); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "a2a task evidence: %s\n", a2aTaskEvidenceSummary(state.A2AChecks)); err != nil {
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
	fallbackLine, err := routeFallback(ctx, mesh, ids)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "route fallback: %s\n", fallbackLine); err != nil {
		return err
	}
	cancelLine, err := cancelReviewTask(ctx, mesh, ids)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "cancel evidence: %s\n", cancelLine); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "policy events: %s\n", strings.Join(state.PolicyLabels, " -> ")); err != nil {
		return err
	}
	failureLine, err := missingAgentFailureAttribution(ctx, mesh, ids)
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

func bootstrapAgentDiscovery(ctx context.Context, mesh *a2a.Mesh, agents []localAgent) ([]a2a.AgentCard, string, func(), error) {
	listers := []a2a.CardLister{}
	sources := []string{}
	if path := strings.TrimSpace(os.Getenv(a2aRegistryFileEnv)); path != "" {
		discoverer, err := a2a.NewFileDiscoverer(path)
		if err != nil {
			return nil, "", func() {}, err
		}
		listers = append(listers, discoverer)
		sources = append(sources, "file registry")
	}
	if registryURL := strings.TrimSpace(os.Getenv(a2aRegistryURLEnv)); registryURL != "" {
		registry, err := a2a.NewHTTPRegistry(registryURL)
		if err != nil {
			return nil, "", func() {}, err
		}
		listers = append(listers, registry)
		sources = append(sources, "HTTP registry")
	}
	if endpoints := envList(a2aEndpointsEnv); len(endpoints) > 0 {
		endpointListers, err := a2a.NewHTTPCardListers(endpoints)
		if err != nil {
			return nil, "", func() {}, err
		}
		listers = append(listers, endpointListers...)
		sources = append(sources, "HTTP endpoints")
	}
	if len(listers) > 0 {
		bootstrap, err := mesh.Bootstrap(ctx, listers...)
		if err != nil {
			return nil, "", func() {}, err
		}
		return bootstrap.Cards, configuredDiscoveryLabel(sources), func() {}, nil
	}

	servers := make([]*httptest.Server, 0, len(agents))
	cleanup := func() {
		for _, server := range servers {
			server.Close()
		}
	}
	registryFile, err := os.CreateTemp("", "gopact-agent-registry-*.json")
	if err != nil {
		return nil, "", cleanup, err
	}
	registryPath := registryFile.Name()
	cleanupFile := func() {
		cleanup()
		_ = os.Remove(registryPath)
	}
	cards := make([]a2a.AgentCard, 0, len(agents))
	for i := range agents {
		server := httptest.NewServer(a2a.NewHTTPHandler(agents[i]))
		servers = append(servers, server)
		card := agents[i].Card()
		card.URL = server.URL
		cards = append(cards, card)
	}
	if err := json.NewEncoder(registryFile).Encode(cards); err != nil {
		_ = registryFile.Close()
		return nil, "", cleanupFile, err
	}
	if err := registryFile.Close(); err != nil {
		return nil, "", cleanupFile, err
	}
	discoverer, err := a2a.NewFileDiscoverer(registryPath)
	if err != nil {
		return nil, "", cleanupFile, err
	}
	bootstrap, err := mesh.Bootstrap(ctx, discoverer)
	if err != nil {
		return nil, "", cleanupFile, err
	}
	return bootstrap.Cards, "file registry", cleanupFile, nil
}

func configuredDiscoveryLabel(sources []string) string {
	if len(sources) == 1 {
		return "configured " + sources[0]
	}
	return "configured discovery sources"
}

func envList(key string) []string {
	var values []string
	for _, part := range strings.Split(os.Getenv(key), ",") {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}
	return values
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
	return checks, checkEvidenceSummary(checks[0]), nil
}

func fileSnapshotChecks(ctx context.Context, path string) ([]gopact.VerificationCheck, string, error) {
	actualPath, err := findUp(path)
	if err != nil {
		return nil, "", err
	}
	snapshot, err := filesnapshot.Scan(ctx, actualPath)
	if err != nil {
		return nil, "", err
	}
	snapshot.Path = path

	recorder := gopact.NewVerificationRecorder()
	if err := gopacttest.RecordFileSnapshotCheck(recorder, snapshot); err != nil {
		return nil, "", err
	}
	checks := recorder.Checks()
	if len(checks) != 1 {
		return nil, "", fmt.Errorf("file snapshot checks=%d, want 1", len(checks))
	}
	return checks, checkEvidenceSummary(checks[0]), nil
}

func findUp(name string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%s not found", name)
		}
		dir = parent
	}
}

func checkEvidenceSummary(check gopact.VerificationCheck) string {
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

func newAgentClusterWorkflow(mesh *a2a.Mesh) (*graph.Runnable[clusterState], error) {
	g := graph.New[clusterState]()
	g.AddNode("planner-agent", agentCallNode(mesh, "planner-agent"))
	g.AddNode("research-agent", agentCallNode(mesh, "research-agent"))
	g.AddNode("code-agent", agentCallNode(mesh, "code-agent", "code", "local"))
	g.AddNode("review-agent", func(ctx context.Context, state clusterState) (clusterState, error) {
		policy := gopact.PolicyFunc(func(context.Context, gopact.PolicyRequest) (gopact.PolicyDecision, error) {
			return gopact.PolicyDecision{Action: gopact.PolicyAllow, Reason: "local demo allow"}, nil
		})
		policyLabels, err := authorizeA2AStream(ctx, policy, "review-agent", "review-task")
		if err != nil {
			return state, err
		}
		labels, checks, err := collectReviewStream(ctx, mesh)
		if err != nil {
			return state, err
		}
		state.ReviewLabels = labels
		state.A2AChecks = append(state.A2AChecks, checks...)
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

func missingAgentFailureAttribution(ctx context.Context, mesh *a2a.Mesh, ids gopact.RuntimeIDs) (string, error) {
	task := a2a.Task{ID: "missing-agent-task", IDs: ids, Input: "route to missing agent"}
	_, err := mesh.Call(ctx, "missing-agent", task)
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

func routeFallback(ctx context.Context, mesh *a2a.Mesh, ids gopact.RuntimeIDs) (string, error) {
	task := a2a.Task{ID: "code-fallback-task", IDs: ids, Input: "fallback route"}
	_, err := mesh.Route(ctx, a2a.RouteQuery{Tags: []string{"missing-code-route"}, Task: task})
	if !errors.Is(err, a2a.ErrAgentNotFound) {
		if err == nil {
			return "", fmt.Errorf("missing tag route unexpectedly resolved")
		}
		return "", err
	}

	result, err := mesh.Call(ctx, "code-agent", task)
	if err != nil {
		return "", err
	}
	check, err := recordA2ATaskCheck("code-agent", task, a2a.TaskEvent{
		TaskID: task.ID,
		IDs:    ids,
		Status: a2a.TaskStatusCompleted,
		Result: &result,
	})
	if err != nil {
		return "", err
	}
	return "code-agent missing tag -> " + checkEvidenceSummary(check), nil
}

func agentCallNode(mesh *a2a.Mesh, name string, tags ...string) graph.NodeFunc[clusterState] {
	return func(ctx context.Context, state clusterState) (clusterState, error) {
		task := a2a.Task{ID: name + "-task", IDs: runtimeIDs(ctx), Input: state.Input}
		var result a2a.Result
		var err error
		if len(tags) > 0 {
			result, err = mesh.Route(ctx, a2a.RouteQuery{Tags: tags, Task: task})
			if errors.Is(err, a2a.ErrAgentNotFound) {
				result, err = mesh.Call(ctx, name, task)
			}
		} else {
			result, err = mesh.Call(ctx, name, task)
		}
		if err != nil {
			return state, err
		}
		check, err := recordA2ATaskCheck(name, task, a2a.TaskEvent{
			TaskID: task.ID,
			Status: a2a.TaskStatusCompleted,
			Result: &result,
		})
		if err != nil {
			return state, err
		}
		if state.Results == nil {
			state.Results = map[string]string{}
		}
		state.Results[name] = result.Output
		state.Artifacts = append(state.Artifacts, result.Artifacts...)
		state.A2AChecks = append(state.A2AChecks, check)
		state.Trace = append(state.Trace, name)
		return state, nil
	}
}

func collectReviewStream(ctx context.Context, mesh *a2a.Mesh) ([]string, []gopact.VerificationCheck, error) {
	labels := []string{}
	checks := []gopact.VerificationCheck{}
	task := a2a.Task{ID: "review-task", IDs: runtimeIDs(ctx), Input: "review the slice"}
	for event, err := range mesh.Stream(ctx, "review-agent", task) {
		if err != nil {
			return nil, nil, err
		}
		label := string(event.Status)
		if event.Message != "" {
			label += "(" + event.Message + ")"
		}
		if event.Result != nil && event.Result.Output != "" {
			label += "(" + event.Result.Output + ")"
		}
		if event.Status == a2a.TaskStatusCompleted || event.Status == a2a.TaskStatusFailed || event.Status == a2a.TaskStatusCanceled {
			check, err := recordA2ATaskCheck("review-agent", task, event)
			if err != nil {
				return nil, nil, err
			}
			checks = append(checks, check)
		}
		labels = append(labels, label)
	}
	return labels, checks, nil
}

func cancelReviewTask(ctx context.Context, mesh *a2a.Mesh, ids gopact.RuntimeIDs) (string, error) {
	const taskID = "review-cancel-task"
	result, err := mesh.Cancel(gopact.ContextWithRuntimeIDs(ctx, ids), "review-agent", taskID)
	if err != nil {
		return "", err
	}
	if result.TaskID != taskID || len(result.Events) != 1 || result.Events[0].Type != gopact.EventA2ATaskCanceled {
		return "", fmt.Errorf("cancel result = %+v, want one canceled event", result)
	}
	check, err := recordTerminalA2ATaskCheck("review-agent", a2a.Task{ID: taskID, IDs: ids}, a2a.TaskEvent{
		TaskID: taskID,
		IDs:    ids,
		Status: a2a.TaskStatusCanceled,
	})
	if err != nil {
		return "", err
	}
	return checkEvidenceSummary(check), nil
}

func recordA2ATaskCheck(agentName string, task a2a.Task, event a2a.TaskEvent) (gopact.VerificationCheck, error) {
	recorder := gopact.NewVerificationRecorder()
	err := a2a.RecordTaskEventCheck(recorder, a2a.TaskEventSnapshot{
		Agent: a2a.AgentCard{Name: agentName},
		Task:  task,
		Event: event,
	})
	checks := recorder.Checks()
	if err != nil {
		return gopact.VerificationCheck{}, err
	}
	if len(checks) != 1 {
		return gopact.VerificationCheck{}, fmt.Errorf("a2a task checks=%d, want 1", len(checks))
	}
	return checks[0], nil
}

func recordTerminalA2ATaskCheck(agentName string, task a2a.Task, event a2a.TaskEvent) (gopact.VerificationCheck, error) {
	recorder := gopact.NewVerificationRecorder()
	err := a2a.RecordTaskEventCheck(recorder, a2a.TaskEventSnapshot{
		Agent: a2a.AgentCard{Name: agentName},
		Task:  task,
		Event: event,
	})
	if err != nil && !errors.Is(err, a2a.ErrTaskEventFailed) {
		return gopact.VerificationCheck{}, err
	}
	checks := recorder.Checks()
	if len(checks) != 1 {
		return gopact.VerificationCheck{}, fmt.Errorf("a2a terminal task checks=%d, want 1", len(checks))
	}
	return checks[0], nil
}

func runtimeIDs(ctx context.Context) gopact.RuntimeIDs {
	ids, _ := gopact.RuntimeIDsFromContext(ctx)
	return ids
}

func a2aTaskEvidenceSummary(checks []gopact.VerificationCheck) string {
	parts := make([]string, 0, len(checks))
	for _, check := range checks {
		parts = append(parts, checkEvidenceSummary(check))
	}
	return strings.Join(parts, " -> ")
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
