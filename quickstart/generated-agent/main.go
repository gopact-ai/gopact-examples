package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	gopactVersion = "v0.0.19"
	agentName     = "generated-agent"
	modulePath    = "example.com/generated-agent"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := run(ctx, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	dir, err := os.MkdirTemp("", "gopact-generated-agent-*")
	if err != nil {
		return err
	}
	target := filepath.Join(dir, agentName)

	if err := runCommand(ctx, "", "go", "run", "github.com/gopact-ai/gopact/cmd/gopact@"+gopactVersion,
		"agent", "init", agentName,
		"-out", target,
		"-module", modulePath,
		"-sdk-version", gopactVersion,
	); err != nil {
		return err
	}
	if err := runCommand(ctx, target, "go", "mod", "tidy"); err != nil {
		return err
	}
	if err := runCommand(ctx, target, "go", "test", "./..."); err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "generated %s at %s\n", agentName, target)
	return err
}

func runCommand(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GOPRIVATE=github.com/gopact-ai/*",
		"GONOSUMDB=github.com/gopact-ai/*",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, output)
	}
	return nil
}
