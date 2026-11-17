package plugin

import (
	"log/slog"
	"time"

	githubplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type exitCause uint8

const (
	exitStopped exitCause = iota
	exitDied
)

func (c exitCause) String() string {
	switch c {
	case exitDied:
		return "died"
	default:
		return "stopped"
	}
}

// runLog writes process start/exit onto a workflow node logger.
// The zero value writes nothing (Bootstrap, access check, DNS factory).
type runLog struct {
	node *slog.Logger
}

func (l runLog) started(provider string, pid int) {
	if l.node == nil {
		return
	}
	l.node.Info("plugin process started", lifecycleAttrs(provider, pid)...)
}

func (l runLog) exited(provider string, pid int, cause exitCause) {
	if l.node == nil {
		return
	}
	args := append(lifecycleAttrs(provider, pid), slog.String("cause", cause.String()))
	if cause == exitDied {
		l.node.Warn("plugin process exited", args...)
		return
	}
	l.node.Info("plugin process exited", args...)
}

func lifecycleAttrs(provider string, pid int) []any {
	args := []any{slog.String("provider", provider)}
	if pid > 0 {
		args = append(args, slog.Int("pid", pid))
	}
	return args
}

func clientPID(client *githubplugin.Client) int {
	if client == nil {
		return 0
	}
	cfg := client.ReattachConfig()
	if cfg == nil {
		return 0
	}
	return cfg.Pid
}

func transportLost(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch st.Code() {
	case codes.Unavailable, codes.Canceled:
		return true
	default:
		return false
	}
}

// go-plugin sets Exited after stdout and stderr drain, which can lag the RPC error.
func waitClientExited(client *githubplugin.Client) {
	if client == nil || client.Exited() {
		return
	}
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
		if client.Exited() {
			return
		}
	}
}
