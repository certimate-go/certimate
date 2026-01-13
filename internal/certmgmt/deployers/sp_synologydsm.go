package deployers

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/synologydsm"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeSynologyDSM, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForSynologyDSM{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		// Parse serverUrl to extract hostname, port, and scheme
		parsedURL, err := url.Parse(credentials.ServerUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to parse server URL: %w", err)
		}

		hostname := parsedURL.Hostname()
		scheme := parsedURL.Scheme
		if scheme == "" {
			scheme = "http"
		}

		var port int32
		if parsedURL.Port() != "" {
			portNum, err := strconv.Atoi(parsedURL.Port())
			if err != nil {
				return nil, fmt.Errorf("failed to parse port number: %w", err)
			}
			port = int32(portNum)
		}

		provider, err := synologydsm.NewDeployer(&synologydsm.DeployerConfig{
			Hostname:                 hostname,
			Port:                     port,
			Scheme:                   scheme,
			Username:                 credentials.Username,
			Password:                 credentials.Password,
			TotpSecret:               credentials.TotpSecret,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			CertificateID:            xmaps.GetString(options.ProviderExtendedConfig, "certificateId"),
			CertificateName:          xmaps.GetString(options.ProviderExtendedConfig, "certificateName"),
			IsDefault:                xmaps.GetBool(options.ProviderExtendedConfig, "isDefault"),
		})
		return provider, err
	})
}

