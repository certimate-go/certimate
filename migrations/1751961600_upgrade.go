package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tracer := NewTracer("(v0.3)1751961600")
		tracer.Printf("go ...")

		// migrate data
		{
			workflows, err := app.FindAllRecords("workflow")
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

			deepChangeFn := func(node *dWorkflowNode) bool {
				stack := []*dWorkflowNode{node}

				for len(stack) > 0 {
					current := stack[len(stack)-1]
					stack = stack[:len(stack)-1]

					if current.Type == "deploy" {
						configMap := current.Config
						if configMap != nil {
							if provider, ok := configMap["provider"]; ok {
								if provider.(string) == "tencentcloud-eo" {
									if providerConfig, ok := configMap["providerConfig"]; ok {
										if providerConfigMap, ok := providerConfig.(map[string]any); ok {
											if _, ok := providerConfigMap["domain"]; ok {
												providerConfigMap["domains"] = providerConfigMap["domain"]
												delete(providerConfigMap, "domain")
												configMap["providerConfig"] = providerConfigMap
												return true
											}
										}
									}
								}
							}
						}
					}

					if current.Next != nil {
						stack = append(stack, current.Next)
					}

					if current.Branches != nil {
						for i := len(current.Branches) - 1; i >= 0; i-- {
							stack = append(stack, current.Branches[i])
						}
					}
				}

				return false
			}

			for _, workflow := range workflows {
				changed := false

				rootNodeContent := &dWorkflowNode{}
				if err := workflow.UnmarshalJSONField("content", rootNodeContent); err != nil {
					return err
				} else {
					if deepChangeFn(rootNodeContent) {
						workflow.Set("content", rootNodeContent)
						changed = true
					}
				}

				rootNodeDraft := &dWorkflowNode{}
				if err := workflow.UnmarshalJSONField("draft", rootNodeDraft); err != nil {
					return err
				} else {
					if deepChangeFn(rootNodeDraft) {
						workflow.Set("draft", rootNodeDraft)
						changed = true
					}
				}

				if changed {
					err = app.Save(workflow)
					if err != nil {
						return err
					}

					tracer.Printf("record #%s in collection '%s' updated", workflow.Id, workflow.Collection().Name)
				}
			}
		}

		tracer.Printf("done")
		return nil
	}, func(app core.App) error {
		return nil
	})
}
