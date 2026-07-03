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

type registryAgent struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

const (
	gopactVersion = "v0.0.49"
	clusterName   = "generated-cluster"
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
	dir, err := os.MkdirTemp("", "gopact-generated-cluster-*")
	if err != nil {
		return err
	}
	target := filepath.Join(dir, clusterName)

	if err := runCommand(ctx, "", "go", "run", "github.com/gopact-ai/gopact/cmd/gopact@"+gopactVersion,
		"agent", "init-cluster", clusterName,
		"-out", target,
	); err != nil {
		return err
	}
	if err := runCommand(ctx, target, "go", "mod", "tidy"); err != nil {
		return err
	}
	if err := runCommand(ctx, "", "go", "run", "github.com/gopact-ai/gopact/cmd/gopact@"+gopactVersion, "agent", "verify", target); err != nil {
		return err
	}
	url, err := runGeneratedClusterSmoke(ctx, target)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(out, "generated %s at %s\n", clusterName, target); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "verified %s with gopact agent verify\n", clusterName); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "served %s at %s\n", clusterName, url)
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

func runGeneratedClusterSmoke(ctx context.Context, target string) (string, error) {
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
		"GOPACT_CLUSTER_ADDR="+addr,
		"GOPACT_CLUSTER_URL="+url,
		"GOPACT_A2A_REGISTRY_URL="+url+"/agents.json",
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

	if err := waitForGeneratedCluster(smokeCtx, url); err != nil {
		cancel()
		stopGeneratedCluster(cmd, waitErr)
		return "", fmt.Errorf("generated cluster smoke: %w\n%s", err, output.String())
	}
	cancel()
	stopGeneratedCluster(cmd, waitErr)
	return url, nil
}

func stopGeneratedCluster(cmd *exec.Cmd, waitErr <-chan error) {
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

func waitForGeneratedCluster(ctx context.Context, url string) error {
	client := http.Client{Timeout: 2 * time.Second}
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := checkGeneratedCluster(client, url)
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

func checkGeneratedCluster(client http.Client, url string) error {
	for _, path := range []string{"/readyz", "/agents/planner-agent/readyz"} {
		resp, err := client.Get(url + path)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s status %d", path, resp.StatusCode)
		}
	}

	resp, err := client.Get(url + "/agents.json")
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agents.json status %d", resp.StatusCode)
	}
	agents, err := decodeRegistryAgents(resp.Body)
	if err != nil {
		return err
	}
	wantNames := []string{"planner-agent", "worker-agent", "reviewer-agent"}
	if len(agents) != len(wantNames) {
		return fmt.Errorf("agents.json agents = %+v", agents)
	}
	for i, want := range wantNames {
		if agents[i].Name != want || agents[i].URL != url+"/agents/"+want {
			return fmt.Errorf("agents.json agents = %+v", agents)
		}
	}
	return nil
}

func decodeRegistryAgents(r io.Reader) ([]registryAgent, error) {
	var raw json.RawMessage
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, err
	}
	var agents []registryAgent
	if err := json.Unmarshal(raw, &agents); err == nil {
		return agents, nil
	}
	var wrapped struct {
		Agents []registryAgent `json:"agents"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}
	if wrapped.Agents == nil {
		return nil, fmt.Errorf("agents.json missing agents")
	}
	return wrapped.Agents, nil
}
