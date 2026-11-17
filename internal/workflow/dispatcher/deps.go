package dispatcher

import (
	"context"

	"github.com/certimate-go/certimate/internal/domain"
)

type workflowRepository interface {
	GetById(ctx context.Context, id string) (*domain.Workflow, error)
	Save(ctx context.Context, workflow *domain.Workflow) (*domain.Workflow, error)
}

type workflowRunRepository interface {
	GetById(ctx context.Context, id string) (*domain.WorkflowRun, error)
	Save(ctx context.Context, workflowRun *domain.WorkflowRun) (*domain.WorkflowRun, error)
	SaveWithCascading(ctx context.Context, workflowRun *domain.WorkflowRun) (*domain.WorkflowRun, error)
	ResetStatusIfHanging(ctx context.Context) error
}

type workflowLogRepository interface {
	Save(ctx context.Context, workflowLog *domain.WorkflowLog) (*domain.WorkflowLog, error)
}
