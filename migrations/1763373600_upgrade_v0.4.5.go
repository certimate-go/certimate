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
			Bind(dbx.Params{"file": "1763373600_m0.4.5.go"}).
			One(&struct{}{}); err == nil {
			return nil
		}

		tracer := NewTracer("v0.4.5")
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
				case "aliyun-waf":
					{
						if providerCfg, ok := nodeCfg["providerConfig"].(map[string]any); ok {
							providerCfg["serviceType"] = "cname"
							nodeCfg["providerConfig"] = providerCfg

							_changed = true
							return
						}
					}

				case "baishan-cdn":
				case "ksyun-cdn":
				case "rainyun-rcdn":
					{
						if providerCfg, ok := nodeCfg["providerConfig"].(map[string]any); ok {
							if providerCfg["certificateId"] != nil && providerCfg["certificateId"].(string) != "" {
								providerCfg["resourceType"] = "certificate"
							} else {
								providerCfg["resourceType"] = "domain"
							}
							nodeCfg["providerConfig"] = providerCfg

							_changed = true
							return
						}
					}

				case "tencentcloud-ssldeploy":
					{
						if providerCfg, ok := nodeCfg["providerConfig"].(map[string]any); ok {
							providerCfg["resourceProduct"] = providerCfg["resourceType"]
							delete(providerCfg, "resourceType")
							nodeCfg["providerConfig"] = providerCfg

							_changed = true
							return
						}
					}

				case "tencentcloud-sslupdate":
					{
						if providerCfg, ok := nodeCfg["providerConfig"].(map[string]any); ok {
							providerCfg["resourceProducts"] = providerCfg["resourceTypes"]
							delete(providerCfg, "resourceTypes")
							nodeCfg["providerConfig"] = providerCfg

							_changed = true
							return
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
