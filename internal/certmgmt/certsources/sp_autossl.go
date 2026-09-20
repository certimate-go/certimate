package certsources

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	csrcimpl "github.com/certimate-go/certimate/pkg/core/certsource/providers/autossl"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.CertsourceProviderTypeAutoSSL, func(options *ProviderFactoryOptions) (core.Certsource, error) {
		credentials := domain.AccessConfigForAutoSSL{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := csrcimpl.NewCertsource(&csrcimpl.CertsourceConfig{
			AccessKey: credentials.AccessKey,
			SecretKey: credentials.SecretKey,
		})
		return provider, err
	})
}
