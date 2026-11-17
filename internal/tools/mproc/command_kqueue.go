//go:build darwin || freebsd

package mproc

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func waitForWorkerExit(pid int) error {
	syscall.ForkLock.RLock()
	queue, err := unix.Kqueue()
	if err == nil {
		unix.CloseOnExec(queue)
	}
	syscall.ForkLock.RUnlock()
	if err != nil {
		return err
	}
	defer unix.Close(queue)

	change := make([]unix.Kevent_t, 1)
	unix.SetKevent(&change[0], pid, unix.EVFILT_PROC, unix.EV_ADD|unix.EV_ONESHOT)
	change[0].Fflags = unix.NOTE_EXIT
	event := make([]unix.Kevent_t, 1)
	for {
		n, err := unix.Kevent(queue, change, event, nil)
		if err == unix.EINTR {
			continue
		}
		if err == unix.ESRCH {
			return nil
		}
		if err != nil {
			return err
		}
		if n > 0 && event[0].Flags&unix.EV_ERROR != 0 {
			eventErr := unix.Errno(event[0].Data)
			if eventErr != unix.ESRCH {
				return eventErr
			}
		}
		return nil
	}
}
