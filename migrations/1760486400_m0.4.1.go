package migrations

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("v0.4.1")
		tracer.Printf("go ...")

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

			type dWorkflowNode struct {
				Id     string           `json:"id"`
				Type   string           `json:"type"`
				Data   map[string]any   `json:"data"`
				Blocks []*dWorkflowNode `json:"blocks,omitempty,omitzero"`
			}

			var deepMigrateNode func(node *dWorkflowNode) (_node *dWorkflowNode, _migrated bool)
			var deepMigrateNodes func(nodes []*dWorkflowNode) (_nodes []*dWorkflowNode, _migrated bool)
			deepMigrateNode = func(node *dWorkflowNode) (*dWorkflowNode, bool) {
				migrated := false

				if node.Type == "bizDeploy" {
					if node.Data != nil {
						if _, ok := node.Data["config"]; ok {
							nodeCfg := node.Data["config"].(map[string]any)
							if nodeCfg["provider"] == "local" && nodeCfg["providerAccessId"] != nil {
								delete(nodeCfg, "providerAccessId")
								node.Data["config"] = nodeCfg
								migrated = true
							}
						}
					}
				}

				if len(node.Blocks) > 0 {
					if newBlocks, changed := deepMigrateNodes(node.Blocks); changed {
						node.Blocks = newBlocks
						migrated = true
					}
				}

				return node, migrated
			}
			deepMigrateNodes = func(nodes []*dWorkflowNode) ([]*dWorkflowNode, bool) {
				migrated := false

				for i, node := range nodes {
					if newNode, changed := deepMigrateNode(node); changed {
						nodes[i] = newNode
						migrated = true
					}
				}

				return nodes, migrated
			}

			for _, record := range records {
				changed := false

				graphDraft := make(map[string]any)
				if err := record.UnmarshalJSONField("graphDraft", &graphDraft); err == nil {
					if _, ok := graphDraft["nodes"]; ok {
						nodes := make([]*dWorkflowNode, 0)
						if err := mapstructure.Decode(graphDraft["nodes"], &nodes); err != nil {
							return err
						}

						if newNodes, migrated := deepMigrateNodes(nodes); migrated {
							graphDraft["nodes"] = newNodes
							record.Set("graphDraft", graphDraft)
							changed = true
						}
					}
				}

				graphContent := make(map[string]any)
				if err := record.UnmarshalJSONField("graphContent", &graphContent); err == nil {
					if _, ok := graphContent["nodes"]; ok {
						nodes := make([]*dWorkflowNode, 0)
						if err := mapstructure.Decode(graphContent["nodes"], &nodes); err != nil {
							return err
						}

						if newNodes, migrated := deepMigrateNodes(nodes); migrated {
							graphContent["nodes"] = newNodes
							record.Set("graphContent", graphContent)
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

		return nil
	}, func(app core.App) error {
		return nil
	})
}
