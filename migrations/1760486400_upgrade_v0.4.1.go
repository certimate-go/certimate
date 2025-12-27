package migrations

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		if mr, _ := app.FindFirstRecordByFilter("_migrations", "file='1760486400_m0.4.1.go'"); mr != nil {
			return nil
		}

		tracer := NewTracer("v0.4.1")
		tracer.Printf("go ...")

		// adapt to new workflow data structure
		{
			walker := &mWorkflowGraphWalker{}
			walker.Define(func(node *mWorkflowNode) (_changed bool, _err error) {
				_changed = false
				_err = nil

				if node.Type != "bizDeploy" {
					return
				}

				if node.Data == nil {
					return
				}

				if _, ok := node.Data["config"]; !ok {
					return
				}

				nodeCfg := node.Data["config"].(map[string]any)

				provider := nodeCfg["provider"]
				switch provider {
				case "local":
					{
						if nodeCfg["providerAccessId"] != nil {
							delete(nodeCfg, "providerAccessId")

							node.Data["config"] = nodeCfg
							_changed = true
							return
						}
					}
				}

				return
			})

			// update collection `workflow`
			//   - fix #982
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
		}

		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		return nil
	})
}
