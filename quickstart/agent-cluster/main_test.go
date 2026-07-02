package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact/a2a"
	"github.com/gopact-ai/gopact/gopacttest"
)

const wantDevAgentEvidencePurpose = "self-bootstrap-dev-agent"

func TestRunShowsLocalAgentCluster(t *testing.T) {
	t.Setenv("GOPACT_A2A_REGISTRY_FILE", " ")
	t.Setenv("GOPACT_A2A_REGISTRY_URL", " ")
	t.Setenv("GOPACT_A2A_ENDPOINTS", " ")
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"gateway: accepted self-bootstrap slice",
		"bootstrap discovery: 4 file registry agent cards",
		"cards: planner-agent, research-agent, code-agent, review-agent",
		"workflow events: run_started -> node_started(planner-agent) -> node_completed(planner-agent) -> node_started(research-agent) -> node_completed(research-agent) -> node_started(code-agent) -> node_completed(code-agent) -> node_started(review-agent) -> node_completed(review-agent) -> run_completed",
		"run export: completed events=10 steps=4 verification_reports=1",
		"git diff evidence:",
		"file snapshot evidence: sha256 ",
		"feature coverage evidence: sha256 ",
		"a2a task evidence: planner-agent completed -> research-agent completed -> code-agent completed -> review-agent completed",
		"a2a retry evidence: code-agent attempts=2",
		"dev agent evidence: unit gate passed -> review approved",
		"release gate: passed checks=",
		"requirements=15",
		"replay evidence: run-effect-replay:self-bootstrap run_effect_replay",
		"command evidence: command:(cd gopact-ext/models/agnes && go test -tags=integration -count=1 ./...)",
		"command:(cd gopact-ext/tests/agents && go test -tags=integration -count=1 ./...)",
		"command:(cd gopact-examples && go test -tags=integration -count=1 ./quickstart/agnes-chat)",
		"checkpoint resume: loaded review-agent step=4 events=run_started -> checkpoint_loaded(review-agent) -> run_completed",
		"planner-agent: plan: research -> code -> review",
		"research-agent: research: graph, a2a, examples",
		"artifacts: plan.md(memory://planner-plan) -> research.md(memory://research-notes) -> patch.diff(memory://code-patch)",
		"code-agent: code: prepare a small tested patch",
		"review stream: running(reviewing evidence) -> completed(review: pass)",
		"route fallback: code-agent missing tag -> code-agent completed",
		"cancel evidence: review-agent canceled",
		"a2a lease heartbeat: lease-agent renewed lease active evidence=a2a_agent_heartbeat",
		"policy events: policy_requested -> policy_decided",
		"policy deny: policy_requested -> policy_decided deny",
		"policy review: policy_requested -> policy_decided review",
		"failure attribution: external missing-agent check=failure-attribution:missing-agent",
		"agent trace: planner-agent -> research-agent -> code-agent -> review-agent",
		"summary: local agent cluster completed 4 calls",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}

func TestRunExportMatchesAgentClusterGoldenTrajectory(t *testing.T) {
	t.Setenv("GOPACT_A2A_REGISTRY_FILE", " ")
	t.Setenv("GOPACT_A2A_REGISTRY_URL", " ")
	t.Setenv("GOPACT_A2A_ENDPOINTS", " ")

	export, err := runCluster(context.Background(), io.Discard)
	if err != nil {
		t.Fatalf("runCluster() error = %v", err)
	}
	gopacttest.RequireRunExportGoldenTrajectoryFrames(t, "testdata/agent_cluster_run_export.golden.json", export)
}

func TestRunExportCarriesDevAgentEvidenceMetadata(t *testing.T) {
	t.Setenv("GOPACT_A2A_REGISTRY_FILE", " ")
	t.Setenv("GOPACT_A2A_REGISTRY_URL", " ")
	t.Setenv("GOPACT_A2A_ENDPOINTS", " ")

	export, err := runCluster(context.Background(), io.Discard)
	if err != nil {
		t.Fatalf("runCluster() error = %v", err)
	}
	if len(export.VerificationReports) != 1 {
		t.Fatalf("VerificationReports = %+v, want one release gate report", export.VerificationReports)
	}
	report := export.VerificationReports[0]

	ciGateCheck := requireVerificationCheck(t, report, "ci-gates:dev-agent-local")
	requireEvidenceMetadata(t, ciGateCheck, gopacttest.VerificationEvidenceTypeCIGate, "purpose", wantDevAgentEvidencePurpose)

	reviewCheck := requireVerificationCheck(t, report, "review:dev-agent-local")
	requireEvidenceMetadata(t, reviewCheck, gopacttest.VerificationEvidenceTypeReview, "purpose", wantDevAgentEvidencePurpose)
}

