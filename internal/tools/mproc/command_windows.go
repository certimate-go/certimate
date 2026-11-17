package mproc

import "os/exec"

func runWorkerCommand(cmd *exec.Cmd) error {
	return cmd.Run()
}
