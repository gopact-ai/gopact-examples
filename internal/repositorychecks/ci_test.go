package repositorychecks

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

func TestExamplesCIMockGateIsDocumented(t *testing.T) {
	workflow := readText(t, "../../.github/workflows/ci.yml")
	readme := readText(t, "../../README.md")

	for _, command := range []string{
		"git diff --check",
		"go mod tidy",
		"git diff --exit-code",
		"go test -count=1 ./...",
		"go test -race -count=1 ./...",
		"go vet ./...",
		"golangci-lint run ./...",
		"go test -coverprofile=coverage.out ./...",
		"govulncheck ./...",
	} {
		if !strings.Contains(workflow, command) {
			t.Fatalf("workflow missing mock CI command %q", command)
		}
		if !strings.Contains(readme, command) {
			t.Fatalf("README missing mock CI command %q", command)
		}
	}
	for _, action := range []string{"actions/checkout@v7", "actions/setup-go@v6"} {
		if !strings.Contains(workflow, action) {
			t.Fatalf("workflow missing current GitHub Action %q", action)
		}
	}

	for _, forbidden := range []string{"-tags=integration", ".env"} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("workflow contains %q; examples CI must stay mock-only", forbidden)
		}
	}
}

func TestExamplesOpenSourceGovernanceDocsArePresent(t *testing.T) {
	readme := readText(t, "../../README.md")
	for _, doc := range []struct {
		path     string
		sections []string
	}{
		{
			path: "LICENSE",
			sections: []string{
				"MIT License",
				"Permission is hereby granted",
			},
		},
		{
			path: "doc/CONTRIBUTING.md",
			sections: []string{
				"# Contributing to gopact-examples",
				"## Development Setup",
				"## Verification",
				"## Pull Request Checklist",
			},
		},
		{
			path: "doc/SECURITY.md",
			sections: []string{
				"# Security Policy",
				"## Supported Versions",
				"## Reporting a Vulnerability",
			},
		},
		{
			path: "doc/CHANGELOG.md",
			sections: []string{
				"# Changelog",
				"## Unreleased",
			},
		},
		{
			path: "doc/maintainers/repository-governance.md",
			sections: []string{
				"# Repository Governance",
				"## Pull Request Flow",
				"## Admin Auto-Merge",
				"## Public Release Checks",
			},
		},
	} {
		body := readText(t, "../../"+doc.path)
		for _, section := range doc.sections {
			if !strings.Contains(body, section) {
				t.Fatalf("%s missing section %q", doc.path, section)
			}
		}
		if !strings.Contains(readme, doc.path) {
			t.Fatalf("README missing governance doc link %q", doc.path)
		}
	}
}

func TestExamplesPublicReadinessAndPRGovernanceAreConfigured(t *testing.T) {
	workflow := readText(t, "../../.github/workflows/ci.yml")
	readiness := readText(t, "../../scripts/public-readiness-check.sh")
	prGovernance := readText(t, "../../.github/workflows/pr-governance.yml")
	adminAutomerge := readText(t, "../../.github/workflows/admin-automerge.yml")
	governanceDoc := readText(t, "../../doc/maintainers/repository-governance.md")

	for _, want := range []string{
		"permissions:",
		"contents: read",
		"fetch-depth: 0",
		"./scripts/public-readiness-check.sh",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("CI workflow missing public readiness control %q", want)
		}
	}
	for _, want := range []string{
		"git ls-files -- .env '.env.*'",
		"git rev-list --all",
		"commit message",
		"api-key-[0-9]{14,}",
		"sk-vx[[:alnum:]_-]{20,}",
		"ep-[0-9]{14}-[[:alnum:]_-]+",
	} {
		if !strings.Contains(readiness, want) {
			t.Fatalf("public readiness script missing %q", want)
		}
	}
	for _, want := range []string{
		"name: pr-governance",
		"pull_request_target:",
		"pull_request_review:",
		"author-policy",
		"collaborators/${author}/permission",
		"== \"APPROVED\"",
	} {
		if !strings.Contains(prGovernance, want) {
			t.Fatalf("PR governance workflow missing %q", want)
		}
	}
	for _, want := range []string{
		"name: admin-automerge",
		"pull_request_target:",
		"gh pr merge",
		"--auto",
		"--squash",
		"--delete-branch",
		"!= \"admin\"",
	} {
		if !strings.Contains(adminAutomerge, want) {
			t.Fatalf("admin automerge workflow missing %q", want)
		}
	}
	for _, want := range []string{
		"author-policy",
		"Admin-authored PRs",
		"Non-admin-authored PRs",
		"Do not configure a global required review count",
		"Require status checks to pass",
	} {
		if !strings.Contains(governanceDoc, want) {
			t.Fatalf("repository governance doc missing %q", want)
		}
	}
}

