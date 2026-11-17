package domain

import (
	"fmt"
	"sort"
	"strings"
)

const CollectionNameWorkflowTemplate = "workflow_template"

const WorkflowTemplateFormatVersion = 1

type WorkflowTemplate struct {
	Meta
	Name          string                 `db:"name"          json:"name"`
	Description   string                 `db:"description"   json:"description"`
	FormatVersion int                    `db:"formatVersion" json:"formatVersion"`
	Revision      int                    `db:"revision"      json:"revision"`
	Graph         *WorkflowGraph         `db:"graph"         json:"graph"`
	Params        WorkflowTemplateParams `db:"params"        json:"params"`
}

type WorkflowTemplateParamKind string

func (t WorkflowTemplateParamKind) String() string {
	return string(t)
}

const (
	WorkflowTemplateParamKindDomains    WorkflowTemplateParamKind = "domains"
	WorkflowTemplateParamKindDeployment WorkflowTemplateParamKind = "deployment"
)

type WorkflowTemplateParam struct {
	Id           string                    `json:"id"`
	Kind         WorkflowTemplateParamKind `json:"kind"`
	Title        string                    `json:"title"`
	Description  string                    `json:"description,omitempty"`
	Order        int                       `json:"order"`
	ApplyNodeIds []string                  `json:"applyNodeIds,omitempty"`
	DeployNodeId string                    `json:"deployNodeId,omitempty"`
}

type WorkflowTemplateParams []*WorkflowTemplateParam

func (s WorkflowTemplateParams) Sorted() WorkflowTemplateParams {
	sorted := make(WorkflowTemplateParams, len(s))
	copy(sorted, s)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Order != sorted[j].Order {
			return sorted[i].Order < sorted[j].Order
		}
		return sorted[i].Id < sorted[j].Id
	})
	return sorted
}

func (s WorkflowTemplateParams) Verify(graph *WorkflowGraph) error {
	paramIds := make(map[string]bool)
	usedApplyNodeIds := make(map[string]string)
	usedDeployNodeIds := make(map[string]string)

	for _, param := range s {
		if strings.TrimSpace(param.Id) == "" {
			return fmt.Errorf("the parameter ID cannot be blank")
		}
		if len(param.Id) > 32 {
			return fmt.Errorf("the parameter ID #%s is too long", param.Id)
		}
		if paramIds[param.Id] {
			return fmt.Errorf("duplicated parameter ID #%s", param.Id)
		}
		paramIds[param.Id] = true

		if strings.TrimSpace(param.Title) == "" {
			return fmt.Errorf("the parameter #%s title cannot be blank", param.Id)
		}

		switch param.Kind {
		case WorkflowTemplateParamKindDomains:
			{
				if len(param.ApplyNodeIds) == 0 {
					return fmt.Errorf("the parameter #%s is not bound to any apply nodes", param.Id)
				}

				for _, nodeId := range param.ApplyNodeIds {
					node, ok := graph.GetNodeById(nodeId)
					if !ok {
						return fmt.Errorf("the parameter #%s is bound to an unknown node #%s", param.Id, nodeId)
					}
					if node.Type != WorkflowNodeTypeBizApply {
						return fmt.Errorf("the parameter #%s is bound to a node #%s which is not an apply node", param.Id, nodeId)
					}

					identifier, _ := node.Data.Config["identifier"].(string)
					if identifier != "domain" {
						return fmt.Errorf("the parameter #%s is bound to an apply node #%s which is not configured for domains", param.Id, nodeId)
					}

					if prevParamId, ok := usedApplyNodeIds[nodeId]; ok {
						return fmt.Errorf("the node #%s is bound to both parameters #%s and #%s", nodeId, prevParamId, param.Id)
					}
					usedApplyNodeIds[nodeId] = param.Id
				}
			}

		case WorkflowTemplateParamKindDeployment:
			{
				if strings.TrimSpace(param.DeployNodeId) == "" {
					return fmt.Errorf("the parameter #%s is not bound to any deploy node", param.Id)
				}

				node, ok := graph.GetNodeById(param.DeployNodeId)
				if !ok {
					return fmt.Errorf("the parameter #%s is bound to an unknown node #%s", param.Id, param.DeployNodeId)
				}
				if node.Type != WorkflowNodeTypeBizDeploy {
					return fmt.Errorf("the parameter #%s is bound to a node #%s which is not a deploy node", param.Id, param.DeployNodeId)
				}

				provider, _ := node.Data.Config["provider"].(string)
				if strings.TrimSpace(provider) == "" {
					return fmt.Errorf("the parameter #%s is bound to a deploy node #%s without a specified provider", param.Id, param.DeployNodeId)
				}

				if prevParamId, ok := usedDeployNodeIds[param.DeployNodeId]; ok {
					return fmt.Errorf("the node #%s is bound to both parameters #%s and #%s", param.DeployNodeId, prevParamId, param.Id)
				}
				usedDeployNodeIds[param.DeployNodeId] = param.Id
			}

		default:
			return fmt.Errorf("the parameter #%s has an invalid kind", param.Id)
		}
	}

	return nil
}
