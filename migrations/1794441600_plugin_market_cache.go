package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		_, err := app.DB().NewQuery(`CREATE TABLE plugin_market_cache (
			source TEXT NOT NULL,
			provider_type TEXT NOT NULL,
			payload TEXT NOT NULL,
			PRIMARY KEY (source, provider_type)
		) WITHOUT ROWID`).Execute()
		return err
	}, func(app core.App) error {
		_, err := app.DB().NewQuery("DROP TABLE plugin_market_cache").Execute()
		return err
	})
}
