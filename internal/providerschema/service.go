package providerschema

import "context"

type ProviderSchemaService struct {
	repo providerSchemaRepository
}

func NewProviderSchemaService(repo providerSchemaRepository) *ProviderSchemaService {
	return &ProviderSchemaService{repo: repo}
}

func (s *ProviderSchemaService) GetByProviderType(ctx context.Context, providerType string) (*Envelope, error) {
	schema, err := s.repo.Get(ctx, providerType)
	if err != nil {
		return nil, err
	}
	return Emit(schema), nil
}

func (s *ProviderSchemaService) List(ctx context.Context) ([]*Envelope, error) {
	schemas, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*Envelope, len(schemas))
	for i, sc := range schemas {
		out[i] = Emit(sc)
	}
	return out, nil
}
