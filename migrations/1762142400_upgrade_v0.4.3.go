package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"

	snaps "github.com/certimate-go/certimate/migrations/snaps/v0.4"
)

func init() {
	m.Register(func(app core.App) error {
		if err := app.DB().
			NewQuery("SELECT (1) FROM _migrations WHERE file={:file} LIMIT 1").
			Bind(dbx.Params{"file": "1762142400_m0.4.3.go"}).
			One(&struct{}{}); err == nil {
			return nil
		}

		tracer := NewTracer("v0.4.3")
		tracer.Printf("go ...")

		// update collection `certificate`
		//   - rename field `acmeRenewed` to `isRenewed`
		//   - add field `isRevoked`
		//   - add field `validityInterval`
		{
			collection, err := app.FindCollectionByNameOrId("4szxr9x43tpj6np")
			if err != nil {
				return err
			}

			if err := collection.Fields.AddMarshaledJSONAt(11, []byte(`{
				"hidden": false,
				"id": "number2453290051",
				"max": null,
				"min": null,
				"name": "validityInterval",
				"onlyInt": false,
				"presentable": false,
				"required": false,
				"system": false,
				"type": "number"
			}`)); err != nil {
				return err
			}

			if err := collection.Fields.AddMarshaledJSONAt(14, []byte(`{
				"hidden": false,
				"id": "bool810050391",
				"name": "isRenewed",
				"presentable": false,
				"required": false,
				"system": false,
				"type": "bool"
			}`)); err != nil {
				return err
			}

			if err := collection.Fields.AddMarshaledJSONAt(15, []byte(`{
				"hidden": false,
				"id": "bool3680845581",
				"name": "isRevoked",
				"presentable": false,
				"required": false,
				"system": false,
				"type": "bool"
			}`)); err != nil {
				return err
			}

			if err := app.Save(collection); err != nil {
				return err
			}

			if _, err := app.DB().NewQuery("UPDATE certificate SET validityInterval = (STRFTIME('%s', validityNotAfter) - STRFTIME('%s', validityNotBefore))").Execute(); err != nil {
				return err
			}

			tracer.Printf("collection '%s' updated", collection.Name)
		}

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

				if nodeCfg["keySource"] == nil || nodeCfg["keySource"] == "" {
					nodeCfg["keySource"] = "auto"

					_changed = true
					return
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

				if nodeCfg["source"] == nil || nodeCfg["source"] == "" {
					nodeCfg["source"] = "form"

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

		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		return nil
	})
}
