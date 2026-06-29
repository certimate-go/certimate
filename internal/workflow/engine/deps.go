package engine

import (
	"context"

	"github.com/certimate-go/certimate/internal/domain"
)

type accessRepository interface {
	GetById(ctx context.Context, id string) (*domain.Access, error)
}

type certificateRepository interface {
	GetById(ctx context.Context, id string) (*domain.Certificate, error)
	GetByWorkflowRunIdAndNodeId(ctx context.Context, workflowRunId string, workflowNodeId string) (*domain.Certificate, error)
	Save(ctx context.Context, certificate *domain.Certificate) (*domain.Certificate, error)
}

type workflowOutputRepository interface {
	GetByWorkflowIdAndNodeId(ctx context.Context, workflowId string, workflowNodeId string) (*domain.WorkflowOutput, error)
	Save(ctx context.Context, workflowOutput *domain.WorkflowOutput) (*domain.WorkflowOutput, error)
}
