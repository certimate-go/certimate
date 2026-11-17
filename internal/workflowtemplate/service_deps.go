package workflowtemplate

import (
	"context"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/providerschema"
)

type workflowTemplateRepository interface {
	GetById(ctx context.Context, id string) (*domain.WorkflowTemplate, error)
	List(ctx context.Context, keyword string) ([]*domain.WorkflowTemplate, error)
	Save(ctx context.Context, template *domain.WorkflowTemplate) (*domain.WorkflowTemplate, error)
	Delete(ctx context.Context, id string) error
}

type workflowRepository interface {
	Save(ctx context.Context, workflow *domain.Workflow) (*domain.Workflow, error)
}

type accessRepository interface {
	GetById(ctx context.Context, id string) (*domain.Access, error)
}

type providerSchemaService interface {
	GetByProviderType(ctx context.Context, providerType string, categories ...providerschema.Category) (*providerschema.Envelope, error)
}
