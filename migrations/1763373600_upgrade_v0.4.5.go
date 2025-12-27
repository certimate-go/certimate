package migrations

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		if mr, _ := app.FindFirstRecordByFilter("_migrations", "file='1763373600_m0.4.5.go'"); mr != nil {
			return nil
		}

		tracer := NewTracer("v0.4.5")
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

				if _, ok := node.Data["config"]; ok {
					nodeCfg := node.Data["config"].(map[string]any)

					provider := nodeCfg["provider"]
					switch provider {
					case "aliyun-waf":
						{
							if nodeCfg["providerConfig"] != nil {
								providerCfg := nodeCfg["providerConfig"].(map[string]any)
								providerCfg["serviceType"] = "cname"
								nodeCfg["providerConfig"] = providerCfg

								node.Data["config"] = nodeCfg
								_changed = true
								return
							}
						}

					case "baishan-cdn":
					case "ksyun-cdn":
					case "rainyun-rcdn":
						{
							if nodeCfg["providerConfig"] != nil {
								providerCfg := nodeCfg["providerConfig"].(map[string]any)
								if providerCfg["certificateId"] != nil && providerCfg["certificateId"].(string) != "" {
									providerCfg["resourceType"] = "certificate"
								} else {
									providerCfg["resourceType"] = "domain"
								}
								nodeCfg["providerConfig"] = providerCfg

								node.Data["config"] = nodeCfg
								_changed = true
								return
							}
						}

					case "tencentcloud-ssldeploy":
						{
							if nodeCfg["providerConfig"] != nil {
								providerCfg := nodeCfg["providerConfig"].(map[string]any)
								providerCfg["resourceProduct"] = providerCfg["resourceType"]
								delete(providerCfg, "resourceType")
								nodeCfg["providerConfig"] = providerCfg

								node.Data["config"] = nodeCfg
								_changed = true
								return
							}
						}

					case "tencentcloud-sslupdate":
						{
							if nodeCfg["providerConfig"] != nil {
								providerCfg := nodeCfg["providerConfig"].(map[string]any)
								providerCfg["resourceProducts"] = providerCfg["resourceTypes"]
								delete(providerCfg, "resourceTypes")
								nodeCfg["providerConfig"] = providerCfg

								node.Data["config"] = nodeCfg
								_changed = true
								return
							}
						}
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

				if _, err := app.DB().NewQuery("UPDATE workflow SET graphDraft = REPLACE(graphDraft, '\"matchPattern\"', '\"domainMatchPattern\"')").Execute(); err != nil {
					return err
				}

				if _, err := app.DB().NewQuery("UPDATE workflow SET graphContent = REPLACE(graphContent, '\"matchPattern\"', '\"domainMatchPattern\"')").Execute(); err != nil {
					return err
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

				if _, err := app.DB().NewQuery("UPDATE workflow_run SET graph = REPLACE(graph, '\"matchPattern\"', '\"domainMatchPattern\"')").Execute(); err != nil {
					return err
				}
			}

			// update collection `workflow_output`
			//   - migrate field `nodeConfig`
			{
				if _, err := app.DB().NewQuery("UPDATE workflow_output SET nodeConfig = REPLACE(nodeConfig, '\"matchPattern\"', '\"domainMatchPattern\"')").Execute(); err != nil {
					return err
				}

				if _, err := app.DB().NewQuery("UPDATE workflow_output SET nodeConfig = REPLACE(nodeConfig, '\"resourceType\"', '\"resourceProduct\"') WHERE nodeConfig LIKE '%\"provider\":\"tencentcloud-ssldeploy\"%' OR nodeConfig LIKE '%\"provider\":\"tencentcloud-sslupdate\"%'").Execute(); err != nil {
					return err
				}
			}
		}

		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		return nil
	})
}
