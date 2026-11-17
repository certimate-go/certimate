package migrations

import (
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	m "ganigeorgiev.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		_, err := app.DB().NewQuery(`CREATE TABLE plugin_market_sources (
   id TEXT PRIMARY KEY NOT NULL,
   repository TEXT UNIQUE NOT NULL
  ) WITHOUT ROWID`).Execute()
		return err
	}, func(app core.App) error {
		_, err := app.DB().NewQuery("DROP TABLE plugin_market_sources").Execute()
		return err
	})
}