func TestRunExportCarriesReplayAndCommandEvidence(t *testing.T) {
	t.Setenv("GOPACT_A2A_REGISTRY_FILE", " ")
	t.Setenv("GOPACT_A2A_REGISTRY_URL", " ")
	t.Setenv("GOPACT_A2A_ENDPOINTS", " ")

	export, err := runCluster(context.Background(), io.Discard)
	if err != nil {
		t.Fatalf("runCluster() error = %v", err)
	}
	if len(export.VerificationReports) != 1 {
		t.Fatalf("VerificationReports = %+v, want one release gate report", export.VerificationReports)
	}
	report := export.VerificationReports[0]

	replayCheck := requireVerificationCheck(t, report, gopacttest.SelfBootstrapCheckRunEffectReplay)
	requireEvidenceType(t, replayCheck, gopact.VerificationEvidenceTypeRunEffectReplay)

	for _, id := range []string{
		gopacttest.SelfBootstrapCheckAgnesProviderIntegrationCommand,
		gopacttest.SelfBootstrapCheckAgnesAgentTemplatesIntegrationCommand,
		gopacttest.SelfBootstrapCheckAgnesExamplesIntegrationCommand,
	} {
		check := requireVerificationCheck(t, report, id)
		requireEvidenceType(t, check, gopacttest.VerificationEvidenceTypeCommand)
	}
}

func TestRunBootstrapsConfiguredFileRegistry(t *testing.T) {
	t.Setenv("GOPACT_A2A_ENDPOINTS", " ")
	t.Setenv("GOPACT_A2A_REGISTRY_URL", " ")
	cards, _ := startTestAgentServers(t, testClusterAgents())
	for i := range cards {
		cards[i].Tags = nil
	}
	t.Setenv("GOPACT_A2A_REGISTRY_FILE", writeAgentRegistry(t, cards))

	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"bootstrap discovery: 4 configured file registry agent cards",
		"cards: planner-agent, research-agent, code-agent, review-agent",
		"summary: local agent cluster completed 4 calls",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}

func TestRunBootstrapsConfiguredHTTPEndpoints(t *testing.T) {
	_, endpoints := startTestAgentServers(t, testClusterAgents())
	t.Setenv("GOPACT_A2A_REGISTRY_FILE", " ")
	t.Setenv("GOPACT_A2A_REGISTRY_URL", " ")
	t.Setenv("GOPACT_A2A_ENDPOINTS", strings.Join(endpoints, ","))

	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"bootstrap discovery: 4 configured HTTP endpoints agent cards",
		"cards: planner-agent, research-agent, code-agent, review-agent",
		"summary: local agent cluster completed 4 calls",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}

func TestConfiguredHTTPEndpointDiscoveryRequiresReadiness(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/agent-card.json":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(a2a.AgentCard{
				Name:   "planner-agent",
				Health: &a2a.HealthHints{ReadinessPath: "/readyz"},
			})
		case "/readyz":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "not_ready"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Setenv("GOPACT_A2A_REGISTRY_FILE", " ")
	t.Setenv("GOPACT_A2A_REGISTRY_URL", " ")
	t.Setenv("GOPACT_A2A_ENDPOINTS", server.URL)

	mesh, err := a2a.NewMesh()
	if err != nil {
		t.Fatalf("NewMesh() error = %v", err)
	}
	_, _, cleanup, err := bootstrapAgentDiscovery(context.Background(), mesh, nil)
	defer cleanup()
	if !errors.Is(err, a2a.ErrHTTPStatus) {
		t.Fatalf("bootstrapAgentDiscovery() error = %v, want ErrHTTPStatus", err)
	}
}

func TestRunBootstrapsConfiguredHTTPRegistryURL(t *testing.T) {
	cards, _ := startTestAgentServers(t, testClusterAgents())
	registry := httptest.NewServer(a2a.NewHTTPRegistryHandler(a2a.NewStaticDiscoverer(cards...)))
	defer registry.Close()
	t.Setenv("GOPACT_A2A_REGISTRY_FILE", " ")
	t.Setenv("GOPACT_A2A_REGISTRY_URL", registry.URL+"/agents.json")
	t.Setenv("GOPACT_A2A_ENDPOINTS", " ")

	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"bootstrap discovery: 4 configured HTTP registry agent cards",
		"cards: planner-agent, research-agent, code-agent, review-agent",
		"summary: local agent cluster completed 4 calls",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}

func TestRunBootstrapsConfiguredDiscoverySources(t *testing.T) {
	cards, endpoints := startTestAgentServers(t, testClusterAgents())
	registryFile := writeAgentRegistry(t, cards[:2])
	registry := httptest.NewServer(a2a.NewHTTPRegistryHandler(a2a.NewStaticDiscoverer(cards[2])))
	defer registry.Close()

	t.Setenv("GOPACT_A2A_REGISTRY_FILE", registryFile)
	t.Setenv("GOPACT_A2A_REGISTRY_URL", registry.URL+"/agents.json")
	t.Setenv("GOPACT_A2A_ENDPOINTS", endpoints[3])

	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"bootstrap discovery: 4 configured discovery sources agent cards",
		"cards: planner-agent, research-agent, code-agent, review-agent",
		"summary: local agent cluster completed 4 calls",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}

