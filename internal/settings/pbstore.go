package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
)

func GetGlobalSettingsForSSLProvider() domain.SettingsContentForSSLProvider {
	pb := app.GetApp()
	name := domain.SettingsNameSSLProvider
	content := pb.Store().Get(buildPbStoreKey(name))
	if content == nil {
		content = domain.SettingsContent{}
	}
	return *content.(domain.SettingsContent).AsSSLProvider()
}

func GetGlobalSettingsForPersistence() domain.SettingsContentForPersistence {
	pb := app.GetApp()
	name := domain.SettingsNamePersistence
	content := pb.Store().Get(buildPbStoreKey(name))
	if content == nil {
		content = domain.SettingsContent{}
	}
	return *content.(domain.SettingsContent).AsPersistence()
}

func registerSettingsStoreByName(settingsName string) error {
	settingsRepo := repository.NewSettingsRepository()
	settings, err := settingsRepo.GetByName(context.Background(), settingsName)
	if err != nil {
		return err
	}

	pb := app.GetApp()
	pb.Store().Set(buildPbStoreKey(settingsName), settings.Content)
	return nil
}

func buildPbStoreKey(settingsName string) string {
	return fmt.Sprintf("%s|settings|%s", strings.ToLower(app.AppName), strings.ToLower(settingsName))
}

func SetGlobalSettingsForSSLProvider(content *domain.SettingsContentForSSLProvider) error {
	data, err := json.Marshal(content)
	if err != nil {
		return err
	}

	settingsContent := domain.SettingsContent{}
	if err := json.Unmarshal(data, &settingsContent); err != nil {
		return err
	}

	settingsRepo := repository.NewSettingsRepository()
	settings, err := settingsRepo.GetByName(context.Background(), domain.SettingsNameSSLProvider)
	if err != nil {
		if !domain.IsRecordNotFoundError(err) {
			return err
		}
		settings = &domain.Settings{Name: domain.SettingsNameSSLProvider}
	}

	settings.Content = settingsContent
	if _, err := settingsRepo.Save(context.Background(), settings); err != nil {
		return err
	}

	pb := app.GetApp()
	pb.Store().Set(buildPbStoreKey(domain.SettingsNameSSLProvider), settingsContent)
	return nil
}
