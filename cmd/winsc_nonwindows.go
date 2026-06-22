//go:build !windows
// +build !windows

package cmd

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func NewWinscCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:   "winsc",
		Short: "Manages Windows service (Not supported on non-Windows OS)",
	}

	return command
}
