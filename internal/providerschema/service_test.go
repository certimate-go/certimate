package providerschema

import (
	"context"
	"errors"
	"testing"

	"github.com/certimate-go/certimate/internal/domain"
)

type fakeRepo struct {
	schemas map[string]*Schema
	err     error
}

func (f *fakeRepo) Get(_ context.Context, providerType string) (*Schema, error) {
	if f.err != nil {
		return nil, f.err
	}
	if s, ok := f.schemas[providerType]; ok {
		return s, nil
	}
	return nil, domain.ErrRecordNotFound
}

func (f *fakeRepo) List(_ context.Context) ([]*Schema, error) {
	if f.err != nil {
		return nil, f.err
	}
	out := make([]*Schema, 0, len(f.schemas))
	for _, s := range f.schemas {
		out = append(out, s)
	}
	return out, nil
}

func (f *fakeRepo) GetEnvelope(_ context.Context, _ string) (*Envelope, bool, error) {
	return nil, false, nil
}

func (f *fakeRepo) ListEnvelopes(_ context.Context) ([]*Envelope, error) {
	return nil, nil
}

type envelopeRepo struct {
	fakeRepo
	envelopes map[string]*Envelope
}

func (e *envelopeRepo) GetEnvelope(_ context.Context, providerType string) (*Envelope, bool, error) {
	env, ok := e.envelopes[providerType]
	return env, ok, nil
}

func (e *envelopeRepo) ListEnvelopes(_ context.Context) ([]*Envelope, error) {
	out := make([]*Envelope, 0, len(e.envelopes))
	for _, env := range e.envelopes {
		out = append(out, env)
	}
	return out, nil
}

func TestService_GetByProviderType_PrefersEnvelope(t *testing.T) {
	builtIn, _ := New("plugin-demo", CategoryDeploy).Field("a", ValueTypeText).Build()
	env := &Envelope{SchemaVersion: "form/v1", Provider: "plugin-demo", Category: CategoryDeploy, Schema: EnvelopeSchema{Columns: []Column{{Name: "envelopeField"}}}}
	repo := &envelopeRepo{
		fakeRepo:  fakeRepo{schemas: map[string]*Schema{"plugin-demo": builtIn}},
		envelopes: map[string]*Envelope{"plugin-demo": env},
	}
	svc := NewProviderSchemaService(repo)

	got, err := svc.GetByProviderType(context.Background(), "plugin-demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Schema.Columns) != 1 || got.Schema.Columns[0].Name != "envelopeField" {
		t.Fatalf("expected envelope to win, got %+v", got.Schema.Columns)
	}
}

func TestService_GetByProviderType_Emits(t *testing.T) {
	s, _ := New("aliyun-cdn", CategoryDeploy).Field("region", ValueTypeText).Build()
	repo := &fakeRepo{schemas: map[string]*Schema{"aliyun-cdn": s}}
	svc := NewProviderSchemaService(repo)

	env, err := svc.GetByProviderType(context.Background(), "aliyun-cdn")
	if err != nil {
		t.Fatal(err)
	}
	if env.Provider != "aliyun-cdn" || env.SchemaVersion != "form/v1" {
		t.Fatalf("envelope not emitted: %+v", env)
	}
	if len(env.Schema.Columns) != 1 || env.Schema.Columns[0].Name != "region" {
		t.Fatalf("column not emitted: %+v", env.Schema.Columns)
	}
}

func TestService_GetByProviderType_NotFound(t *testing.T) {
	svc := NewProviderSchemaService(&fakeRepo{schemas: map[string]*Schema{}})
	_, err := svc.GetByProviderType(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected not-found error")
	}
	if !errors.Is(err, domain.ErrRecordNotFound) {
		var de *domain.Error
		if !errors.As(err, &de) || de.Code != domain.ErrRecordNotFound.Code {
			t.Fatalf("expected 404 domain error, got %v", err)
		}
	}
}

func TestService_List_EmitsAll(t *testing.T) {
	a, _ := New("aliyun-cdn", CategoryDeploy).Field("region", ValueTypeText).Build()
	b, _ := New("tencentcloud-cdn", CategoryDeploy).Field("region", ValueTypeText).Build()
	repo := &fakeRepo{schemas: map[string]*Schema{
		"aliyun-cdn":       a,
		"tencentcloud-cdn": b,
	}}
	svc := NewProviderSchemaService(repo)

	envs, err := svc.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(envs) != 2 {
		t.Fatalf("envelope count = %d, want 2", len(envs))
	}
	for _, env := range envs {
		if env.SchemaVersion != "form/v1" {
			t.Fatalf("envelope schemaVersion = %q", env.SchemaVersion)
		}
	}
}

func TestService_PropagatesRepoError(t *testing.T) {
	boom := errors.New("repo boom")
	svc := NewProviderSchemaService(&fakeRepo{err: boom})
	if _, err := svc.GetByProviderType(context.Background(), "x"); err != boom {
		t.Fatalf("expected propagated error, got %v", err)
	}
	if _, err := svc.List(context.Background()); err != boom {
		t.Fatalf("expected propagated error, got %v", err)
	}
}
