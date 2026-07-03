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
		"workflow: analyze -> plan -> write -> test -> review",
		"evidence: ci_gate, command, diff, file_snapshot, review, run_export",
		"report: passed checks=6 failures=0",
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
	})
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
