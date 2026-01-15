package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("v0.4.14")
		tracer.Printf("go ...")

		// update collection `settings`
		//   - modify field `content` schema of `sslProvider`
		{
			collection, err := app.FindCollectionByNameOrId("dy6ccjb60spfy6p")
			if err != nil {
				return err
			}

			records, err := app.FindRecordsByFilter(collection, "name=\"sslProvider\"", "", 1, 0)
			if err != nil {
				return err
			} else if len(records) != 0 {
				record := records[0]
				changed := false

				content := make(map[string]any)
				if err := record.UnmarshalJSONField("content", &content); err != nil {
					return err
				} else {
					if _, ok := content["config"]; ok {
						content["configs"] = content["config"]
						delete(content, "config")

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
