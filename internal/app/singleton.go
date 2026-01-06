package app

import (
	"log/slog"
	"sync"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

var (
	instance    core.App
	intanceOnce sync.Once
)

func GetApp() core.App {
	intanceOnce.Do(func() {
		pb := pocketbase.NewWithConfig(pocketbase.Config{
			HideStartBanner: true,
		})

		pb.RootCmd.Flags().MarkHidden("encryptionEnv")
		pb.RootCmd.Flags().MarkHidden("queryTimeout")

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

		instance = pb
	})

	return instance
}

func GetDB() dbx.Builder {
	return GetApp().DB()
}

func GetLogger() *slog.Logger {
	return GetApp().Logger()
}
