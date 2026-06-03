package certifiers

import (
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	chlgimpl "github.com/certimate-go/certimate/pkg/core/certifier/challengers/http01/local"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEHttp01Registries.MustRegister(domain.ACMEHttp01ProviderTypeLocal, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		provider, err := chlgimpl.NewChallenger(&chlgimpl.ChallengerConfig{
			WebRootPath: xmaps.GetString(options.ProviderExtendedConfig, "webRootPath"),
		})
		return provider, err
	})
}