func TestExamplesUsePatchedProtobuf(t *testing.T) {
	goMod := readText(t, "../../go.mod")
	if err := checkGoModModuleAtLeast(goMod, "go.mod", "google.golang.org/protobuf", "v1.33.0"); err != nil {
		t.Fatal(err)
	}
}

func TestGoModModuleVersionContractAcceptsFuturePatchedVersion(t *testing.T) {
	goMod := `module example.test

require (
	google.golang.org/protobuf v1.34.1
)
`
	if err := checkGoModModuleAtLeast(goMod, "test/go.mod", "google.golang.org/protobuf", "v1.33.0"); err != nil {
		t.Fatal(err)
	}
}

func TestGoModModuleVersionContractRejectsVulnerableRequire(t *testing.T) {
	goMod := `module example.test

require google.golang.org/protobuf v1.32.0
`
	if err := checkGoModModuleAtLeast(goMod, "test/go.mod", "google.golang.org/protobuf", "v1.33.0"); err == nil {
		t.Fatal("expected vulnerable protobuf require to fail")
	}
}

func TestGoModModuleVersionContractRejectsVulnerableReplace(t *testing.T) {
	goMod := `module example.test

require google.golang.org/protobuf v1.34.0

replace google.golang.org/protobuf => google.golang.org/protobuf v1.32.0
`
	if err := checkGoModModuleAtLeast(goMod, "test/go.mod", "google.golang.org/protobuf", "v1.33.0"); err == nil {
		t.Fatal("expected vulnerable protobuf replace to fail")
	}
}

func TestGoModModuleVersionContractRejectsDifferentReplacePath(t *testing.T) {
	goMod := `module example.test

require google.golang.org/protobuf v1.34.0

replace google.golang.org/protobuf => example.com/protobuf v1.34.0
`
	err := checkGoModModuleAtLeast(goMod, "test/go.mod", "google.golang.org/protobuf", "v1.33.0")
	if err == nil {
		t.Fatal("expected protobuf replace to a different module path to fail")
	}
	if !strings.Contains(err.Error(), "different module path") {
		t.Fatalf("error = %q, want different module path", err)
	}
}

func TestGoModModuleVersionContractIgnoresUnmatchedVersionReplace(t *testing.T) {
	goMod := `module example.test

require google.golang.org/protobuf v1.34.0

replace google.golang.org/protobuf v1.32.0 => example.com/protobuf v1.34.0
`
	if err := checkGoModModuleAtLeast(goMod, "test/go.mod", "google.golang.org/protobuf", "v1.33.0"); err != nil {
		t.Fatal(err)
	}
}

func TestExamplesCIWorkflowOptimizesIndependentGatesForParallelFeedback(t *testing.T) {
	workflow := readText(t, "../../.github/workflows/ci.yml")

	for _, want := range []string{
		"concurrency:",
		"group: ${{ github.workflow }}-${{ github.event.pull_request.number || github.ref }}",
		"cancel-in-progress: ${{ github.event_name == 'pull_request' }}",
		"hygiene:",
		"unit:",
		"race:",
		"static:",
		"coverage:",
		"security:",
		"self-bootstrap:",
		"ecosystem-self-bootstrap:",
		"needs: [hygiene, unit, race, static, coverage, security, self-bootstrap, ecosystem-self-bootstrap]",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("workflow missing parallel feedback control %q", want)
		}
	}
}

func TestQuickstartsAreDocumentedAndTested(t *testing.T) {
	entries, err := os.ReadDir("../../quickstart")
	if err != nil {
		t.Fatalf("read quickstart: %v", err)
	}

	wantCommands := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, file := range []string{"README.md", "main.go", "main_test.go"} {
			path := filepath.Join("../../quickstart", name, file)
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("quickstart/%s missing %s: %v", name, file, err)
			}
		}
		wantCommands = append(wantCommands, "go run ./quickstart/"+name)
	}
	slices.Sort(wantCommands)

	gotCommands := quickstartCommands(readText(t, "../../README.md"))
	if !slices.Equal(gotCommands, wantCommands) {
		t.Fatalf("README quickstart commands = %v, want %v", gotCommands, wantCommands)
	}
}

