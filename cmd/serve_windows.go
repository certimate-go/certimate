//go:build windows
// +build windows

package cmd

import (
	"fmt"

	"ganigeorgiev.com/pocketbase/pocketbase"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

type winscHandler struct {
	pb   *pocketbase.PocketBase
	elog *eventlog.Log
}

func (h *winscHandler) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	go func() {
		if err := h.pb.Start(); err != nil {
			h.elog.Error(999, fmt.Sprintf("Start failed: %v", err))
		}
	}()

	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
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
				h.elog.Warning(998, fmt.Sprintf("unexpected control request: %v", c.Cmd))
			}
		}
	}
}

func Serve(app *pocketbase.PocketBase) error {
	if isWinsc, _ := svc.IsWindowsService(); isWinsc {
		elog, _ := eventlog.Open(winscName)
		defer elog.Close()
		return svc.Run(winscName, &winscHandler{pb: app, elog: elog})
	}

	return app.Start()
}
