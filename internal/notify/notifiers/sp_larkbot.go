package notifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	ntfyimpl "github.com/certimate-go/certimate/pkg/core/notifier/providers/larkbot"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.NotificationProviderTypeLarkBot, func(options *ProviderFactoryOptions) (core.Notifier, error) {
		credentials := domain.AccessConfigForLarkBot{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := ntfyimpl.NewNotifier(&ntfyimpl.NotifierConfig{
			WebhookUrl:    credentials.WebhookUrl,
			Secret:        credentials.Secret,
			CustomPayload: credentials.CustomPayload,
		})
		return provider, err
	})
}