func TestAgnesLocalIntegrationCommandIsDocumented(t *testing.T) {
	readme := readText(t, "../../README.md")
	envExample := readText(t, "../../.env.example")
	quickstartReadme := readText(t, "../../quickstart/agnes-chat/README.md")
	script := readText(t, "../../scripts/local-agnes-integration.sh")

	command := "go test -tags=integration -count=1 ./quickstart/agnes-chat"
	if !strings.Contains(readme, command) {
		t.Fatalf("README missing Agnes integration command %q", command)
	}
	if !strings.Contains(quickstartReadme, command) {
		t.Fatalf("quickstart/agnes-chat/README.md missing Agnes integration command %q", command)
	}
	if !strings.Contains(readme, "./scripts/local-agnes-integration.sh") {
		t.Fatal("README missing local Agnes integration script")
	}
	if !strings.Contains(script, command) {
		t.Fatalf("local Agnes integration script missing command %q", command)
	}
	for _, key := range []string{"GOPACT_AGNES_API_KEY", "GOPACT_AGNES_SK", "GOPACT_LLM_TOKEN"} {
		if !strings.Contains(readme, key) {
			t.Fatalf("README missing Agnes credential variable %q", key)
		}
		if !strings.Contains(envExample, key) {
			t.Fatalf(".env.example missing Agnes credential variable %q", key)
		}
		if !strings.Contains(quickstartReadme, key) {
			t.Fatalf("quickstart/agnes-chat/README.md missing Agnes credential variable %q", key)
		}
	}
}

func TestSelfBootstrapMockSuiteIsExecutableAndUsedByCI(t *testing.T) {
	workflow := readText(t, "../../.github/workflows/ci.yml")
	readme := readText(t, "../../README.md")
	readmeZH := readText(t, "../../README_zh.md")
	scriptPath := "../../scripts/self-bootstrap-mock-suite.sh"
	script := readText(t, scriptPath)

	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("stat self-bootstrap mock suite: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatal("self-bootstrap mock suite must be executable")
	}

	command := "./scripts/self-bootstrap-mock-suite.sh"
	for _, want := range []string{
		"go test -count=1 ./quickstart/react-agent",
		"go test -count=1 ./quickstart/workflow-graph",
		"go test -count=1 ./quickstart/agent-scaffold",
		"go test -count=1 ./quickstart/generated-agent",
		"go test -count=1 ./quickstart/generated-cluster",
		"go test -count=1 ./quickstart/plan-exec",
		"go test -count=1 ./quickstart/supervisor",
		"go test -count=1 ./quickstart/agent-as-tool",
		"go test -count=1 ./quickstart/self-bootstrap",
		"go test -count=1 ./quickstart/release-bundle",
		"go test -count=1 ./quickstart/agent-node",
		"go test -count=1 ./quickstart/agent-cluster",
		"go test -count=1 ./quickstart/openai-chat",
		"go test -count=1 ./quickstart/openai-streaming",
		"go test -count=1 ./quickstart/tool-calling",
		"go test -count=1 ./quickstart/structured-output",
		"go test -count=1 ./quickstart/ark-chat",
		"go test -count=1 ./quickstart/ark-streaming",
		"go test -count=1 ./quickstart/agnes-chat",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("self-bootstrap mock suite missing %q", want)
		}
	}
	for _, forbidden := range []string{"-tags=integration", ".env"} {
		if strings.Contains(script, forbidden) {
			t.Fatalf("self-bootstrap mock suite contains %q; it must stay mock-only", forbidden)
		}
	}
	for _, want := range []string{
		"self-bootstrap:",
		command,
		"needs: [hygiene, unit, race, static, coverage, security, self-bootstrap, ecosystem-self-bootstrap]",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("workflow missing self-bootstrap suite control %q", want)
		}
	}
	for path, body := range map[string]string{
		"README.md":    readme,
		"README_zh.md": readmeZH,
	} {
		if !strings.Contains(body, command) {
			t.Fatalf("%s missing self-bootstrap mock suite command %q", path, command)
		}
	}
}

