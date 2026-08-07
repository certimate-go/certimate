package settings

import (
	"github.com/certimate-go/certimate/internal/app"
)

func initPbSettings() {
	pb := app.GetApp()

	settings := pb.Settings()
	changed := false

	if settings.Meta.AppName != app.AppName {
		settings.Meta.AppName = app.AppName
		changed = true
	}

	if settings.Batch.Enabled != true {
		settings.Batch.Enabled = true
		settings.Batch.MaxRequests = 1000
		settings.Batch.Timeout = 30
		changed = true
	}

	if changed {
		if err := pb.Save(settings); err != nil {
			panic(err)
		}
	}
}
