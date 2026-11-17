package repository

import (
	"context"
	"testing"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/providerschema"
)

const testPSRProvider = "__test_provider_schema_repo__"

func registerTestSchema(t *testing.T) {
	t.Helper()
	if providerschema.Registries.Has(testPSRProvider) {
		return
	}
	s, err := providerschema.New(string(testPSRProvider), providerschema.CategoryDeploy).
		Field("region", providerschema.ValueTypeText).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	providerschema.Registries.MustRegister(testPSRProvider, s)
}

func TestProviderSchemaRepository_GetAndList(t *testing.T) {
	registerTestSchema(t)
	repo := NewProviderSchemaRepository()

	s, err := repo.Get(context.Background(), testPSRProvider)
	if err != nil {
		t.Fatal(err)
	}
	if s.Provider != string(testPSRProvider) {
		t.Fatalf("provider = %q", s.Provider)
	}

	all, err := repo.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, sc := range all {
		if sc.Provider == string(testPSRProvider) {
			found = true
		}
	}
	if !found {
		t.Fatal("List did not include the registered test schema")
	}
}

func TestProviderSchemaRepository_GetUnknownReturns404(t *testing.T) {
	repo := NewProviderSchemaRepository()
	_, err := repo.Get(context.Background(), "__definitely_not_registered__")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	de, ok := err.(*domain.Error)
	if !ok {
		t.Fatalf("expected *domain.Error, got %T", err)
	}
	if de.Code != domain.ErrRecordNotFound.Code {
		t.Fatalf("expected code %d, got %d", domain.ErrRecordNotFound.Code, de.Code)
	}
}
