package repository

import (
	"context"
	"fmt"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
)

// MergeMatrixSessionIntoAccess stores Matrix session token and device id in access config.
func (r *AccessRepository) MergeMatrixSessionIntoAccess(ctx context.Context, accessID, sessionToken, deviceID string) error {
	if accessID == "" || sessionToken == "" {
		return nil
	}

	record, err := app.GetApp().FindRecordById(domain.CollectionNameAccess, accessID)
	if err != nil {
		return fmt.Errorf("find access %s: %w", accessID, err)
	}

	config := make(map[string]any)
	if err := record.UnmarshalJSONField("config", &config); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	config["sessionAccessToken"] = sessionToken
	if deviceID != "" {
		config["sessionDeviceId"] = deviceID
	}

	record.Set("config", config)
	if err := app.GetApp().Save(record); err != nil {
		return fmt.Errorf("save access %s: %w", accessID, err)
	}
	return nil
}
