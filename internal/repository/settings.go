package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/pocketbase/dbx"
)

type SettingsRepository struct{}

func NewSettingsRepository() *SettingsRepository {
	return &SettingsRepository{}
}

func (r *SettingsRepository) GetByName(ctx context.Context, name string) (*domain.Settings, error) {
	record, err := app.GetApp().FindFirstRecordByFilter(
		domain.CollectionNameSettings,
		"name={:name}",
		dbx.Params{"name": name},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	content := make(map[string]any)
	if err := record.UnmarshalJSONField("content", &content); err != nil {
		return nil, fmt.Errorf("field 'content' is malformed")
	}

	settings := &domain.Settings{
		Meta: domain.Meta{
			Id:        record.Id,
			CreatedAt: record.GetDateTime("created").Time(),
			UpdatedAt: record.GetDateTime("updated").Time(),
		},
		Name:    record.GetString("name"),
		Content: content,
	}
	return settings, nil
}
