package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-examples/internal/exampleenv"
	"github.com/gopact-ai/gopact-ext/devagent/selfbootstrap"
	"github.com/gopact-ai/gopact-ext/devagent/workspace"
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
	if _, err := fmt.Fprintln(out, "workspace: temp git repo + local go test gate"); err != nil {
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
	root, cleanup, err := prepareWorkspace(ctx)
	if err != nil {
		return demoResult{}, err
	}
	defer cleanup()

	ws, err := workspace.New(root, workspace.WithMetadata(map[string]any{"quickstart": "self-bootstrap"}))
	if err != nil {
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
		selfbootstrap.WithWriter(ws.Writer("hello.go")),
		selfbootstrap.WithTester(ws.Tester(workspace.Command{
			Gate: gopacttest.SelfBootstrapCIGateUnit,
			Args: []string{"go", "test", "./..."},
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

func prepareWorkspace(ctx context.Context) (string, func(), error) {
	root, err := os.MkdirTemp("", "gopact-self-bootstrap-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp workspace: %w", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(root)
	}
	if err := writeWorkspaceFile(root, "go.mod", "module example.test/selfbootstrap\n\ngo 1.25\n"); err != nil {
		cleanup()
		return "", nil, err
	}
	initial := "package hello\n\nfunc Message() string {\n\treturn \"hello\"\n}\n"
	if err := writeWorkspaceFile(root, "hello.go", initial); err != nil {
		cleanup()
		return "", nil, err
	}
	test := "package hello\n\nimport \"testing\"\n\nfunc TestMessage(t *testing.T) {\n\tif Message() != \"hello workspace\" {\n\t\tt.Fatalf(\"Message() = %q\", Message())\n\t}\n}\n"
	if err := writeWorkspaceFile(root, "hello_test.go", test); err != nil {
		cleanup()
		return "", nil, err
	}
	if err := runGit(ctx, root, "init"); err != nil {
		cleanup()
		return "", nil, err
	}
	if err := runGit(ctx, root, "add", "."); err != nil {
		cleanup()
		return "", nil, err
	}
	if err := runGit(ctx, root, "-c", "user.name=gopact", "-c", "user.email=gopact@example.test", "commit", "-m", "initial"); err != nil {
		cleanup()
		return "", nil, err
	}
	updated := "package hello\n\nfunc Message() string {\n\treturn \"hello workspace\"\n}\n"
	if err := writeWorkspaceFile(root, "hello.go", updated); err != nil {
		cleanup()
		return "", nil, err
	}
	return root, cleanup, nil
}

func writeWorkspaceFile(root, name, body string) error {
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create workspace dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return fmt.Errorf("write workspace file %s: %w", name, err)
	}
	return nil
}

func runGit(ctx context.Context, root string, args ...string) error {
	gitArgs := append([]string{"-c", "gc.auto=0", "-c", "maintenance.auto=false"}, args...)
	cmd := exec.CommandContext(ctx, "git", gitArgs...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
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
