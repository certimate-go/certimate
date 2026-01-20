package migrations

import (
	"net"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"

	snaps "github.com/certimate-go/certimate/migrations/snaps/v0.4"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("v0.4.11")
		tracer.Printf("go ...")

		// adapt to new workflow data structure
		{
			walker := &snaps.WorkflowGraphWalker{}
			walker.Define(func(node *snaps.WorkflowNode) (_changed bool, _err error) {
				_changed = false
				_err = nil

				if node.Type != "bizApply" {
					return
				}

				nodeCfg := node.Data.Config

				if nodeCfg["identifier"] == nil || nodeCfg["identifier"] == "" {
					if nodeCfg["domains"] != nil && nodeCfg["domains"].(string) != "" {
						if ip := net.ParseIP(nodeCfg["domains"].(string)); ip != nil {
							nodeCfg["identifier"] = "ip"
						} else {
							nodeCfg["identifier"] = "domain"
						}

						_changed = true
						return
					}
				}

				return
			})
			walker.Define(func(node *snaps.WorkflowNode) (_changed bool, _err error) {
				_changed = false
				_err = nil

				if node.Type != "bizUpload" {
					return
				}

				nodeCfg := node.Data.Config

				if nodeCfg["domains"] != nil {
					delete(nodeCfg, "domains")

					_changed = true
					return
				}

				return
			})

			// update collection `workflow`
			//   - migrate field `graphDraft` / `graphContent`
			{
				collection, err := app.FindCollectionByNameOrId("tovyif5ax6j62ur")
				if err != nil {
					return err
				}

				records, err := app.FindAllRecords(collection)
				if err != nil {
					return err
				}

				for _, record := range records {
					changed := false

					if ret, err := walker.Migrate(record, "graphDraft"); err != nil {
						return err
					} else {
						changed = changed || ret
					}

					if ret, err := walker.Migrate(record, "graphContent"); err != nil {
						return err
					} else {
						changed = changed || ret
					}

					if changed {
						if err := app.Save(record); err != nil {
							return err
						}

						tracer.Printf("record #%s in collection '%s' updated", record.Id, collection.Name)
					}
				}
			}

			// update collection `workflow_run`
			//   - migrate field `graph`
			{
				collection, err := app.FindCollectionByNameOrId("qjp8lygssgwyqyz")
				if err != nil {
					return err
				}

				records, err := app.FindAllRecords(collection)
				if err != nil {
					return err
				}

				for _, record := range records {
					changed := false

					if ret, err := walker.Migrate(record, "graph"); err != nil {
						return err
					} else {
						changed = changed || ret
					}

					if changed {
						if err := app.Save(record); err != nil {
							return err
						}

						tracer.Printf("record #%s in collection '%s' updated", record.Id, collection.Name)
					}
				}
			}
		}

		// clean old migrations
		{
			migrations := []string{
				"1757476800_m0.4.0_migrate.go",
				"1757476801_m0.4.0_initialize.go",
				"1760486400_m0.4.1.go",
				"1762142400_m0.4.3.go",
				"1762516800_m0.4.4.go",
				"1763373600_m0.4.5.go",
				"1763553600_m0.4.6.go",
				"1763640000_m0.4.6.go",
			}
			for _, name := range migrations {
				app.DB().NewQuery("DELETE FROM _migrations WHERE file='" + name + "'").Execute()
			}
		}

		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		return nil
	})
}
