package certifiers

import (
	"fmt"

	"github.com/go-acme/lego/v4/challenge"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/http01/s3"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEHttp01Registries.MustRegister(domain.ACMEHttp01ProviderTypeS3, func(options *ProviderFactoryOptions) (challenge.Provider, error) {
		credentials := domain.AccessConfigForS3{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := s3.NewChallenger(&s3.ChallengerConfig{
			Endpoint:                 credentials.Endpoint,
			AccessKey:                credentials.AccessKey,
			SecretKey:                credentials.SecretKey,
			SignatureVersion:         credentials.SignatureVersion,
			UsePathStyle:             credentials.UsePathStyle,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			Region:                   xmaps.GetString(options.ProviderExtendedConfig, "region"),
			Bucket:                   xmaps.GetString(options.ProviderExtendedConfig, "bucket"),
		})
		return provider, err
	})
}
