package cmd

import (
	"fmt"
	"runtime"

	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/certimate-go/certimate/internal/app"
)

func NewVersionCommand(_ core.App) *cobra.Command {
	command := &cobra.Command{
		Use:   "version",
		Short: "Prints version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Certimate v%s\n", app.AppVersion)
			fmt.Printf("Build with %s on %s_%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
		},
	}

	return command
}
