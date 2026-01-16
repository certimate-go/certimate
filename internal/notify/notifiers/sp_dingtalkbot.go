package notifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/notifier"
	"github.com/certimate-go/certimate/pkg/core/notifier/providers/dingtalkbot"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.NotificationProviderTypeDingTalkBot, func(options *ProviderFactoryOptions) (notifier.Provider, error) {
		credentials := domain.AccessConfigForDingTalkBot{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dingtalkbot.NewNotifier(&dingtalkbot.NotifierConfig{
			WebhookUrl:    credentials.WebhookUrl,
			Secret:        credentials.Secret,
			CustomPayload: credentials.CustomPayload,
		})
		return provider, err
	})
}
