package providerschema

import (
	"context"
	"fmt"
)

type ProviderSchemaService struct {
	repo providerSchemaRepository
}

func NewProviderSchemaService(repo providerSchemaRepository) *ProviderSchemaService {
	return &ProviderSchemaService{repo: repo}
}

func (s *ProviderSchemaService) GetByProviderType(ctx context.Context, providerType string, categories ...Category) (*Envelope, error) {
	if env, ok, _ := s.repo.GetEnvelope(ctx, providerType, categories...); ok && env != nil {
		return env, nil
	}
	schema, err := s.repo.Get(ctx, providerType)
	if err != nil {
		return nil, err
	}
	if len(categories) > 0 && categories[0] != "" && schema.Category != categories[0] {
		return nil, fmt.Errorf("schema for provider %q in category %q not found", providerType, categories[0])
	}
	return Emit(schema), nil
}

func (s *ProviderSchemaService) List(ctx context.Context) ([]*Envelope, error) {
	builtIn, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*Envelope, 0, len(builtIn))
	for _, sc := range builtIn {
		out = append(out, Emit(sc))
	}
	if envs, err := s.repo.ListEnvelopes(ctx); err == nil {
		out = append(out, envs...)
	}
	return out, nil
}
