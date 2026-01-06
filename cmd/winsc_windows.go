//go:build windows
// +build windows

package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/certimate-go/certimate/internal/app"
)

const winscName = "certimate"

func NewWinscCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:   "winsc",
		Short: "Install/Uninstall Windows service",
	}

	command.AddCommand(winscInstallCommand(app))
	command.AddCommand(winscUninstallCommand(app))
	command.AddCommand(winscStartCommand(app))
	command.AddCommand(winscStopCommand(app))

	return command
}

func winscInstallCommand(_ core.App) *cobra.Command {
	command := &cobra.Command{
		Use:     "install [args...]",
		Example: "winsc install",
		Run: func(cmd *cobra.Command, args []string) {
			srvPath, err := os.Executable()
			if err != nil {
				srvPath = os.Args[0]
			}

			srvArgs := []string{"serve"}
			srvArgs = append(srvArgs, args...)

			manager, err := mgr.Connect()
			if err != nil {
				slog.Error(fmt.Sprintf("failed to connect to service manager: %v", err))
				return
			}
			defer manager.Disconnect()

			config := mgr.Config{
				DisplayName: app.AppName,
				Description: "https://github.com/certimate-go/certimate",
				StartType:   mgr.StartAutomatic,
			}
			service, err := manager.CreateService(winscName, srvPath, config, srvArgs...)
			if err != nil {
				slog.Error(fmt.Sprintf("failed to create service: %v", err))
				return
			}
			defer service.Close()

			eventlog.InstallAsEventCreate(winscName, eventlog.Error|eventlog.Warning|eventlog.Info)
			slog.Info(fmt.Sprintf("service '%s' installed", winscName))

			if err := service.Start(); err != nil {
				slog.Warn(fmt.Sprintf("failed to start service: %v", err))
			}

			slog.Info(fmt.Sprintf("service '%s' started", winscName))
		},
		DisableFlagParsing: true,
	}

	return command
}

func winscUninstallCommand(_ core.App) *cobra.Command {
	command := &cobra.Command{
		Use:     "uninstall",
		Example: "winsc uninstall",
		Run: func(cmd *cobra.Command, args []string) {
			manager, err := mgr.Connect()
			if err != nil {
				slog.Error(fmt.Sprintf("failed to connect to service manager: %v", err))
				return
			}
			defer manager.Disconnect()

			service, err := manager.OpenService(winscName)
			if err != nil {
				slog.Error(fmt.Sprintf("failed to open service: %v", err))
				return
			}
			defer service.Close()

			status, err := service.Query()
			if err == nil && status.State != svc.Stopped {
				_, err = service.Control(svc.Stop)
				if err != nil {
					slog.Warn(fmt.Sprintf("failed to stop service: %v", err))
				}

				time.Sleep(3 * time.Second)
				slog.Info(fmt.Sprintf("service '%s' stopped", winscName))
			}

			if err = service.Delete(); err != nil {
				slog.Error(fmt.Sprintf("failed to delete service: %v", err))
				return
			}

			eventlog.Remove(winscName)
			slog.Info(fmt.Sprintf("service '%s' uninstalled", winscName))
		},
	}

	return command
}

func winscStartCommand(_ core.App) *cobra.Command {
	command := &cobra.Command{
		Use:     "start",
		Example: "winsc start",
		Run: func(cmd *cobra.Command, args []string) {
			manager, err := mgr.Connect()
			if err != nil {
				slog.Error(fmt.Sprintf("failed to connect to service manager: %v", err))
				return
			}
			defer manager.Disconnect()

			service, err := manager.OpenService(winscName)
			if err != nil {
				slog.Error(fmt.Sprintf("failed to open service: %v", err))
				return
			}
			defer service.Close()

			if err := service.Start(); err != nil {
				slog.Error(fmt.Sprintf("failed to start service: %v", err))
				return
			}

			slog.Info(fmt.Sprintf("service '%s' started", winscName))
		},
	}

	return command
}

func winscStopCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:     "stop",
		Example: "winsc stop",
		Run: func(cmd *cobra.Command, args []string) {
			manager, err := mgr.Connect()
			if err != nil {
				slog.Error(fmt.Sprintf("failed to connect to service manager: %v", err))
				return
			}
			defer manager.Disconnect()

			service, err := manager.OpenService(winscName)
			if err != nil {
				slog.Error(fmt.Sprintf("failed to open service: %v", err))
				return
			}
			defer service.Close()

			status, err := service.Query()
			if err == nil && status.State != svc.Stopped {
				_, err = service.Control(svc.Stop)
				if err != nil {
					slog.Warn(fmt.Sprintf("failed to stop service: %v", err))
				}

				time.Sleep(3 * time.Second)
				slog.Info(fmt.Sprintf("service '%s' stopped", winscName))
			}
		},
	}

	return command
}
