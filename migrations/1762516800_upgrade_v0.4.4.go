package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		if err := app.DB().
			NewQuery("SELECT (1) FROM _migrations WHERE file={:file} LIMIT 1").
			Bind(dbx.Params{"file": "1762516800_m0.4.4.go"}).
			One(&struct{}{}); err == nil {
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
			collection, err := app.FindCollectionByNameOrId("dy6ccjb60spfy6p")
			if err != nil {
				return err
			}

			records, err := app.FindRecordsByFilter(collection, "name=\"persistence\"", "", 1, 0)
			if err != nil {
				return err
			} else if len(records) != 0 {
				record := records[0]
				changed := false

				content := make(map[string]any)
				if err := record.UnmarshalJSONField("content", &content); err != nil {
					return err
				} else {
					if _, ok := content["expiredCertificatesMaxDaysRetention"]; ok {
						content["certificatesRetentionMaxDays"] = content["expiredCertificatesMaxDaysRetention"]
						delete(content, "expiredCertificatesMaxDaysRetention")

						record.Set("content", content)
						changed = true
					}

					if _, ok := content["workflowRunsMaxDaysRetention"]; ok {
						content["workflowRunsRetentionMaxDays"] = content["workflowRunsMaxDaysRetention"]
						delete(content, "workflowRunsMaxDaysRetention")

						record.Set("content", content)
						changed = true
					}
				}

				if changed {
					if err := app.Save(record); err != nil {
						return err
					}

					tracer.Printf("record #%s in collection '%s' updated", record.Id, collection.Name)
				}
			}
		}

		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		return nil
	})
}