func TestEcosystemSelfBootstrapMockSmokeIsExecutableAndUsedByCI(t *testing.T) {
	workflow := readText(t, "../../.github/workflows/ci.yml")
	readme := readText(t, "../../README.md")
	readmeZH := readText(t, "../../README_zh.md")
	scriptPath := "../../scripts/ecosystem-self-bootstrap-mock-suite.sh"
	script := readText(t, scriptPath)

	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("stat ecosystem self-bootstrap mock suite: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatal("ecosystem self-bootstrap mock suite must be executable")
	}

	command := "./scripts/ecosystem-self-bootstrap-mock-suite.sh"
	for _, want := range []string{
		"GOPACT_ECOSYSTEM_ROOT",
		"GOPACT_ECOSYSTEM_FETCH",
		"gopact",
		"gopact-ext",
		"gopact-examples",
		"github.com/gopact-ai/gopact.git",
		"github.com/gopact-ai/gopact-ext.git",
		"github.com/gopact-ai/gopact-examples.git",
		"refs/tags/*:refs/tags/*",
		`run_repo_suite "gopact" "https://github.com/gopact-ai/gopact.git" "1"`,
		`run_repo_suite "gopact-ext" "https://github.com/gopact-ai/gopact-ext.git" "0"`,
		"scripts/self-bootstrap-mock-suite.sh",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("ecosystem self-bootstrap mock suite missing %q", want)
		}
	}
	for _, forbidden := range []string{"-tags=integration", ".env", "GOPACT_LLM_TOKEN", "GOPACT_AGNES", "GOPACT_ARK"} {
		if strings.Contains(script, forbidden) {
			t.Fatalf("ecosystem self-bootstrap mock suite contains %q; it must stay mock-only", forbidden)
		}
	}
	for _, want := range []string{
		"ecosystem-self-bootstrap:",
		command,
		"needs: [hygiene, unit, race, static, coverage, security, self-bootstrap, ecosystem-self-bootstrap]",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("workflow missing ecosystem self-bootstrap suite control %q", want)
		}
	}
	for path, body := range map[string]string{
		"README.md":    readme,
		"README_zh.md": readmeZH,
	} {
		if !strings.Contains(body, command) {
			t.Fatalf("%s missing ecosystem self-bootstrap mock suite command %q", path, command)
		}
	}
}

func TestAgentClusterDiscoveryEnvIsDocumented(t *testing.T) {
	for _, path := range []string{"../../README.md", "../../.env.example", "../../quickstart/agent-cluster/README.md"} {
		text := readText(t, path)
		for _, key := range []string{"GOPACT_A2A_REGISTRY_FILE", "GOPACT_A2A_REGISTRY_URL", "GOPACT_A2A_ENDPOINTS"} {
			if !strings.Contains(text, key) {
				t.Fatalf("%s missing %s", path, key)
			}
		}
	}

	for _, path := range []string{"../../README.md", "../../quickstart/agent-cluster/README.md"} {
		if !strings.Contains(readText(t, path), "Mesh.SyncEnv") {
			t.Fatalf("%s must document Mesh.SyncEnv for A2A env mesh sync", path)
		}
		if !strings.Contains(readText(t, path), "Mesh.SyncEnvEvery") {
			t.Fatalf("%s must document Mesh.SyncEnvEvery for continuous A2A env mesh sync", path)
		}
		if !strings.Contains(readText(t, path), "WithMeshHTTPAgentOptions") {
			t.Fatalf("%s must document mesh-level HTTP options for A2A env mesh sync", path)
		}
	}
}

func TestAgentClusterReadmesDocumentReleaseEvidence(t *testing.T) {
	for _, path := range []string{
		"../../README.md",
		"../../README_zh.md",
		"../../quickstart/agent-cluster/README.md",
		"../../quickstart/agent-cluster/README_zh.md",
	} {
		text := readText(t, path)
		for _, want := range []string{"replay", "command evidence"} {
			if !strings.Contains(text, want) {
				t.Fatalf("%s missing agent-cluster release evidence phrase %q", path, want)
			}
		}
	}
}

