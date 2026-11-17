package mproc

import "golang.org/x/sys/unix"

func waitForWorkerExit(pid int) error {
	for {
		err := unix.Waitid(unix.P_PID, pid, &unix.Siginfo{}, unix.WEXITED|unix.WNOWAIT, nil)
		if err != unix.EINTR {
			return err
		}
	}
}
