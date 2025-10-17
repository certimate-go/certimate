package domain

import (
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

const CollectionNameSettings = "settings"

type Settings struct {
	Meta
	Name    string          `json:"name" db:"name"`
	Content SettingsContent `json:"content" db:"content"`
}

type SettingsContent map[string]any

type SettingsContentForSSLProvider struct {
	Provider CAProviderType                    `json:"provider"`
	Config   map[CAProviderType]map[string]any `json:"config"`
}

type SettingsContentForPersistence struct {
	WorkflowRunsMaxDaysRetention        int `json:"workflowRunsMaxDaysRetention"`
	ExpiredCertificatesMaxDaysRetention int `json:"expiredCertificatesMaxDaysRetention"`
}

func (c SettingsContent) AsSSLProvider() *SettingsContentForSSLProvider {
	content := &SettingsContentForSSLProvider{}
	xmaps.Populate(c, content)

	if content.Provider == "" {
		content.Provider = CAProviderTypeLetsEncrypt
	}

	return content
}

func (c SettingsContent) AsPersistence() *SettingsContentForPersistence {
	content := &SettingsContentForPersistence{}
	xmaps.Populate(c, content)
	return content
}