func TestExamplesUseCurrentReleasedModules(t *testing.T) {
	goMod := readText(t, "../../go.mod")
	generatedAgent := readText(t, "../../quickstart/generated-agent/main.go")
	generatedCluster := readText(t, "../../quickstart/generated-cluster/main.go")

	for _, requirement := range []string{
		"github.com/gopact-ai/gopact v0.0.50",
		"github.com/gopact-ai/gopact-ext/agents/agentnode v0.1.6",
		"github.com/gopact-ai/gopact-ext/agents/agenttool v0.1.25",
		"github.com/gopact-ai/gopact-ext/agents/planexec v0.2.26",
		"github.com/gopact-ai/gopact-ext/agents/react v0.2.24",
		"github.com/gopact-ai/gopact-ext/agents/scheduler v0.1.3",
		"github.com/gopact-ai/gopact-ext/agents/supervisor v0.1.12",
		"github.com/gopact-ai/gopact-ext/devagent/filesnapshot v0.1.23",
		"github.com/gopact-ai/gopact-ext/devagent/gitdiff v0.1.23",
		"github.com/gopact-ai/gopact-ext/devagent/selfbootstrap v0.1.4",
		"github.com/gopact-ai/gopact-ext/devagent/workspace v0.1.5",
		"github.com/gopact-ai/gopact-ext/models/agnes v0.1.27",
		"github.com/gopact-ai/gopact-ext/models/ark v0.2.24",
		"github.com/gopact-ai/gopact-ext/models/openai v0.5.26",
	} {
		if !strings.Contains(goMod, requirement) {
			t.Fatalf("go.mod missing current released module %q", requirement)
		}
	}
	if !strings.Contains(generatedAgent, `gopactVersion = "v0.0.50"`) {
		t.Fatal("quickstart/generated-agent must exercise gopact agent init at current core SDK v0.0.50")
	}
	if !strings.Contains(generatedCluster, `gopactVersion = "v0.0.50"`) {
		t.Fatal("quickstart/generated-cluster must exercise gopact agent init-cluster at current core SDK v0.0.50")
	}
}

func TestScaffoldPathUsesCredentialFreeQuickstarts(t *testing.T) {
	readme := readText(t, "../../README.md")
	for _, phrase := range []string{
		"## Scaffold Path",
		"Start without credentials:",
		"go run ./quickstart/react-agent",
		"go run ./quickstart/plan-exec",
		"go run ./quickstart/supervisor",
		"go run ./quickstart/agent-as-tool",
		"go run ./quickstart/self-bootstrap",
		"go run ./quickstart/release-bundle",
		"go run ./quickstart/agent-node",
		"go run ./quickstart/agent-cluster",
		"Use provider quickstarts after `.env` is configured.",
	} {
		if !strings.Contains(readme, phrase) {
			t.Fatalf("README missing scaffold path phrase %q", phrase)
		}
	}
}

func readText(t *testing.T, path string) string {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func quickstartCommands(readme string) []string {
	seen := map[string]struct{}{}
	var commands []string
	for _, line := range strings.Split(readme, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go run ./quickstart/") {
			if _, ok := seen[line]; ok {
				continue
			}
			seen[line] = struct{}{}
			commands = append(commands, line)
		}
	}
	slices.Sort(commands)
	return commands
}

func checkGoModModuleAtLeast(goMod, path, module, minVersion string) error {
	file, err := modfile.Parse(path, []byte(goMod), nil)
	if err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	requiredVersion := ""
	for _, requirement := range file.Require {
		if requirement.Mod.Path == module {
			requiredVersion = requirement.Mod.Version
			break
		}
	}
	if requiredVersion == "" {
		return fmt.Errorf("%s must require %s >= %s", path, module, minVersion)
	}
	if err := checkStableSemverAtLeast(requiredVersion, minVersion); err != nil {
		return fmt.Errorf("%s requires %s %s: %w", path, module, requiredVersion, err)
	}

	for _, replacement := range file.Replace {
		if replacement.Old.Path != module {
			continue
		}
		if replacement.Old.Version != "" && replacement.Old.Version != requiredVersion {
			continue
		}
		if replacement.New.Path != module {
			return fmt.Errorf("%s replaces %s with different module path %s", path, module, replacement.New.Path)
		}
		if replacement.New.Version == "" {
			return fmt.Errorf("%s replaces %s with %s without a verifiable module version", path, module, replacement.New.Path)
		}
		if err := checkStableSemverAtLeast(replacement.New.Version, minVersion); err != nil {
			return fmt.Errorf("%s replaces %s with %s %s: %w", path, module, replacement.New.Path, replacement.New.Version, err)
		}
	}
	return nil
}

func checkStableSemverAtLeast(version, minVersion string) error {
	if !semver.IsValid(version) || semver.Prerelease(version) != "" {
		return fmt.Errorf("version %q is not a stable Go semver", version)
	}
	if !semver.IsValid(minVersion) || semver.Prerelease(minVersion) != "" {
		return fmt.Errorf("minimum version %q is not a stable Go semver", minVersion)
	}
	if semver.Compare(version, minVersion) < 0 {
		return fmt.Errorf("version must be >= %s", minVersion)
	}
	return nil
}
