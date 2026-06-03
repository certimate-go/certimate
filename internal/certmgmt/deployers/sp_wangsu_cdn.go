package deployers

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/wangsu-cdn"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeWangsuCDN, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForWangsu{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			AccessKeyId:        credentials.AccessKeyId,
			AccessKeySecret:    credentials.AccessKeySecret,
			DomainMatchPattern: xmaps.GetString(options.ProviderExtendedConfig, "domainMatchPattern"),
			Domains:            lo.Filter(strings.Split(xmaps.GetString(options.ProviderExtendedConfig, "domains"), ";"), func(s string, _ int) bool { return s != "" }),
		})
		return provider, err
	})
}
