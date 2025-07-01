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
