package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gopact-ai/gopact/a2a"
	"github.com/gopact-ai/gopact/gopacttest"
)

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
		"release gate: passed checks=",
		"requirements=14",
		"checkpoint resume: loaded review-agent step=4 events=run_started -> checkpoint_loaded(review-agent) -> run_completed",
		"planner-agent: plan: research -> code -> review",
		"research-agent: research: graph, a2a, examples",
		"artifacts: plan.md(memory://planner-plan) -> research.md(memory://research-notes) -> patch.diff(memory://code-patch)",
		"code-agent: code: prepare a small tested patch",
		"review stream: running(reviewing evidence) -> completed(review: pass)",
		"route fallback: code-agent missing tag -> code-agent completed",
		"cancel evidence: review-agent canceled",
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
