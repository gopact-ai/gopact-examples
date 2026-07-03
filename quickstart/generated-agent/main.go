package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	gopactVersion = "v0.0.44"
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
	); err != nil {
		return err
	}
	if err := runCommand(ctx, target, "go", "mod", "tidy"); err != nil {
		return err
	}
	if err := runCommand(ctx, target, "go", "test", "./..."); err != nil {
		return err
	}
	url, err := runGeneratedAgentSmoke(ctx, target)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(out, "generated %s at %s\n", agentName, target); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "served %s at %s\n", agentName, url)
	return err
}

func runCommand(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GOPRIVATE=github.com/gopact-ai/*",
		"GONOSUMDB=github.com/gopact-ai/*",
		"GONOPROXY=github.com/gopact-ai/*",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, output)
	}
	return nil
}

func runGeneratedAgentSmoke(ctx context.Context, target string) (string, error) {
	addr, err := freeLocalAddr()
	if err != nil {
		return "", err
	}
	url := "http://" + addr
	smokeCtx, stop := context.WithTimeout(ctx, 45*time.Second)
	defer stop()
	runCtx, cancel := context.WithCancel(smokeCtx)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "go", "run", "github.com/gopact-ai/gopact/cmd/gopact@"+gopactVersion, "agent", "run", target)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = append(os.Environ(),
		"GOPRIVATE=github.com/gopact-ai/*",
		"GONOSUMDB=github.com/gopact-ai/*",
		"GONOPROXY=github.com/gopact-ai/*",
		"GOPACT_AGENT_ADDR="+addr,
		"GOPACT_AGENT_URL="+url,
	)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		return "", err
	}
	waitErr := make(chan error, 1)
	go func() {
		waitErr <- cmd.Wait()
	}()

	if err := waitForGeneratedAgent(smokeCtx, url); err != nil {
		cancel()
		stopGeneratedAgent(cmd, waitErr)
		return "", fmt.Errorf("generated agent smoke: %w\n%s", err, output.String())
	}
	cancel()
	stopGeneratedAgent(cmd, waitErr)
	return url, nil
}

func stopGeneratedAgent(cmd *exec.Cmd, waitErr <-chan error) {
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	<-waitErr
}

func freeLocalAddr() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = listener.Close()
	}()
	return listener.Addr().String(), nil
}

func waitForGeneratedAgent(ctx context.Context, url string) error {
	client := http.Client{Timeout: 2 * time.Second}
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := checkGeneratedAgent(client, url)
		if err == nil {
			return nil
		}
		lastErr = err
		timer := time.NewTimer(200 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("%w: %w", ctx.Err(), lastErr)
		case <-timer.C:
		}
	}
}

func checkGeneratedAgent(client http.Client, url string) error {
	resp, err := client.Get(url + "/readyz")
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("readyz status %d", resp.StatusCode)
	}

	resp, err = client.Get(url + "/agents.json")
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agents.json status %d", resp.StatusCode)
	}
	var registry struct {
		Agents []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"agents"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
		return err
	}
	if len(registry.Agents) != 1 || registry.Agents[0].Name != agentName || registry.Agents[0].URL != url {
		return fmt.Errorf("agents.json agents = %+v", registry.Agents)
	}
	return nil
}
