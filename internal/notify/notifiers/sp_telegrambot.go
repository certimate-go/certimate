package notifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	ntfyimpl "github.com/certimate-go/certimate/pkg/core/notifier/providers/telegrambot"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.NotificationProviderTypeTelegramBot, func(options *ProviderFactoryOptions) (core.Notifier, error) {
		credentials := domain.AccessConfigForTelegramBot{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := ntfyimpl.NewNotifier(&ntfyimpl.NotifierConfig{
			BotToken: credentials.BotToken,
			ChatId:   xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "chatId", credentials.ChatId),
		})
		return provider, err
	})
}
