package repository

import (
	"context"
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/providerschema"
)

type ProviderSchemaRepository struct{}

func NewProviderSchemaRepository() *ProviderSchemaRepository {
	return &ProviderSchemaRepository{}
}

func (r *ProviderSchemaRepository) Get(_ context.Context, providerType string) (*providerschema.Schema, error) {
	if !providerschema.Registries.Has(providerType) {
		return nil, domain.NewError(domain.ErrRecordNotFound.Code, fmt.Sprintf("schema for provider %q not found", providerType))
	}
	return providerschema.Registries.Get(providerType)
}

func (r *ProviderSchemaRepository) List(_ context.Context) ([]*providerschema.Schema, error) {
	keys := providerschema.Registries.List()
	out := make([]*providerschema.Schema, 0, len(keys))
	for _, k := range keys {
		s, err := providerschema.Registries.Get(k)
		if err != nil {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *ProviderSchemaRepository) GetEnvelope(_ context.Context, providerType string) (*providerschema.Envelope, bool, error) {
	env, ok := providerschema.Envelopes.GetEnvelope(providerType)
	return env, ok, nil
}

func (r *ProviderSchemaRepository) ListEnvelopes(_ context.Context) ([]*providerschema.Envelope, error) {
	return providerschema.Envelopes.ListEnvelopes(), nil
}
