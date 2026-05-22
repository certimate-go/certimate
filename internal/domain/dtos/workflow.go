package dtos

import (
	"github.com/certimate-go/certimate/internal/domain"
)

type WorkflowStartRunReq struct {
	WorkflowId string                     `json:"-"`
	RunTrigger domain.WorkflowTriggerType `json:"trigger"`
}

type WorkflowStartRunResp struct {
	RunId string `json:"runId"`
}

type WorkflowCancelRunReq struct {
	WorkflowId string `bind:"path" json:"-"`
	RunId      string `bind:"path" json:"-"`
}

type WorkflowCancelRunResp struct{}

type WorkflowStatisticsResp struct {
	Concurrency      int      `json:"concurrency"`
	PendingRunIds    []string `json:"pendingRunIds"`
	ProcessingRunIds []string `json:"processingRunIds"`
}
