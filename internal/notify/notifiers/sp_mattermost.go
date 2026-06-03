package notifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	ntfyimpl "github.com/certimate-go/certimate/pkg/core/notifier/providers/mattermost"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.NotificationProviderTypeMattermost, func(options *ProviderFactoryOptions) (core.Notifier, error) {
		credentials := domain.AccessConfigForMattermost{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := ntfyimpl.NewNotifier(&ntfyimpl.NotifierConfig{
			ServerUrl: credentials.ServerUrl,
			Username:  credentials.Username,
			Password:  credentials.Password,
			ChannelId: xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "channelId", credentials.ChannelId),
		})
		return provider, err
	})
}
