package main

import (
	"log/slog"
	"os"
	"strings"
	_ "time/tzdata"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/hook"
	"github.com/spf13/pflag"

	"github.com/certimate-go/certimate/cmd"
	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/rest/routes"
	"github.com/certimate-go/certimate/internal/scheduler"
	"github.com/certimate-go/certimate/internal/workflow"
	"github.com/certimate-go/certimate/ui"

	_ "github.com/certimate-go/certimate/migrations"
)

func main() {
	pb := app.GetApp().(*pocketbase.PocketBase)
	if len(os.Args) < 2 {
		slog.Error("[CERTIMATE] missing exec args, maybe you forgot the 'serve' command?")
		os.Exit(1)
		return
	}

	migratecmd.MustRegister(pb, pb.RootCmd, migratecmd.Config{
		Automigrate: strings.HasPrefix(os.Args[0], os.TempDir()),
	})

	pb.RootCmd.AddCommand(cmd.NewInternalCommand(pb))
	pb.RootCmd.AddCommand(cmd.NewVersionCommand(pb))
	pb.RootCmd.AddCommand(cmd.NewWinscCommand(pb))

	isServeCmd := os.Args[1] == "serve"

	if isServeCmd {
		var flagHttp string
		pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
		pflag.CommandLine.Parse(os.Args[2:]) // skip the first two arguments: "main.go serve"
		pflag.StringVar(&flagHttp, "http", "127.0.0.1:8090", "HTTP server address")
		pflag.Parse()

		pb.OnServe().BindFunc(func(e *core.ServeEvent) error {
			scheduler.Setup()
			workflow.Setup()
			routes.BindRouter(e.Router)
			return e.Next()
		})

		pb.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
			Func: func(e *core.ServeEvent) error {
				e.Router.
					GET("/{path...}", apis.Static(ui.DistDirFS, false)).
					Bind(apis.Gzip())
				return e.Next()
			},
			Priority: 999,
		})

		pb.OnServe().BindFunc(func(e *core.ServeEvent) error {
			slog.Info("[CERTIMATE] Visit the website: http://" + flagHttp)
			return e.Next()
		})

		pb.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
			err := e.Next()
			if err != nil {
				return err
			}

			settings := pb.Settings()
			if !settings.Batch.Enabled {
				settings.Batch.Enabled = true
				settings.Batch.MaxRequests = 1000
				settings.Batch.Timeout = 30
				if err := pb.Save(settings); err != nil {
					return err
				}
			}

			return nil
		})

		pb.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
			if pb.IsBootstrapped() {
				workflow.Teardown()
			}

			return e.Next()
		})
	}

	if err := cmd.Serve(pb); err != nil {
		if isServeCmd {
			slog.Error("[CERTIMATE] Serve failed.", slog.Any("error", err))
		} else {
			slog.Error("[CERTIMATE] Start failed.", slog.Any("error", err))
		}
	}
}
