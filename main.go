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
	app := app.GetApp().(*pocketbase.PocketBase)
	if len(os.Args) < 2 {
		slog.Error("[CERTIMATE] missing exec args, maybe you forget the 'serve' command?")
		os.Exit(1)
		return
	}

	var flagHttp string
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	pflag.CommandLine.Parse(os.Args[2:]) // skip the first two arguments: "main.go serve"
	pflag.StringVar(&flagHttp, "http", "127.0.0.1:8090", "HTTP server address")
	pflag.Parse()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Admin UI
		// (the isGoRun check is to enable it only during development)
		Automigrate: strings.HasPrefix(os.Args[0], os.TempDir()),
	})

	app.RootCmd.AddCommand(cmd.NewInternalCommand(app))
	app.RootCmd.AddCommand(cmd.NewWinscCommand(app))

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		scheduler.Setup()
		workflow.Setup()
		routes.SetupRouter(e.Router)
		return e.Next()
	})

	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Func: func(e *core.ServeEvent) error {
			e.Router.
				GET("/{path...}", apis.Static(ui.DistDirFS, false)).
				Bind(apis.Gzip())
			return e.Next()
		},
		Priority: 999,
	})

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		slog.Info("[CERTIMATE] Visit the website: http://" + flagHttp)
		return e.Next()
	})

	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		workflow.Teardown()
		return e.Next()
	})

	if err := cmd.Serve(app); err != nil {
		slog.Error("[CERTIMATE] Start failed.", slog.Any("error", err))
	}
}
