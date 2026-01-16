package notifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/notifier"
	"github.com/certimate-go/certimate/pkg/core/notifier/providers/wecombot"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.NotificationProviderTypeWeComBot, func(options *ProviderFactoryOptions) (notifier.Provider, error) {
		credentials := domain.AccessConfigForWeComBot{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := wecombot.NewNotifier(&wecombot.NotifierConfig{
			WebhookUrl:    credentials.WebhookUrl,
			CustomPayload: credentials.CustomPayload,
		})
		return provider, err
	})
}
