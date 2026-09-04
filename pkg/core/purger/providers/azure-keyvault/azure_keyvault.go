package azurekeyvault

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azcertificates"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xazure "github.com/certimate-go/certimate/pkg/utils/third-party/azure"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider    = core.Purger
	PurgeResult = core.PurgerPurgeResult
)

type PurgerConfig struct {
	// Azure TenantId。
	TenantId string `json:"tenantId"`
	// Azure ClientId。
	ClientId string `json:"clientId"`
	// Azure ClientSecret。
	ClientSecret string `json:"clientSecret"`
	// Azure 主权云环境。
	CloudName string `json:"cloudName,omitempty"`
	// Key Vault 名称。
	KeyVaultName string `json:"keyvaultName"`
}

type Purger struct {
	config    *PurgerConfig
	logger    *slog.Logger
	sdkClient *azcertificates.Client
}

var _ Provider = (*Purger)(nil)

func NewPurger(config *PurgerConfig) (*Purger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the purger provider is nil")
	}

	client, err := createSDKClient(config.CloudName, config.TenantId, config.ClientId, config.ClientSecret, config.KeyVaultName)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Purger{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (p *Purger) SetLogger(logger *slog.Logger) {
	if logger == nil {
		p.logger = slog.New(slog.DiscardHandler)
	} else {
		p.logger = logger
	}
}

func (p *Purger) Purge(ctx context.Context, expiry time.Duration) (*PurgeResult, error) {
	purgingCertIds := make([]azcertificates.ID, 0)

	// 获取证书列表
	// REF: https://learn.microsoft.com/en-us/rest/api/keyvault/certificates/get-certificates/get-certificates
	listCertificatesPager := p.sdkClient.NewListCertificatePropertiesPager(&azcertificates.ListCertificatePropertiesOptions{})
	for listCertificatesPager.More() {
		page, err := listCertificatesPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'keyvault.GetCertificates': %w", err)
		}

		for _, certItem := range page.Value {
			if certItem.Attributes == nil || certItem.Attributes.Expires == nil {
				continue
			}

			certNotAfter := lo.FromPtr(certItem.Attributes.Expires)
			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, *certItem.ID)
			}
		}
	}

	// 删除证书
	// REF: https://learn.microsoft.com/en-us/rest/api/keyvault/certificates/delete-certificate/delete-certificate
	purgedCount := 0
	if len(purgingCertIds) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certIds", purgingCertIds))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertIds, func(ctx context.Context, certId azcertificates.ID, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			deleteCertificateResp, err := p.sdkClient.DeleteCertificate(ctx, certId.Name(), nil)
			p.logger.Debug("sdk request 'keyvault.DeleteCertificate'", slog.String("params.certificateName", certId.Name()), slog.Any("response", deleteCertificateResp))
			if err != nil {
				var sdkErr *azcore.ResponseError
				if errors.As(err, &sdkErr) {
					if sdkErrCode := sdkErr.ErrorCode; sdkErrCode == "ResourceNotFound" || sdkErrCode == "CertificateNotFound" {
						return nil
					}
				}

				return fmt.Errorf("failed to execute sdk request 'keyvault.DeleteCertificate': %w", err)
			}

			purgedCount++
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return &PurgeResult{
		Count: purgedCount,
	}, nil
}

func createSDKClient(cloudName, tenantId, clientId, clientSecret, keyvaultName string) (*azcertificates.Client, error) {
	env, err := xazure.GetCloudEnvConfiguration(cloudName)
	if err != nil {
		return nil, err
	}

	clientOptions := azcore.ClientOptions{Cloud: env}
	credential, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret,
		&azidentity.ClientSecretCredentialOptions{ClientOptions: clientOptions})
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("https://%s.vault.azure.net", keyvaultName)
	if xazure.IsUSGovernmentEnv(cloudName) {
		endpoint = fmt.Sprintf("https://%s.vault.usgovcloudapi.net", keyvaultName)
	} else if xazure.IsChinaEnv(cloudName) {
		endpoint = fmt.Sprintf("https://%s.vault.azure.cn", keyvaultName)
	}

	client, err := azcertificates.NewClient(endpoint, credential, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}
