package domain

import (
	"time"
)

const CollectionNameWorkflowRun = "workflow_run"

type WorkflowRun struct {
	Meta
	WorkflowId string                `db:"workflowRef" json:"workflowId"`
	Status     WorkflowRunStatusType `db:"status"      json:"status"`
	Trigger    WorkflowTriggerType   `db:"trigger"     json:"trigger"`
	StartedAt  time.Time             `db:"startedAt"   json:"startedAt"`
	EndedAt    time.Time             `db:"endedAt"     json:"endedAt"`
	Graph      *WorkflowGraph        `db:"graph"       json:"graph"`
	Error      string                `db:"error"       json:"error"`
}

type WorkflowRunStatusType string

func (t WorkflowRunStatusType) String() string {
	return string(t)
}

const (
	WorkflowRunStatusTypePending    WorkflowRunStatusType = "pending"
	WorkflowRunStatusTypeProcessing WorkflowRunStatusType = "processing"
	WorkflowRunStatusTypeSucceeded  WorkflowRunStatusType = "succeeded"
	WorkflowRunStatusTypeFailed     WorkflowRunStatusType = "failed"
	WorkflowRunStatusTypeCanceled   WorkflowRunStatusType = "canceled"
)
