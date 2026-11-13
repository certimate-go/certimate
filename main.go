package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	_ "time/tzdata"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/hook"

	"github.com/certimate-go/certimate/cmd"
	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/pluginhost"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/internal/rest/routes"
	"github.com/certimate-go/certimate/internal/scheduler"
	"github.com/certimate-go/certimate/internal/settings"
	"github.com/certimate-go/certimate/internal/tools/mproc"
	"github.com/certimate-go/certimate/internal/workflow"
	"github.com/certimate-go/certimate/pkg/plugin"
	"github.com/certimate-go/certimate/ui"

	_ "github.com/certimate-go/certimate/internal/accessschema"
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

	isServeCmd := slices.Contains(os.Args[1:], "serve")
	isInterCmd := len(os.Args) > 1 && os.Args[1] == "intercmd"
	if isInterCmd && len(os.Args) > 2 && os.Args[2] == "certapply" {
		if err := mproc.BindWorkerChildLifetime(); err != nil {
			slog.Error("[CERTIMATE] Failed to bind certificate worker child processes.", slog.Any("error", err))
			os.Exit(1)
		}
	}
	stopPlugins := func() {}

	if isServeCmd || isInterCmd {
		pb.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
			if err := e.Next(); err != nil {
				return err
			}

			if isServeCmd {
				settings.Setup()
				stopPlugins = scanPlugins()
			} else {
				childLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(slog.String("component", "pluginhost"))
				stopPlugins = registerPlugins(childLogger)
			}
			return nil
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
			scheduler.Setup()
			workflow.Setup()
			routes.BindRouter(e.Router)

			if err := e.Next(); err != nil {
				return err
			}

			slog.Info("[CERTIMATE] Serving on " + e.Server.Addr)
			return nil
		})

		pb.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
			if pb.IsBootstrapped() {
				workflow.Teardown()
				stopPlugins()
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

func registerPlugins(logger *slog.Logger) func() {
	pluginDir := resolvePluginDir()
	if pluginDir == "" {
		return func() {}
	}
	cfg := plugin.PluginConfig{
		PluginDir:   pluginDir,
		CoreVersion: app.AppVersion,
	}

	catalog, errs := pluginhost.ScanAndRegister(context.Background(), cfg, logger)
	pluginhost.SetGlobalCatalog(catalog)
	for _, err := range errs {
		logger.Warn("[CERTIMATE] plugin registration error", slog.Any("error", err))
	}
	return func() {}
}

func scanPlugins() func() {
	pluginDir := resolvePluginDir()
	if pluginDir == "" {
		return func() {}
	}
	logger := slog.Default().With(slog.String("component", "pluginhost"))

	pluginhost.SweepOrphans(pluginDir, logger)

	stop := registerPlugins(logger)

	cfg := plugin.PluginConfig{
		PluginDir:   pluginDir,
		CoreVersion: app.AppVersion,
	}
	reloader := pluginhost.NewReloader(cfg, pluginhost.GlobalCatalog(), logger)
	pluginhost.SetGlobalReloader(reloader)

	marketSvc := pluginhost.NewMarketService(pluginhost.MarketConfig{
		Cache:          repository.NewPluginMarketCache(app.GetApp()),
		Sources:        repository.NewPluginMarketSources(app.GetApp()),
		PluginDir:      pluginDir,
		DownloadMirror: os.Getenv("CERTIMATE_PLUGIN_DOWNLOAD_MIRROR"),
		Logger:         logger,
	})
	pluginhost.SetGlobalMarketService(marketSvc)

	watcher := pluginhost.NewWatcher(pluginDir, logger)
	ctx, cancel := context.WithCancel(context.Background())
	watcher.Start(ctx)
	reloader.Start(ctx, watcher)
	return func() { stop(); cancel(); reloader.Close() }
}

func resolvePluginDir() string {
	if dir := strings.TrimSpace(os.Getenv("CERTIMATE_PLUGIN_DIR")); dir != "" {
		return dir
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(cwd, "plugins")
}
