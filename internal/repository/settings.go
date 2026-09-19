package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
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

func (r *SettingsRepository) Save(ctx context.Context, settings *domain.Settings) (*domain.Settings, error) {
	collection, err := app.GetApp().FindCollectionByNameOrId(domain.CollectionNameSettings)
	if err != nil {
		return settings, err
	}

	var record *core.Record
	if existing, err := r.GetByName(ctx, settings.Name); err != nil {
		if !domain.IsRecordNotFoundError(err) {
			return settings, err
		}
	} else {
		record, err = app.GetApp().FindRecordById(collection, existing.Id)
		if err != nil {
			return settings, err
		}
	}
	if record == nil {
		record = core.NewRecord(collection)
	}

	record.Set("name", settings.Name)
	record.Set("content", settings.Content)
	if err := app.GetApp().Save(record); err != nil {
		return settings, err
	}

	settings.Id = record.Id
	settings.CreatedAt = record.GetDateTime("created").Time()
	settings.UpdatedAt = record.GetDateTime("updated").Time()
	return settings, nil
}
