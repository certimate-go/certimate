package notify

import (
	"context"
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/notify/notifiers"
)

type SendNotificationRequest struct {
	// 提供商相关
	Provider               string
	ProviderAccessConfig   map[string]any
	ProviderExtendedConfig map[string]any

	// 通知相关
	Subject string
	Message string
}

type SendNotificationResponse struct{}

func (c *Client) SendNotification(ctx context.Context, request *SendNotificationRequest) (*SendNotificationResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("the request is nil")
	}

	providerFactory, err := notifiers.Registries.Get(domain.NotificationProviderType(request.Provider))
	if err != nil {
		return nil, err
	}

	provider, err := providerFactory(&notifiers.ProviderFactoryOptions{
		ProviderAccessConfig:   request.ProviderAccessConfig,
		ProviderExtendedConfig: request.ProviderExtendedConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize notification provider '%s': %w", request.Provider, err)
	}

	provider.SetLogger(c.logger)
	if _, err := provider.Notify(ctx, request.Subject, request.Message); err != nil {
		return nil, err
	}

	return &SendNotificationResponse{}, nil
}
