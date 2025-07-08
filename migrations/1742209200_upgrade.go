package migrations

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("(v0.3)1742209200")
		tracer.Printf("go ...")

		// create collection `workflow_logs`
		{
			jsonData := `{
				"createRule": null,
				"deleteRule": null,
				"fields": [
					{
						"autogeneratePattern": "[a-z0-9]{15}",
						"hidden": false,
						"id": "text3208210256",
						"max": 15,
						"min": 15,
						"name": "id",
						"pattern": "^[a-z0-9]+$",
						"presentable": false,
						"primaryKey": true,
						"required": true,
						"system": true,
						"type": "text"
					},
					{
						"cascadeDelete": true,
						"collectionId": "tovyif5ax6j62ur",
						"hidden": false,
						"id": "relation3371272342",
						"maxSelect": 1,
						"minSelect": 0,
						"name": "workflowId",
						"presentable": false,
						"required": false,
						"system": false,
						"type": "relation"
					},
					{
						"cascadeDelete": true,
						"collectionId": "qjp8lygssgwyqyz",
						"hidden": false,
						"id": "relation821863227",
						"maxSelect": 1,
						"minSelect": 0,
						"name": "runId",
						"presentable": false,
						"required": false,
						"system": false,
						"type": "relation"
					},
					{
						"autogeneratePattern": "",
						"hidden": false,
						"id": "text157423495",
						"max": 0,
						"min": 0,
						"name": "nodeId",
						"pattern": "",
						"presentable": false,
						"primaryKey": false,
						"required": false,
						"system": false,
						"type": "text"
					},
					{
						"autogeneratePattern": "",
						"hidden": false,
						"id": "text3227511481",
						"max": 0,
						"min": 0,
						"name": "nodeName",
						"pattern": "",
						"presentable": false,
						"primaryKey": false,
						"required": false,
						"system": false,
						"type": "text"
					},
					{
						"hidden": false,
						"id": "number2782324286",
						"max": null,
						"min": null,
						"name": "timestamp",
						"onlyInt": false,
						"presentable": false,
						"required": false,
						"system": false,
						"type": "number"
					},
					{
						"autogeneratePattern": "",
						"hidden": false,
						"id": "text2599078931",
						"max": 0,
						"min": 0,
						"name": "level",
						"pattern": "",
						"presentable": false,
						"primaryKey": false,
						"required": false,
						"system": false,
						"type": "text"
					},
					{
						"autogeneratePattern": "",
						"hidden": false,
						"id": "text3065852031",
						"max": 0,
						"min": 0,
						"name": "message",
						"pattern": "",
						"presentable": false,
						"primaryKey": false,
						"required": false,
						"system": false,
						"type": "text"
					},
					{
						"hidden": false,
						"id": "json2918445923",
						"maxSize": 0,
						"name": "data",
						"presentable": false,
						"required": false,
						"system": false,
						"type": "json"
					},
					{
						"hidden": false,
						"id": "autodate2990389176",
						"name": "created",
						"onCreate": true,
						"onUpdate": false,
						"presentable": false,
						"system": false,
						"type": "autodate"
					}
				],
				"id": "pbc_1682296116",
				"indexes": [
					"CREATE INDEX ` + "`" + `idx_IOlpy6XuJ2` + "`" + ` ON ` + "`" + `workflow_logs` + "`" + ` (` + "`" + `workflowId` + "`" + `)",
					"CREATE INDEX ` + "`" + `idx_qVlTb2yl7v` + "`" + ` ON ` + "`" + `workflow_logs` + "`" + ` (` + "`" + `runId` + "`" + `)"
				],
				"listRule": null,
				"name": "workflow_logs",
				"system": false,
				"type": "base",
				"updateRule": null,
				"viewRule": null
			}`

			collection := &core.Collection{}
			if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
				return err
			}

			if err := app.Save(collection); err != nil {
				return err
			}

			tracer.Printf("collection '%s' created", collection.Name)
		}

		// migrate data
		{
			workflowRuns, err := app.FindAllRecords("workflow_run")
			if err != nil {
				return err
			}

			for _, workflowRun := range workflowRuns {
				type dWorkflowRunLogRecord struct {
					Time    string `json:"time"`
					Level   string `json:"level"`
					Content string `json:"content"`
					Error   string `json:"error"`
				}
				type dWorkflowRunLog struct {
					NodeId   string                  `json:"nodeId"`
					NodeName string                  `json:"nodeName"`
					Records  []dWorkflowRunLogRecord `json:"records"`
					Error    string                  `json:"error"`
				}

				logs := make([]dWorkflowRunLog, 0)
				if err := workflowRun.UnmarshalJSONField("logs", &logs); err != nil {
					continue
				}

				collection, err := app.FindCollectionByNameOrId("workflow_logs")
				if err != nil {
					return err
				}

				for _, log := range logs {
					for _, logRecord := range log.Records {
						record := core.NewRecord(collection)
						createdAt, _ := time.Parse(time.RFC3339, logRecord.Time)
						record.Set("workflowId", workflowRun.Get("workflowId"))
						record.Set("runId", workflowRun.Get("id"))
						record.Set("nodeId", log.NodeId)
						record.Set("nodeName", log.NodeName)
						record.Set("timestamp", createdAt.UnixMilli())
						record.Set("level", logRecord.Level)
						record.Set("message", strings.TrimSpace(logRecord.Content+" "+logRecord.Error))
						record.Set("created", createdAt)
						if err := app.Save(record); err != nil {
							return err
						}

						tracer.Printf("record #%s in collection '%s' updated", record.Id, collection.Name)
					}
				}
			}
		}

		// update collection `workflow_run`
		{
			collection, err := app.FindCollectionByNameOrId("workflow_run")
			if err != nil {
				return err
			}

			if err := collection.Fields.AddMarshaledJSONAt(6, []byte(`{
				"hidden": false,
				"id": "json772177811",
				"maxSize": 5000000,
				"name": "detail",
				"presentable": false,
				"required": false,
				"system": false,
				"type": "json"
			}`)); err != nil {
				return err
			}

			if err := app.Save(collection); err != nil {
				return err
			}

			tracer.Printf("collection '%s' updated", collection.Name)
		}

		// migrate data
		{
			workflowRuns, err := app.FindAllRecords("workflow_run")
			if err != nil {
				return err
			}

			workflowOutputs, err := app.FindAllRecords("workflow_output")
			if err != nil {
				return err
			}

			type dWorkflowNode struct {
				Id        string           `json:"id"`
				Type      string           `json:"type"`
				Name      string           `json:"name"`
				Config    map[string]any   `json:"config,omitempty"`
				Inputs    []map[string]any `json:"inputs,omitempty"`
				Outputs   []map[string]any `json:"outputs,omitempty"`
				Next      *dWorkflowNode   `json:"next,omitempty"`
				Branches  []*dWorkflowNode `json:"branches,omitempty"`
				Validated bool             `json:"validated,omitempty"`
			}

			for _, workflowRun := range workflowRuns {
				node := &dWorkflowNode{}
				for _, workflowOutput := range workflowOutputs {
					if workflowOutput.GetString("runId") != workflowRun.Get("id") {
						continue
					}

					if err := workflowOutput.UnmarshalJSONField("node", node); err != nil {
						continue
					}

					if node.Type != "apply" {
						node = &dWorkflowNode{}
						continue
					}
				}

				if node.Id == "" {
					workflow, _ := app.FindRecordById("workflow", workflowRun.GetString("workflowId"))
					if workflow != nil {
						workflowRun.Set("detail", workflow.Get("content"))
					} else {
						workflowRun.Set("detail", make(map[string]any))
					}
				} else {
					workflow, _ := app.FindRecordById("workflow", workflowRun.GetString("workflowId"))
					if workflow != nil {
						rootNode := &dWorkflowNode{}
						if err := workflow.UnmarshalJSONField("content", rootNode); err != nil {
							return err
						}

						rootNode.Next = node
						workflowRun.Set("detail", rootNode)
					} else {
						rootNode := &dWorkflowNode{
							Id:   core.GenerateDefaultRandomId(),
							Type: "start",
							Name: "开始",
							Config: map[string]any{
								"trigger": "manual",
							},
							Next:      node,
							Validated: true,
						}
						workflowRun.Set("detail", rootNode)
					}
				}

				if err := app.Save(workflowRun); err != nil {
					return err
				}

				tracer.Printf("record #%s in collection '%s' updated", workflowRun.Id, workflowRun.Collection().Name)
			}
		}

		// update collection `workflow_run`
		{
			collection, err := app.FindCollectionByNameOrId("workflow_run")
			if err != nil {
				return err
			}

			collection.Fields.RemoveByName("logs")

			if err := app.Save(collection); err != nil {
				return err
			}

			tracer.Printf("collection '%s' updated", collection.Name)
		}

		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		return nil
	})
}
