package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
)

type AccessRepository struct{}

func NewAccessRepository() *AccessRepository {
	return &AccessRepository{}
}

func (r *AccessRepository) GetById(ctx context.Context, id string) (*domain.Access, error) {
	record, err := app.GetApp().FindRecordById(domain.CollectionNameAccess, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	if !record.GetDateTime("deleted").Time().IsZero() {
		return nil, domain.ErrRecordNotFound
	}

	return r.castRecordToModel(record)
}

func (r *AccessRepository) castRecordToModel(record *core.Record) (*domain.Access, error) {
	if record == nil {
		return nil, fmt.Errorf("the record is nil")
	}

	config := make(map[string]any)
	if err := record.UnmarshalJSONField("config", &config); err != nil {
		return nil, fmt.Errorf("field 'config' is malformed")
	}

	access := &domain.Access{
		Meta: domain.Meta{
			Id:        record.Id,
			CreatedAt: record.GetDateTime("created").Time(),
			UpdatedAt: record.GetDateTime("updated").Time(),
		},
		Name:     record.GetString("name"),
		Provider: record.GetString("provider"),
		Config:   config,
		Reserve:  record.GetString("reserve"),
	}
	return access, nil
}
