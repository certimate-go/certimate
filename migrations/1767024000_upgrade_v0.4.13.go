package migrations

import (
	"errors"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"

	snaps "github.com/certimate-go/certimate/migrations/snaps/v0.4"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("v0.4.13")
		tracer.Printf("go ...")

		// adapt to new workflow data structure
		{
			walker := &snaps.WorkflowGraphWalker{}
			walker.Define(func(node *snaps.WorkflowNode) (_changed bool, _err error) {
				_changed = false
				_err = nil

				if node.Type != "bizDeploy" {
					return
				}

				nodeCfg := node.Data.Config

				switch nodeCfg["provider"] {
				case "1panel-site":
					{
						nodeCfg["provider"] = "1panel"

						_changed = true
						return
					}

				case "baotapanel-site":
					{
						nodeCfg["provider"] = "baotapanel"

						_changed = true
						return
					}

				case "baotapanelgo-site":
					{
						nodeCfg["provider"] = "baotapanelgo"

						_changed = true
						return
					}

				case "baotawaf-site":
					{
						nodeCfg["provider"] = "baotawaf"

						_changed = true
						return
					}

				case "cdnfly":
					{
						if providerCfg, ok := nodeCfg["providerConfig"].(map[string]any); ok {
							if providerCfg["resourceType"] == "site" {
								providerCfg["resourceType"] = "website"
								nodeCfg["providerConfig"] = providerCfg

								_changed = true
								return
							}
						}
					}

				case "cpanel-site":
					{
						if providerCfg, ok := nodeCfg["providerConfig"].(map[string]any); ok {
							providerCfg["resourceType"] = "website"
							nodeCfg["providerConfig"] = providerCfg
						}

						nodeCfg["provider"] = "cpanel"

						_changed = true
						return
					}

				case "netlify-site":
					{
						if providerCfg, ok := nodeCfg["providerConfig"].(map[string]any); ok {
							providerCfg["resourceType"] = "website"
							nodeCfg["providerConfig"] = providerCfg
						}

						nodeCfg["provider"] = "netlify"

						_changed = true
						return
					}

				case "ratpanel-site":
					{
						if providerCfg, ok := nodeCfg["providerConfig"].(map[string]any); ok {
							providerCfg["resourceType"] = "website"
							nodeCfg["providerConfig"] = providerCfg
						}

						nodeCfg["provider"] = "ratpanel"

						_changed = true
						return
					}

				case "safeline-site":
					{
						nodeCfg["provider"] = "safeline"

						_changed = true
						return
					}
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
		return errors.ErrUnsupported
	})
}
