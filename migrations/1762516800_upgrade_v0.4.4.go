package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		if mr, _ := app.FindFirstRecordByFilter("_migrations", "file='1762516800_app_v0.4.4.go'"); mr != nil {
			return nil
		}

		tracer := NewTracer("v0.4.4")
		tracer.Printf("go ...")

		// update collection `access`
		//   - fix #1027
		{
			if _, err := app.DB().NewQuery("UPDATE access SET provider = 'hostingde' WHERE provider = 'hostingDE'").Execute(); err != nil {
				return err
			}
		}

		// update collection `workflow`
		//   - fix #1027
		{
			if _, err := app.DB().NewQuery("UPDATE workflow SET graphDraft = REPLACE(graphDraft, '\"hostingDE\"', '\"hostingde\"')").Execute(); err != nil {
				return err
			}

			if _, err := app.DB().NewQuery("UPDATE workflow SET graphContent = REPLACE(graphContent, '\"hostingDE\"', '\"hostingde\"')").Execute(); err != nil {
				return err
			}
		}

		// update collection `settings`
		//   - modify field `content` schema of `persistence`
		{
			if _, err := app.DB().NewQuery("UPDATE settings SET content = REPLACE(content, '\"expiredCertificatesMaxDaysRetention\"', '\"certificatesRetentionMaxDays\"') WHERE name = 'persistence'").Execute(); err != nil {
				return err
			}

			if _, err := app.DB().NewQuery("UPDATE settings SET content = REPLACE(content, '\"workflowRunsMaxDaysRetention\"', '\"workflowRunsRetentionMaxDays\"') WHERE name = 'persistence'").Execute(); err != nil {
				return err
			}
		}

		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		return nil
	})
}
