package notify

import (
	"fmt"
	"net/http"

	"github.com/usual2970/certimate/internal/domain"
	"github.com/usual2970/certimate/internal/pkg/core/notifier"
	pDingTalkBot "github.com/usual2970/certimate/internal/pkg/core/notifier/providers/dingtalkbot"
	pEmail "github.com/usual2970/certimate/internal/pkg/core/notifier/providers/email"
	pLarkBot "github.com/usual2970/certimate/internal/pkg/core/notifier/providers/larkbot"
	pMattermost "github.com/usual2970/certimate/internal/pkg/core/notifier/providers/mattermost"
	pPushover "github.com/usual2970/certimate/internal/pkg/core/notifier/providers/pushover"
	pTelegramBot "github.com/usual2970/certimate/internal/pkg/core/notifier/providers/telegrambot"
	pWebhook "github.com/usual2970/certimate/internal/pkg/core/notifier/providers/webhook"
	pWeComBot "github.com/usual2970/certimate/internal/pkg/core/notifier/providers/wecombot"
	httputil "github.com/usual2970/certimate/internal/pkg/utils/http"
	maputil "github.com/usual2970/certimate/internal/pkg/utils/map"
)

type notifierProviderOptions struct {
	Provider               domain.NotificationProviderType
	ProviderAccessConfig   map[string]any
	ProviderExtendedConfig map[string]any
}

func createNotifierProvider(options *notifierProviderOptions) (notifier.Notifier, error) {
	/*
	  注意：如果追加新的常量值，请保持以 ASCII 排序。
	  NOTICE: If you add new constant, please keep ASCII order.
	*/
	switch options.Provider {
	case domain.NotificationProviderTypeDingTalkBot:
		{
			access := domain.AccessConfigForDingTalkBot{}
			if err := maputil.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			return pDingTalkBot.NewNotifier(&pDingTalkBot.NotifierConfig{
				WebhookUrl: access.WebhookUrl,
				Secret:     access.Secret,
			})
		}

	case domain.NotificationProviderTypeEmail:
		{
			access := domain.AccessConfigForEmail{}
			if err := maputil.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			return pEmail.NewNotifier(&pEmail.NotifierConfig{
				SmtpHost:        access.SmtpHost,
				SmtpPort:        access.SmtpPort,
				SmtpTls:         access.SmtpTls,
				Username:        access.Username,
				Password:        access.Password,
				SenderAddress:   maputil.GetOrDefaultString(options.ProviderExtendedConfig, "senderAddress", access.DefaultSenderAddress),
				ReceiverAddress: maputil.GetOrDefaultString(options.ProviderExtendedConfig, "receiverAddress", access.DefaultReceiverAddress),
			})
		}

	case domain.NotificationProviderTypeLarkBot:
		{
			access := domain.AccessConfigForLarkBot{}
			if err := maputil.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			return pLarkBot.NewNotifier(&pLarkBot.NotifierConfig{
				WebhookUrl: access.WebhookUrl,
			})
		}

	case domain.NotificationProviderTypeMattermost:
		{
			access := domain.AccessConfigForMattermost{}
			if err := maputil.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			return pMattermost.NewNotifier(&pMattermost.NotifierConfig{
				ServerUrl: access.ServerUrl,
				Username:  access.Username,
				Password:  access.Password,
				ChannelId: maputil.GetOrDefaultString(options.ProviderExtendedConfig, "channelId", access.DefaultChannelId),
			})
		}

	case domain.NotificationProviderTypePushover:
		{
			access := domain.AccessConfigForPushover{}
			if err := maputil.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			return pPushover.NewNotifier(&pPushover.NotifierConfig{
				Token:    access.Token,
				User:     access.User,
				Priority: maputil.GetOrDefaultString(options.ProviderExtendedConfig, "priority", "0"),
			})
		}

	case domain.NotificationProviderTypeTelegramBot:
		{
			access := domain.AccessConfigForTelegramBot{}
			if err := maputil.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			return pTelegramBot.NewNotifier(&pTelegramBot.NotifierConfig{
				BotToken: access.BotToken,
				ChatId:   maputil.GetOrDefaultInt64(options.ProviderExtendedConfig, "chatId", access.DefaultChatId),
			})
		}

	case domain.NotificationProviderTypeWebhook:
		{
			access := domain.AccessConfigForWebhook{}
			if err := maputil.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			mergedHeaders := make(map[string]string)
			if defaultHeadersString := access.HeadersString; defaultHeadersString != "" {
				h, err := httputil.ParseHeaders(defaultHeadersString)
				if err != nil {
					return nil, fmt.Errorf("failed to parse webhook headers: %w", err)
				}
				for key := range h {
					mergedHeaders[http.CanonicalHeaderKey(key)] = h.Get(key)
				}
			}
			if extendedHeadersString := maputil.GetString(options.ProviderExtendedConfig, "headers"); extendedHeadersString != "" {
				h, err := httputil.ParseHeaders(extendedHeadersString)
				if err != nil {
					return nil, fmt.Errorf("failed to parse webhook headers: %w", err)
				}
				for key := range h {
					mergedHeaders[http.CanonicalHeaderKey(key)] = h.Get(key)
				}
			}

			return pWebhook.NewNotifier(&pWebhook.NotifierConfig{
				WebhookUrl:               access.Url,
				WebhookData:              maputil.GetOrDefaultString(options.ProviderExtendedConfig, "webhookData", access.DefaultDataForNotification),
				Method:                   access.Method,
				Headers:                  mergedHeaders,
				AllowInsecureConnections: access.AllowInsecureConnections,
			})
		}

	case domain.NotificationProviderTypeWeComBot:
		{
			access := domain.AccessConfigForWeComBot{}
			if err := maputil.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			return pWeComBot.NewNotifier(&pWeComBot.NotifierConfig{
				WebhookUrl: access.WebhookUrl,
			})
		}
	}

	return nil, fmt.Errorf("unsupported notifier provider '%s'", options.Provider)
}
