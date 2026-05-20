package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	aliyunesa "github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/aliyun-esa"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEDns01Registries.MustRegister(domain.ACMEDns01ProviderTypeAliyunESA, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForAliyun{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := aliyunesa.NewChallenger(&aliyunesa.ChallengerConfig{
			AccessKeyId:           credentials.AccessKeyId,
			AccessKeySecret:       credentials.AccessKeySecret,
			Region:                xmaps.GetString(options.ProviderExtendedConfig, "region"),
			DnsPropagationTimeout: options.DnsPropagationTimeout,
			DnsTTL:                options.DnsTTL,
		})
		return provider, err
	})
}
