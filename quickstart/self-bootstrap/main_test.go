package main

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact/gopacttest"
)

func TestRunShowsSelfBootstrapWorkflow(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"self-bootstrap: dev agent workflow",
		"objective: ship a tested SDK slice",
		"workspace: temp git repo + policy-approved plan patch + local go test gate",
		"workflow: analyze -> plan -> write -> test -> review",
		"evidence: ci_gate, command, diff, file_snapshot, policy_decision, review, run_export",
		"report: passed checks=7 failures=0",
		"summary: release-ready self-bootstrap slice",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}

func TestRunDemoProducesReleaseReadyEvidence(t *testing.T) {
	result, err := runDemo(context.Background())
	if err != nil {
		t.Fatalf("runDemo() error = %v", err)
	}

	if result.Workflow.Report.Status != gopact.VerificationStatusPassed {
		t.Fatalf("report status = %q, want passed", result.Workflow.Report.Status)
	}
	if result.Workflow.RunExport.Outcome != gopact.RunCompleted {
		t.Fatalf("run outcome = %q, want completed", result.Workflow.RunExport.Outcome)
	}
	if len(result.Workflow.RunExport.Failures) != 0 {
		t.Fatalf("failures = %+v, want none", result.Workflow.RunExport.Failures)
	}
	if len(result.Workflow.RunExport.VerificationReports) != 1 {
		t.Fatalf("embedded reports = %d, want 1", len(result.Workflow.RunExport.VerificationReports))
	}
	requireNodes(t, result.Workflow.RunExport, []string{"analyze", "plan", "write", "test", "review"})
	requireEvidenceTypes(t, result.Workflow.Report, []string{
		gopact.VerificationEvidenceTypeRunExport,
		gopacttest.VerificationEvidenceTypeCIGate,
		gopacttest.VerificationEvidenceTypeCommand,
		gopacttest.VerificationEvidenceTypeDiff,
		gopacttest.VerificationEvidenceTypeFileSnapshot,
		gopacttest.VerificationEvidenceTypeReview,
		gopact.VerificationEvidenceTypePolicyDecision,
	})
	requireWorkspaceEvidence(t, result)
}

func requireNodes(t *testing.T, export gopact.RunExport, want []string) {
	t.Helper()
	var got []string
	for _, step := range export.Steps {
		got = append(got, step.Node)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("nodes = %+v, want %+v", got, want)
	}
}

func requireEvidenceTypes(t *testing.T, report gopact.VerificationReport, want []string) {
	t.Helper()
	got := map[string]bool{}
	for _, check := range report.Checks {
		for _, evidence := range check.Evidence {
			got[evidence.Type] = true
		}
	}
	for _, evidenceType := range want {
		if !got[evidenceType] {
			t.Fatalf("report missing evidence type %q; checks=%+v", evidenceType, report.Checks)
		}
	}
}

func requireWorkspaceEvidence(t *testing.T, result demoResult) {
	t.Helper()

	write := result.Workflow.Write
	if write.Diff == nil || len(write.Diff.Files) != 1 || write.Diff.Files[0] != "hello.go" {
		t.Fatalf("write diff = %+v, want hello.go workspace diff", write.Diff)
	}
	if len(write.FileSnapshots) != 1 || write.FileSnapshots[0].Path != "hello.go" {
		t.Fatalf("file snapshots = %+v, want repo-relative hello.go snapshot", write.FileSnapshots)
	}
	if write.FileSnapshots[0].Metadata["source"] != "workspace" ||
		write.FileSnapshots[0].Metadata["patch_id"] != "quickstart-hello-patch" ||
		write.FileSnapshots[0].Metadata["patch_applied"] != true ||
		write.FileSnapshots[0].Metadata["patch_policy_action"] != string(gopact.PolicyAllow) {
		t.Fatalf("snapshot metadata = %+v, want workspace patch metadata", write.FileSnapshots[0].Metadata)
	}
	if write.Metadata["patch_id"] != "quickstart-hello-patch" ||
		write.Metadata["patch_applied"] != true ||
		write.Metadata["patch_policy_action"] != string(gopact.PolicyAllow) ||
		write.Metadata["patch_policy_reason"] != "quickstart plan patch approved" {
		t.Fatalf("write metadata = %+v, want patch apply metadata", write.Metadata)
	}
	if result.Workflow.PatchDecision.Action != gopact.PolicyAllow ||
		result.Workflow.PatchDecision.Reason != "quickstart plan patch approved" {
		t.Fatalf("patch decision = %+v, want approved plan patch", result.Workflow.PatchDecision)
	}
	if result.Workflow.Plan.Patch == nil ||
		result.Workflow.Plan.Patch.ID != "quickstart-hello-patch" ||
		len(result.Workflow.Plan.Patch.Files) != 1 ||
		result.Workflow.Plan.Patch.Files[0].Path != "hello.go" {
		t.Fatalf("plan patch = %+v, want planner-owned patch proposal", result.Workflow.Plan.Patch)
	}

	test := result.Workflow.Test
	if len(test.Commands) != 1 || len(test.Commands[0].Command) != 3 {
		t.Fatalf("commands = %+v, want one go test command", test.Commands)
	}
	if strings.Join(test.Commands[0].Command, " ") != "go test ./..." {
		t.Fatalf("command = %+v, want go test ./...", test.Commands[0].Command)
	}
	if test.Commands[0].Dir != "." || test.Commands[0].ExitCode != 0 {
		t.Fatalf("command result = %+v, want successful repo-root command", test.Commands[0])
	}
}
