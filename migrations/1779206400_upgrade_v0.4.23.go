package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"

	snaps "github.com/certimate-go/certimate/migrations/snaps/v0.4"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("v0.4.23")
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
				case "1panel", "aliyun-alb", "aliyun-clb", "aliyun-ga", "aliyun-nlb", "apisix", "baiducloud-appblb", "baiducloud-blb", "baishan-cdn", "cdnfly", "cpanel", "ctcccloud-elb", "flexcdn", "goedge", "huaweicloud-apig", "huaweicloud-elb", "huaweicloud-waf", "jdcloud-alb", "kong", "ksyun-cdn", "ksyun-slb", "lecdn", "netlify", "nginxproxymanager", "ratpanel", "safeline", "samwaf", "tencentcloud-clb", "tencentcloud-gaap", "ucloud-uclb", "volcengine-alb", "volcengine-clb", "zenlayer-cdn", "zenlayer-ga":
					{
						if providerCfg, ok := nodeCfg["providerConfig"].(map[string]any); ok {
							if providerCfg["resourceType"] != nil && providerCfg["resourceType"].(string) != "" {
								providerCfg["deployTarget"] = providerCfg["resourceType"]
								delete(providerCfg, "resourceType")
								nodeCfg["providerConfig"] = providerCfg

								_changed = true
								return
							}
						}
					}
				case "ftp", "local", "ssh":
					{
						if providerCfg, ok := nodeCfg["providerConfig"].(map[string]any); ok {
							if providerCfg["format"] != nil && providerCfg["format"].(string) != "" {
								providerCfg["fileFormat"] = providerCfg["format"]
								providerCfg["filePathForKey"] = providerCfg["keyPath"]
								providerCfg["filePathForCrt"] = providerCfg["certPath"]
								providerCfg["filePathForCrtOnlyServer"] = providerCfg["certPathForServerOnly"]
								providerCfg["filePathForCrtOnlyIntermedia"] = providerCfg["certPathForIntermediaOnly"]
								delete(providerCfg, "format")
								delete(providerCfg, "keyPath")
								delete(providerCfg, "certPath")
								delete(providerCfg, "certPathForServerOnly")
								delete(providerCfg, "certPathForIntermediaOnly")
								nodeCfg["providerConfig"] = providerCfg

								_changed = true
								return
							}
						}
					}
				case "s3":
					{
						if providerCfg, ok := nodeCfg["providerConfig"].(map[string]any); ok {
							if providerCfg["format"] != nil && providerCfg["format"].(string) != "" {
								providerCfg["fileFormat"] = providerCfg["format"]
								providerCfg["objectKeyForKey"] = providerCfg["keyObjectKey"]
								providerCfg["objectKeyForCrt"] = providerCfg["certObjectKey"]
								providerCfg["objectKeyForCrtOnlyServer"] = providerCfg["certObjectKeyForServerOnly"]
								providerCfg["objectKeyForCrtOnlyIntermedia"] = providerCfg["certObjectKeyForIntermediaOnly"]
								delete(providerCfg, "format")
								delete(providerCfg, "keyObjectKey")
								delete(providerCfg, "certObjectKey")
								delete(providerCfg, "certObjectKeyForServerOnly")
								delete(providerCfg, "certObjectKeyForIntermediaOnly")
								nodeCfg["providerConfig"] = providerCfg

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
