//go:build linux || darwin || freebsd

package mproc

import (
	"errors"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

func runWorkerCommand(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return err
	}

	exitErr := waitForWorkerExit(cmd.Process.Pid)
	if errors.Is(exitErr, unix.ECHILD) {
		return errors.Join(exitErr, cmd.Wait())
	}
	groupErr := unix.Kill(-cmd.Process.Pid, unix.SIGKILL)
	if errors.Is(groupErr, unix.ESRCH) || errors.Is(groupErr, unix.EPERM) {
		groupErr = nil
	}
	return errors.Join(exitErr, groupErr, cmd.Wait())
}
