package settings

import (
	"github.com/certimate-go/certimate/internal/domain"
)

func Setup() {
	initPbSettings()

	registerSettingsStoreByName(domain.SettingsNameSSLProvider)
	registerSettingsStoreByName(domain.SettingsNamePersistence)
	registerSettingsRecordEvents()
}
