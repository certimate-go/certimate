package migrations

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		if mr, _ := app.FindFirstRecordByFilter("_migrations", "file='1762142400_m0.4.3.go'"); mr != nil {
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
		}

		// adapt to new workflow data structure
		{
			walker := &mWorkflowGraphWalker{}
			walker.Define(func(node *mWorkflowNode) (_changed bool, _err error) {
				_changed = false
				_err = nil

				if node.Type != "bizApply" {
					return
				}

				if node.Data == nil {
					return
				}

				if _, ok := node.Data["config"]; ok {
					nodeCfg := node.Data["config"].(map[string]any)
					if nodeCfg["keySource"] == nil || nodeCfg["keySource"] == "" {
						nodeCfg["keySource"] = "auto"

						node.Data["config"] = nodeCfg
						_changed = true
						return
					}
				}

				return
			})
			walker.Define(func(node *mWorkflowNode) (_changed bool, _err error) {
				_changed = false
				_err = nil

				if node.Type != "bizUpload" {
					return
				}

				if node.Data == nil {
					return
				}

				if _, ok := node.Data["config"]; ok {
					nodeCfg := node.Data["config"].(map[string]any)
					if nodeCfg["source"] == nil || nodeCfg["source"] == "" {
						nodeCfg["source"] = "form"

						node.Data["config"] = nodeCfg
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

					if record.GetRaw("graphDraft") != nil {
						graph := make(map[string]any)
						if err := record.UnmarshalJSONField("graphDraft", &graph); err != nil {
							return err
						}

						if _, ok := graph["nodes"]; ok {
							nodes := make([]*mWorkflowNode, 0)
							if err := mapstructure.Decode(graph["nodes"], &nodes); err != nil {
								return err
							}

							nodesChanged, err := walker.Visit(nodes)
							if err != nil {
								return err
							} else if nodesChanged {
								graph["nodes"] = nodes
								record.Set("graphDraft", graph)
								changed = true
							}
						}
					}

					if record.GetRaw("graphContent") != nil {
						graph := make(map[string]any)
						if err := record.UnmarshalJSONField("graphContent", &graph); err != nil {
							return err
						}

						if _, ok := graph["nodes"]; ok {
							nodes := make([]*mWorkflowNode, 0)
							if err := mapstructure.Decode(graph["nodes"], &nodes); err != nil {
								return err
							}

							nodesChanged, err := walker.Visit(nodes)
							if err != nil {
								return err
							} else if nodesChanged {
								graph["nodes"] = nodes
								record.Set("graphContent", graph)
								changed = true
							}
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

					if record.GetRaw("graph") != nil {
						graph := make(map[string]any)
						if err := record.UnmarshalJSONField("graph", &graph); err != nil {
							return err
						}

						if _, ok := graph["nodes"]; ok {
							nodes := make([]*mWorkflowNode, 0)
							if err := mapstructure.Decode(graph["nodes"], &nodes); err != nil {
								return err
							}

							nodesChanged, err := walker.Visit(nodes)
							if err != nil {
								return err
							} else if nodesChanged {
								graph["nodes"] = nodes
								record.Set("graph", graph)
								changed = true
							}
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
		}

		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		return nil
	})
}
