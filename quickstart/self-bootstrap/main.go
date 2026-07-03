package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-examples/internal/exampleenv"
	"github.com/gopact-ai/gopact-ext/devagent/selfbootstrap"
	"github.com/gopact-ai/gopact/gopacttest"
)

type demoResult struct {
	Workflow selfbootstrap.Result
}

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	result, err := runDemo(ctx)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "self-bootstrap: dev agent workflow"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "objective: ship a tested SDK slice"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "workflow: %s\n", strings.Join(stepNodes(result.Workflow.RunExport), " -> ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "evidence: %s\n", strings.Join(evidenceTypes(result.Workflow.Report), ", ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "report: %s checks=%d failures=%d\n",
		result.Workflow.Report.Status,
		len(result.Workflow.Report.Checks),
		len(result.Workflow.RunExport.Failures),
	); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, "summary: release-ready self-bootstrap slice")
	return err
}

func runDemo(ctx context.Context) (demoResult, error) {
	if err := exampleenv.LoadDotEnv(); err != nil {
		return demoResult{}, err
	}

	workflow, err := selfbootstrap.New(
		selfbootstrap.WithAnalyzer(selfbootstrap.AnalyzerFunc(func(context.Context, selfbootstrap.Request) (selfbootstrap.Analysis, error) {
			return selfbootstrap.Analysis{
				Summary:  "scope is small, testable, and provider-neutral",
				Metadata: map[string]any{"stage": "analysis"},
			}, nil
		})),
		selfbootstrap.WithPlanner(selfbootstrap.PlannerFunc(func(context.Context, selfbootstrap.PlanRequest) (selfbootstrap.Plan, error) {
			return selfbootstrap.Plan{
				Summary: "implement one tested self-bootstrap slice",
				Steps: []selfbootstrap.PlanStep{
					{ID: "write", Summary: "produce local code evidence"},
					{ID: "test", Summary: "record command and CI gate evidence"},
					{ID: "review", Summary: "capture explicit review decision"},
				},
			}, nil
		})),
		selfbootstrap.WithWriter(selfbootstrap.WriterFunc(func(context.Context, selfbootstrap.WriteRequest) (selfbootstrap.WriteResult, error) {
			return selfbootstrap.WriteResult{
				Summary: "quickstart patch observed",
				Diff: &gopacttest.DiffSnapshot{
					ID:         "diff:self-bootstrap-quickstart",
					Ref:        "git:worktree",
					Diff:       "diff --git a/quickstart/self-bootstrap/main.go b/quickstart/self-bootstrap/main.go\n",
					Files:      []string{"quickstart/self-bootstrap/main.go"},
					Insertions: 64,
				},
				FileSnapshots: []gopacttest.FileSnapshot{
					{
						ID:            "file-snapshot:quickstart/self-bootstrap/main.go",
						Path:          "quickstart/self-bootstrap/main.go",
						Hash:          "demo-sha256",
						HashAlgorithm: "sha256",
						SizeBytes:     2048,
					},
				},
			}, nil
		})),
		selfbootstrap.WithTester(selfbootstrap.TesterFunc(func(context.Context, selfbootstrap.TestRequest) (selfbootstrap.TestResult, error) {
			command := gopacttest.CommandResult{
				ID:       "command:go test -count=1 ./quickstart/self-bootstrap",
				Command:  []string{"go", "test", "-count=1", "./quickstart/self-bootstrap"},
				ExitCode: 0,
			}
			return selfbootstrap.TestResult{
				Summary:       "mock self-bootstrap gate passed",
				Commands:      []gopacttest.CommandResult{command},
				RequiredGates: []string{gopacttest.SelfBootstrapCIGateUnit},
				Gates: []gopacttest.CIGateResult{
					{Gate: gopacttest.SelfBootstrapCIGateUnit, Result: command},
				},
			}, nil
		})),
		selfbootstrap.WithReviewer(selfbootstrap.ReviewerFunc(func(context.Context, selfbootstrap.ReviewRequest) (gopacttest.ReviewResult, error) {
			return gopacttest.ReviewResult{
				ID:       "review:self-bootstrap-quickstart",
				Reviewer: "local-reviewer",
				Source:   "mock",
				Status:   gopacttest.ReviewStatusApproved,
				Summary:  "approved",
			}, nil
		})),
	)
	if err != nil {
		return demoResult{}, err
	}

	result, err := workflow.Run(ctx, selfbootstrap.Request{
		Objective:  "ship a tested SDK slice",
		Repository: "gopact-examples",
		IDs: gopact.RuntimeIDs{
			RunID:    "self-bootstrap-quickstart",
			ThreadID: "quickstart",
			AgentID:  "devagent-selfbootstrap",
		},
	})
	if err != nil {
		return demoResult{}, err
	}
	return demoResult{Workflow: result}, nil
}

func stepNodes(export gopact.RunExport) []string {
	nodes := make([]string, 0, len(export.Steps))
	for _, step := range export.Steps {
		nodes = append(nodes, step.Node)
	}
	return nodes
}

func evidenceTypes(report gopact.VerificationReport) []string {
	seen := map[string]bool{}
	for _, check := range report.Checks {
		for _, evidence := range check.Evidence {
			seen[evidence.Type] = true
		}
	}
	types := make([]string, 0, len(seen))
	for evidenceType := range seen {
		types = append(types, evidenceType)
	}
	sort.Strings(types)
	return types
}
