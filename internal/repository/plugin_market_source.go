package repository

import (
	"context"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"github.com/certimate-go/certimate/pkg/plugin"
)

type PluginMarketSources struct{ app core.App }

func NewPluginMarketSources(app core.App) *PluginMarketSources { return &PluginMarketSources{app: app} }

func (r *PluginMarketSources) List(ctx context.Context) ([]plugin.MarketSource, error) {
	var rows []struct {
		ID         string `db:"id"`
		Repository string `db:"repository"`
	}
	err := r.app.DB().Select("id", "repository").From("plugin_market_sources").OrderBy("id").WithContext(ctx).All(&rows)
	result := make([]plugin.MarketSource, 0, len(rows))
	for _, row := range rows {
		result = append(result, plugin.MarketSource{ID: row.ID, Repository: row.Repository})
	}
	return result, err
}

func (r *PluginMarketSources) Add(ctx context.Context, source plugin.MarketSource) error {
	_, err := r.app.DB().Insert("plugin_market_sources", dbx.Params{"id": source.ID, "repository": source.Repository}).WithContext(ctx).Execute()
	return err
}

func (r *PluginMarketSources) Remove(ctx context.Context, id string) error {
	_, err := r.app.DB().Delete("plugin_market_sources", dbx.HashExp{"id": id}).WithContext(ctx).Execute()
	return err
}
