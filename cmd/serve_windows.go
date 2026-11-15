//go:build windows
// +build windows

package cmd

import (
	"fmt"

	"ganigeorgiev.com/pocketbase/pocketbase"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/hook"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

type winscHandler struct {
	pb   *pocketbase.PocketBase
	elog *eventlog.Log
}

func (h *winscHandler) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	s <- svc.Status{State: svc.StartPending}
	ready := make(chan struct{})
	finished := make(chan error, 1)
	h.pb.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Priority: -1000,
		Func: func(e *core.ServeEvent) error {
			if err := e.Next(); err != nil {
				return err
			}
			close(ready)
			return nil
		},
	})
	go func() {
		finished <- h.pb.Start()
	}()

	for {
		select {
		case <-ready:
			ready = nil
			s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
		case err := <-finished:
			if err != nil {
				if h.elog != nil {
					h.elog.Error(999, fmt.Sprintf("Start failed: %v", err))
				}
				s <- svc.Status{State: svc.Stopped}
				return true, 1
			}
			s <- svc.Status{State: svc.Stopped}
			return false, 0
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				s <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				event := new(core.TerminateEvent)
				event.App = h.pb
				h.pb.OnTerminate().Trigger(event, func(e *core.TerminateEvent) error {
					return e.App.ResetBootstrapState()
				})
				s <- svc.Status{State: svc.Stopped}
				return false, 0
			default:
				if h.elog != nil {
					h.elog.Warning(998, fmt.Sprintf("unexpected control request: %v", c.Cmd))
				}
			}
		}
	}
}

func Serve(app *pocketbase.PocketBase) error {
	if isWinsc, _ := svc.IsWindowsService(); isWinsc {
		elog, err := eventlog.Open(winscName)
		if err != nil {
			return err
		}
		defer elog.Close()
		return svc.Run(winscName, &winscHandler{pb: app, elog: elog})
	}

	return app.Start()
}
