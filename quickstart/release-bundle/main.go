package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact/gopacttest"
)

type releaseBundle struct {
	RunExport gopact.RunExport          `json:"run_export"`
	Report    gopact.VerificationReport `json:"report"`
}

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	dir, err := os.MkdirTemp("", "gopact-release-bundle-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	export := gopact.RunExport{
		Version: gopact.RunExportVersion,
		IDs:     gopact.RuntimeIDs{RunID: "release-bundle-quickstart", ThreadID: "release-bundle-quickstart"},
		Outcome: gopact.RunCompleted,
	}
	report, err := gopacttest.BuildSelfBootstrapReleaseGateReport(export)
	if err != nil {
		return err
	}

	exportPath := filepath.Join(dir, "run-export.json")
	reportPath := filepath.Join(dir, "verification-report.json")
	if err := writeJSONFile(exportPath, export); err != nil {
		return err
	}
	if err := writeJSONFile(reportPath, report); err != nil {
		return err
	}

	bundle, err := runReleaseBundleCLI(ctx, exportPath, reportPath)
	if err != nil {
		return err
	}
	for _, result := range gopacttest.CheckSelfBootstrapReleaseGate(ctx, bundle.RunExport, bundle.Report) {
		if !result.Passed {
			return fmt.Errorf("release-bundle gate %s: %w", result.Case, result.Err)
		}
	}

	if _, err := fmt.Fprintln(out, "release-bundle: core CLI"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "report: %s\n", bundle.Report.Status); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "bundle: %s verification_reports=%d\n", bundle.RunExport.Outcome, len(bundle.RunExport.VerificationReports)); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, "gate: passed")
	return err
}

func runReleaseBundleCLI(ctx context.Context, exportPath, reportPath string) (releaseBundle, error) {
	cmd := exec.CommandContext(ctx,
		"go", "run", "github.com/gopact-ai/gopact/cmd/gopact",
		"release-bundle",
		"-run-export", exportPath,
		"-report", reportPath,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return releaseBundle{}, fmt.Errorf("gopact release-bundle: %w: %s", err, stderr.String())
	}

	var bundle releaseBundle
	if err := json.Unmarshal(stdout.Bytes(), &bundle); err != nil {
		return releaseBundle{}, fmt.Errorf("decode release bundle: %w", err)
	}
	return bundle, nil
}

func writeJSONFile(path string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}
