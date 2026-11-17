package dtos

import (
	"github.com/certimate-go/certimate/internal/domain"
)

type WorkflowTemplateParamError struct {
	ParamId string `json:"paramId"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

type WorkflowTemplateParamErrors struct {
	Items []WorkflowTemplateParamError
}

func (e *WorkflowTemplateParamErrors) Error() string {
	if len(e.Items) == 0 {
		return "invalid template parameters"
	}
	return e.Items[0].Message
}

type WorkflowTemplateCreateReq struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Graph       *domain.WorkflowGraph         `json:"graph"`
	Params      domain.WorkflowTemplateParams `json:"params"`
}

type WorkflowTemplateUpdateReq struct {
	Name             string                        `json:"name"`
	Description      string                        `json:"description"`
	Graph            *domain.WorkflowGraph         `json:"graph"`
	Params           domain.WorkflowTemplateParams `json:"params"`
	ExpectedRevision int                           `json:"expectedRevision"`
}

type WorkflowTemplateInstantiateReq struct {
	Revision    int                                   `json:"revision"`
	Name        string                                `json:"name"`
	Description string                                `json:"description"`
	Params      map[string]WorkflowTemplateParamValue `json:"params"`
}

type WorkflowTemplateParamValue struct {
	Domains          []string       `json:"domains,omitempty"`
	ProviderAccessId string         `json:"providerAccessId,omitempty"`
	ProviderConfig   map[string]any `json:"providerConfig,omitempty"`
}

type WorkflowTemplateInstantiateResp struct {
	WorkflowId string `json:"workflowId"`
}
