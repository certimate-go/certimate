package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/http01/ftp"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEHttp01Registries.MustRegister(domain.ACMEHttp01ProviderTypeFTP, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForFTP{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := ftp.NewChallenger(&ftp.ChallengerConfig{
			FtpHost:     credentials.Host,
			FtpPort:     credentials.Port,
			FtpUsername: credentials.Username,
			FtpPassword: credentials.Password,
			WebRootPath: xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "webRootPath", "/"),
		})
		return provider, err
	})
}
