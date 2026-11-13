package repository

import (
	"context"
	"encoding/json"
	"time"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/pkg/plugin"
)

type PluginMarketSync struct {
	SyncedAt     time.Time
	ETag         string
	LastModified string
}

type PluginMarketCache struct {
	app core.App
}

func NewPluginMarketCache(app core.App) *PluginMarketCache {
	return &PluginMarketCache{app: app}
}

func (r *PluginMarketCache) Load(ctx context.Context, source string) ([]*plugin.MarketManifest, PluginMarketSync, error) {
	var rows []struct {
		ProviderType string `db:"provider_type"`
		Payload      string `db:"payload"`
	}
	err := r.app.DB().Select("provider_type", "payload").From("plugin_market_cache").
		Where(dbx.HashExp{"source": source}).OrderBy("provider_type").WithContext(ctx).All(&rows)
	var sync PluginMarketSync
	if err != nil || len(rows) == 0 {
		return nil, sync, err
	}
	entries := make([]*plugin.MarketManifest, 0, len(rows)-1)
	for _, row := range rows {
		if row.ProviderType == "" {
			if err := json.Unmarshal([]byte(row.Payload), &sync); err != nil {
				return nil, sync, err
			}
			continue
		}
		var manifest plugin.MarketManifest
		if err := json.Unmarshal([]byte(row.Payload), &manifest); err != nil {
			return nil, sync, err
		}
		entries = append(entries, &manifest)
	}
	return entries, sync, nil
}

func (r *PluginMarketCache) Save(ctx context.Context, source string, entries []*plugin.MarketManifest, sync PluginMarketSync) error {
	return r.app.RunInTransaction(func(tx core.App) error {
		if entries != nil {
			if _, err := tx.DB().Delete("plugin_market_cache", dbx.HashExp{"source": source}).WithContext(ctx).Execute(); err != nil {
				return err
			}
			for _, entry := range entries {
				payload, err := json.Marshal(entry)
				if err != nil {
					return err
				}
				if _, err := tx.DB().Insert("plugin_market_cache", dbx.Params{
					"source": source, "provider_type": entry.ProviderType, "payload": string(payload),
				}).WithContext(ctx).Execute(); err != nil {
					return err
				}
			}
		}
		payload, err := json.Marshal(sync)
		if err != nil {
			return err
		}
		_, err = tx.DB().NewQuery(`INSERT INTO plugin_market_cache (source, provider_type, payload)
			VALUES ({:source}, '', {:payload}) ON CONFLICT (source, provider_type) DO UPDATE SET payload = excluded.payload`).
			Bind(dbx.Params{"source": source, "payload": string(payload)}).WithContext(ctx).Execute()
		return err
	})
}
