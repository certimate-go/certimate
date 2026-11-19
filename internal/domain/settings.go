package domain

import (
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

const CollectionNameSettings = "settings"

type Settings struct {
	Meta
	Name    string          `db:"name"    json:"name"`
	Content SettingsContent `db:"content" json:"content"`
}

const (
	SettingsNameEmails               = "emails"
	SettingsNameNotificationTemplate = "notifyTemplate"
	SettingsNameScriptTemplate       = "scriptTemplate"
	SettingsNameSSLProvider          = "sslProvider"
	SettingsNamePersistence          = "persistence"
	SettingsNamePluginMarket         = "plugin_market"
)

type SettingsContent map[string]any

type SettingsContentForSSLProvider struct {
	Provider CAProviderType                    `json:"provider"`
	Configs  map[CAProviderType]map[string]any `json:"configs"`
	Timeout  int                               `json:"timeout"`
}

type SettingsContentForPersistence struct {
	CertificatesWarningDaysBeforeExpire int `json:"certificatesWarningDaysBeforeExpire"`
	CertificatesRetentionMaxDays        int `json:"certificatesRetentionMaxDays"`
	WorkflowRunsRetentionMaxDays        int `json:"workflowRunsRetentionMaxDays"`
}

func (c SettingsContent) AsSSLProvider() *SettingsContentForSSLProvider {
	content := &SettingsContentForSSLProvider{}
	xmaps.Populate(c, content)

	if content.Provider == "" {
		content.Provider = CAProviderTypeLetsEncrypt
	}

	if content.Timeout < 0 {
		content.Timeout = 0
	}

	return content
}

func (c SettingsContent) AsPersistence() *SettingsContentForPersistence {
	content := &SettingsContentForPersistence{}
	xmaps.Populate(c, content)

	if content.CertificatesWarningDaysBeforeExpire <= 0 {
		content.CertificatesWarningDaysBeforeExpire = 21
	}

	if content.CertificatesRetentionMaxDays < 0 {
		content.CertificatesRetentionMaxDays = 0
	}

	if content.WorkflowRunsRetentionMaxDays < 0 {
		content.WorkflowRunsRetentionMaxDays = 0
	}

	return content
}