func TestAgentClusterSkipsExpiredDiscoveredAgents(t *testing.T) {
	ctx := context.Background()
	mesh, err := a2a.NewMesh()
	if err != nil {
		t.Fatalf("NewMesh() error = %v", err)
	}
	expired := a2a.FakeAgent{
		CardValue: a2a.AgentCard{
			Name:         "expired-code-agent",
			Capabilities: []string{"code.write"},
			ExpiresAt:    time.Now().Add(-time.Minute),
		},
		SendFunc: func(context.Context, a2a.Task) (a2a.Result, error) {
			return a2a.Result{Output: "expired"}, nil
		},
	}
	active := a2a.FakeAgent{
		CardValue: a2a.AgentCard{
			Name:         "active-code-agent",
			Capabilities: []string{"code.write"},
			ExpiresAt:    time.Now().Add(time.Minute),
		},
		SendFunc: func(context.Context, a2a.Task) (a2a.Result, error) {
			return a2a.Result{Output: "active"}, nil
		},
	}
	if _, err := mesh.Register(ctx, expired); err != nil {
		t.Fatalf("Register(expired) error = %v", err)
	}
	if _, err := mesh.Register(ctx, active); err != nil {
		t.Fatalf("Register(active) error = %v", err)
	}

	cards, err := mesh.ListCards(ctx)
	if err != nil {
		t.Fatalf("ListCards() error = %v", err)
	}
	if len(cards) != 1 || cards[0].Name != "active-code-agent" {
		t.Fatalf("ListCards() = %+v, want only active-code-agent", cards)
	}

	result, err := mesh.Route(ctx, a2a.RouteQuery{
		Require: []string{"code.write"},
		Task:    a2a.Task{ID: "task-1"},
	})
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}
	if result.Output != "active" {
		t.Fatalf("Route() output = %q, want active", result.Output)
	}
}

func testClusterAgents() []localAgent {
	return []localAgent{
		{card: a2a.AgentCard{Name: "planner-agent", Capabilities: []string{"planning"}}, output: "plan: research -> code -> review"},
		{card: a2a.AgentCard{Name: "research-agent", Capabilities: []string{"research"}}, output: "research: graph, a2a, examples"},
		{card: a2a.AgentCard{Name: "code-agent", Capabilities: []string{"code.write"}, Tags: []string{"code", "local"}}, output: "code: prepare a small tested patch"},
		{
			card:   a2a.AgentCard{Name: "review-agent", Capabilities: []string{"code.review"}},
			output: "review: pass",
			events: []a2a.TaskEvent{
				{Status: a2a.TaskStatusRunning, Message: "reviewing evidence"},
				{Status: a2a.TaskStatusCompleted, Result: &a2a.Result{Output: "review: pass"}},
			},
		},
	}
}

func writeAgentRegistry(t *testing.T, cards []a2a.AgentCard) string {
	t.Helper()

	raw, err := json.Marshal(cards)
	if err != nil {
		t.Fatalf("Marshal(cards) error = %v", err)
	}
	path := t.TempDir() + "/agents.json"
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("WriteFile(registry) error = %v", err)
	}
	return path
}

func startTestAgentServers(t *testing.T, agents []localAgent) ([]a2a.AgentCard, []string) {
	t.Helper()

	servers := make([]*httptest.Server, 0, len(agents))
	t.Cleanup(func() {
		for _, server := range servers {
			server.Close()
		}
	})
	cards := make([]a2a.AgentCard, 0, len(agents))
	endpoints := make([]string, 0, len(agents))
	for _, agent := range agents {
		server := httptest.NewServer(a2a.NewHTTPHandler(agent))
		servers = append(servers, server)
		card := agent.Card()
		card.URL = server.URL
		cards = append(cards, card)
		endpoints = append(endpoints, server.URL)
	}
	return cards, endpoints
}

func requireVerificationCheck(t *testing.T, report gopact.VerificationReport, id string) gopact.VerificationCheck {
	t.Helper()

	for _, check := range report.Checks {
		if check.ID == id {
			return check
		}
	}
	t.Fatalf("verification report checks missing %q: %+v", id, report.Checks)
	return gopact.VerificationCheck{}
}

func requireEvidenceMetadata(t *testing.T, check gopact.VerificationCheck, evidenceType, key string, want any) {
	t.Helper()

	for _, evidence := range check.Evidence {
		if evidence.Type != evidenceType {
			continue
		}
		if got := evidence.Metadata[key]; got != want {
			t.Fatalf("%s evidence metadata[%q] = %v, want %v", evidenceType, key, got, want)
		}
		return
	}
	t.Fatalf("check %q evidence missing type %q: %+v", check.ID, evidenceType, check.Evidence)
}

func requireEvidenceType(t *testing.T, check gopact.VerificationCheck, evidenceType string) {
	t.Helper()

	for _, evidence := range check.Evidence {
		if evidence.Type == evidenceType {
			return
		}
	}
	t.Fatalf("check %q evidence missing type %q: %+v", check.ID, evidenceType, check.Evidence)
}
