package migrations

import (
	"encoding/json"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"

	snaps "github.com/certimate-go/certimate/migrations/snaps/v0.4"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcertx509 "github.com/certimate-go/certimate/pkg/utils/cert/x509"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("v0.4.15")
		tracer.Printf("go ...")

		// update collection `acme_accounts`
		//   - rebuild indexes
		{
			collection, err := app.FindCollectionByNameOrId("012d7abbod1hwvr")
			if err != nil {
				return err
			}

			if err := json.Unmarshal([]byte(`{
				"indexes": [
					"CREATE INDEX `+"`"+`idx_dQiYzimY7m`+"`"+` ON `+"`"+`acme_accounts`+"`"+` (`+"`"+`ca`+"`"+`)",
					"CREATE INDEX `+"`"+`idx_TjyqY6LAGa`+"`"+` ON `+"`"+`acme_accounts`+"`"+` (\n  `+"`"+`ca`+"`"+`,\n  `+"`"+`acmeDirUrl`+"`"+`\n)",
					"CREATE UNIQUE INDEX `+"`"+`idx_G4brUDgxzc`+"`"+` ON `+"`"+`acme_accounts`+"`"+` (\n  `+"`"+`ca`+"`"+`,\n  `+"`"+`email`+"`"+`,\n  `+"`"+`acmeAcctUrl`+"`"+`,\n  `+"`"+`acmeDirUrl`+"`"+`\n)"
				]
			}`), &collection); err != nil {
				return err
			}

			if err := app.Save(collection); err != nil {
				return err
			}

			tracer.Printf("collection '%s' updated", collection.Name)
		}

		// update collection `certificate`
		//   - update field `subjectAltNames`
		//   - remove field `acmeCertStableUrl`
		{
			collection, err := app.FindCollectionByNameOrId("4szxr9x43tpj6np")
			if err != nil {
				return err
			}

			collection.Fields.RemoveByName("acmeCertStableUrl")

			if err := app.Save(collection); err != nil {
				return err
			}

			tracer.Printf("collection '%s' updated", collection.Name)

			records, err := app.FindAllRecords(collection)
			if err != nil {
				return err
			}

			for _, record := range records {
				changed := false

				if certX509, err := xcert.ParseCertificateFromPEM(record.GetString("certificate")); err == nil {
					certSANs := xcertx509.GetSubjectAltNames(certX509)
					if strings.Join(certSANs, ";") != record.GetString("subjectAltNames") {
						record.Set("subjectAltNames", strings.Join(certSANs, ";"))
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

		// update collection `workflow_output`
		//   - revert data for #1137
		{
			collection, err := app.FindCollectionByNameOrId("bqnxb95f2cooowp")
			if err != nil {
				return err
			}

			records, err := app.FindAllRecords(collection)
			if err != nil {
				return err
			}

			for _, record := range records {
				changed := false

				runRecord, _ := app.FindFirstRecordByFilter("workflow_run", "id={:runId}", dbx.Params{"runId": record.GetString("runRef")})
				if runRecord != nil {
					runGraph := make(map[string]any)
					if err := runRecord.UnmarshalJSONField("graph", &runGraph); err != nil {
						return err
					}

					if _, ok := runGraph["nodes"]; ok {
						nodes := make([]*snaps.WorkflowNode, 0)
						if err := mapstructure.Decode(runGraph["nodes"], &nodes); err != nil {
							return err
						}

						nodeMaybeBrokenId := record.GetString("nodeId")

						var findNode func(blocks []*snaps.WorkflowNode) *snaps.WorkflowNode
						findNode = func(blocks []*snaps.WorkflowNode) *snaps.WorkflowNode {
							for _, node := range blocks {
								if node.Id == nodeMaybeBrokenId {
									return node
								}
								if len(node.Blocks) > 0 {
									if node := findNode(node.Blocks); node != nil {
										return node
									}
								}
							}
							return nil
						}
						if node := findNode(nodes); node != nil {
							continue
						}

						var findNodeEx func(blocks []*snaps.WorkflowNode) *snaps.WorkflowNode
						findNodeEx = func(blocks []*snaps.WorkflowNode) *snaps.WorkflowNode {
							for _, node := range blocks {
								const TRUNCATED_LENGTH = 3 // same as `ATTEMPTS` in '1757476800_upgrade_v0.4.0.go'
								if strings.HasSuffix(node.Id, nodeMaybeBrokenId) && (len(node.Id)-len(nodeMaybeBrokenId) == TRUNCATED_LENGTH) {
									return node
								}
								if len(node.Blocks) > 0 {
									if node := findNodeEx(node.Blocks); node != nil {
										return node
									}
								}
							}
							return nil
						}
						if node := findNodeEx(nodes); node != nil {
							record.Set("nodeId", node.Id)
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

				if nodeCfg["provider"] == "rainyun-rcdn" {
					if providerCfg, ok := nodeCfg["providerConfig"].(map[string]any); ok {
						if providerCfg["resourceType"] == "certificate" {
							delete(providerCfg, "resourceType")
							delete(providerCfg, "instanceId")
							delete(providerCfg, "domainMatchPattern")
							delete(providerCfg, "domain")
							nodeCfg["provider"] = "rainyun-sslcenter"
							nodeCfg["providerConfig"] = providerCfg
						} else {
							delete(providerCfg, "resourceType")
							delete(providerCfg, "certificateId")
							nodeCfg["providerConfig"] = providerCfg
						}
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
		return nil
	})
}
