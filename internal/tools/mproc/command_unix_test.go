//go:build linux || darwin || freebsd

package mproc

import (
	"bytes"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestRunWorkerCommandStopsDescendantsAfterWorkerDeath(t *testing.T) {
	cmd := exec.Command("sh", "-c", `sleep 10 & echo "$!"; kill -KILL "$$"`)
	var output bytes.Buffer
	cmd.Stdout = &output
	done := make(chan error, 1)
	go func() { done <- runWorkerCommand(cmd) }()

	select {
	case err := <-done:
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.Sys().(syscall.WaitStatus).Signal() != syscall.SIGKILL {
			t.Fatalf("worker exit = %v, want SIGKILL", err)
		}
		if pid, err := strconv.Atoi(strings.TrimSpace(output.String())); err != nil || pid <= 0 {
			t.Fatalf("missing descendant PID in worker output: %q", output.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("descendant kept the worker output pipe open after the worker exited")
	}
}

func TestRunWorkerCommandPreservesSuccessfulExit(t *testing.T) {
	cmd := exec.Command("sh", "-c", "echo complete")
	var output bytes.Buffer
	cmd.Stdout = &output
	if err := runWorkerCommand(cmd); err != nil {
		t.Fatal(err)
	}
	if output.String() != "complete\n" {
		t.Fatalf("worker output = %q", output.String())
	}
}
