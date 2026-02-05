package notifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/notifier"
	"github.com/certimate-go/certimate/pkg/core/notifier/providers/email"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.NotificationProviderTypeEmail, func(options *ProviderFactoryOptions) (notifier.Provider, error) {
		credentials := domain.AccessConfigForEmail{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := email.NewNotifier(&email.NotifierConfig{
			SmtpHost:                 credentials.SmtpHost,
			SmtpPort:                 credentials.SmtpPort,
			SmtpTls:                  credentials.SmtpTls,
			Username:                 credentials.Username,
			Password:                 credentials.Password,
			SenderAddress:            credentials.SenderAddress,
			SenderName:               credentials.SenderName,
			ReceiverAddress:          xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "receiverAddress", credentials.ReceiverAddress),
			MessageFormat:            xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "format", email.MESSAGE_FORMAT_PLAIN),
			AllowInsecureConnections: credentials.AllowInsecureConnections,
		})
		return provider, err
	})
}
