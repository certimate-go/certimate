package app

import (
	"log/slog"
	"sync"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
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
	app := GetApp()
	if !app.IsBootstrapped() {
		panic("MUST NOT USE THIS BEFORE APP BOOTSTRAPPED!")
	}

	return app.Logger()
}
