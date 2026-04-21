package deployers

import (
	"fmt"
	"strings"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	k8ssecret "github.com/certimate-go/certimate/pkg/core/deployer/providers/k8s-secret"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeKubernetesSecret, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForKubernetes{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		parseKeyValueMap := func(s string) (map[string]string, error) {
			result := make(map[string]string)

			lines := strings.Split(s, "\n")
			for i, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}

				pos := strings.Index(line, ":")
				if pos == -1 {
					return nil, fmt.Errorf("invalid line format at line %d", i+1)
				}

				key := strings.TrimSpace(line[:pos])
				value := strings.TrimSpace(line[pos+1:])
				if key == "" {
					return nil, fmt.Errorf("invalid key at line %d", i+1)
				}

				result[key] = value
			}

			return result, nil
		}

		secretAnnotations := make(map[string]string)
		if secretAnnotationsString := xmaps.GetString(options.ProviderExtendedConfig, "secretAnnotations"); secretAnnotationsString != "" {
			temp, err := parseKeyValueMap(secretAnnotationsString)
			if err != nil {
				return nil, fmt.Errorf("failed to parse kubernetes secret annotations: %w", err)
			}
			secretAnnotations = temp
		}

		secretLabels := make(map[string]string)
		if secretLabelsString := xmaps.GetString(options.ProviderExtendedConfig, "secretLabels"); secretLabelsString != "" {
			temp, err := parseKeyValueMap(secretLabelsString)
			if err != nil {
				return nil, fmt.Errorf("failed to parse kubernetes secret labels: %w", err)
			}
			secretLabels = temp
		}

		provider, err := k8ssecret.NewDeployer(&k8ssecret.DeployerConfig{
			KubeConfig:             credentials.KubeConfig,
			Namespace:              xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "namespace", "default"),
			SecretName:             xmaps.GetString(options.ProviderExtendedConfig, "secretName"),
			SecretType:             xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "secretType", "kubernetes.io/tls"),
			SecretDataKeyForCrt:    xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "secretDataKeyForCrt", "tls.crt"),
			SecretDataKeyForKey:    xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "secretDataKeyForKey", "tls.key"),
			SecretDataKeyForIssuer: xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "secretDataKeyForIssuer", "ca.crt"),
			SecretAnnotations:      secretAnnotations,
			SecretLabels:           secretLabels,
		})
		return provider, err
	})
}
