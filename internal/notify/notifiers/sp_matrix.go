package notifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/notifier/providers/matrix"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.NotificationProviderTypeMatrix, func(options *ProviderFactoryOptions) (core.Notifier, error) {
		credentials := domain.AccessConfigForMatrix{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := matrix.NewNotifier(&matrix.NotifierConfig{
			ServerUrl:   credentials.ServerUrl,
			UserId:      credentials.UserId,
			AccessToken: credentials.AccessToken,
			RoomId:      xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "roomId", credentials.RoomId),
		})
		return provider, err
	})
}
