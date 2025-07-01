package main

import (
	"flag"
	"log/slog"
	"os"
	"strings"
	_ "time/tzdata"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/hook"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/rest/routes"
	"github.com/certimate-go/certimate/internal/scheduler"
	"github.com/certimate-go/certimate/internal/workflow"
	"github.com/certimate-go/certimate/ui"

	_ "github.com/certimate-go/certimate/migrations"
)

func main() {
	app := app.GetApp().(*pocketbase.PocketBase)

	var flagHttp string
	flag.StringVar(&flagHttp, "http", "127.0.0.1:8090", "HTTP server address")
	if len(os.Args) < 2 {
		slog.Error("[CERTIMATE] missing exec args")
		os.Exit(1)
		return
	}
	_ = flag.CommandLine.Parse(os.Args[2:]) // skip the first two arguments: "main.go serve"

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Admin UI
		// (the isGoRun check is to enable it only during development)
		Automigrate: strings.HasPrefix(os.Args[0], os.TempDir()),
	})

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		scheduler.Register()
		workflow.Register()
		routes.Register(e.Router)
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
		routes.Unregister()
		return e.Next()
	})

	if err := app.Start(); err != nil {
		slog.Error("[CERTIMATE] Start failed.", "err", err)
	}
}
