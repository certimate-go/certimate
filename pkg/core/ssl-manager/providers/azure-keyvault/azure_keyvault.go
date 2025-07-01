package azurekeyvault

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azcertificates"

	"github.com/certimate-go/certimate/pkg/core"
	azenv "github.com/certimate-go/certimate/pkg/sdk3rd/azure/env"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type SSLManagerProviderConfig struct {
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

type SSLManagerProvider struct {
	config    *SSLManagerProviderConfig
	logger    *slog.Logger
	sdkClient *azcertificates.Client
}

var _ core.SSLManager = (*SSLManagerProvider)(nil)

func NewSSLManagerProvider(config *SSLManagerProviderConfig) (*SSLManagerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl manager provider is nil")
	}

	client, err := createSDKClient(config.TenantId, config.ClientId, config.ClientSecret, config.CloudName, config.KeyVaultName)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	return &SSLManagerProvider{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (m *SSLManagerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		m.logger = slog.New(slog.DiscardHandler)
	} else {
		m.logger = logger
	}
}

func (m *SSLManagerProvider) Upload(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLManageUploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 生成 Azure 业务参数
	const TAG_CERTCN = "certimate/cert-cn"
	const TAG_CERTSN = "certimate/cert-sn"
	certCN := certX509.Subject.CommonName
	certSN := certX509.SerialNumber.Text(16)

	// 获取证书列表，避免重复上传
	// REF: https://learn.microsoft.com/en-us/rest/api/keyvault/certificates/get-certificates/get-certificates
	listCertificatesPager := m.sdkClient.NewListCertificatePropertiesPager(nil)
	for listCertificatesPager.More() {
		page, err := listCertificatesPager.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'keyvault.GetCertificates': %w", err)
		}

		for _, certProp := range page.Value {
			// 先对比证书有效期
			if certProp.Attributes == nil {
				continue
			}
			if certProp.Attributes.NotBefore == nil || !certProp.Attributes.NotBefore.Equal(certX509.NotBefore) {
				continue
			}
			if certProp.Attributes.Expires == nil || !certProp.Attributes.Expires.Equal(certX509.NotAfter) {
				continue
			}

			// 再对比 Tag 中的通用名称
			if v, ok := certProp.Tags[TAG_CERTCN]; !ok || v == nil {
				continue
			} else if *v != certCN {
				continue
			}

			// 再对比 Tag 中的序列号
			if v, ok := certProp.Tags[TAG_CERTSN]; !ok || v == nil {
				continue
			} else if *v != certSN {
				continue
			}

			// 最后对比证书内容
			getCertificateResp, err := m.sdkClient.GetCertificate(context.TODO(), certProp.ID.Name(), certProp.ID.Version(), nil)
			m.logger.Debug("sdk request 'keyvault.GetCertificate'", slog.String("request.certificateName", certProp.ID.Name()), slog.String("request.certificateVersion", certProp.ID.Version()), slog.Any("response", getCertificateResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'keyvault.GetCertificate': %w", err)
			} else {
				oldCertX509, err := x509.ParseCertificate(getCertificateResp.CER)
				if err != nil {
					continue
				}

				if !xcert.EqualCertificate(certX509, oldCertX509) {
					continue
				}
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			m.logger.Info("ssl certificate already exists")
			return &core.SSLManageUploadResult{
				CertId:   string(*certProp.ID),
				CertName: certProp.ID.Name(),
			}, nil
		}
	}

	// 生成新证书名（需符合 Azure 命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// Azure Key Vault 不支持导入带有 Certificate Chain 的 PEM 证书。
	// Issue Link: https://github.com/Azure/azure-cli/issues/19017
	// 暂时的解决方法是，将 PEM 证书转换成 PFX 格式，然后再导入。
	certPFX, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, "")
	if err != nil {
		return nil, fmt.Errorf("failed to transform certificate from PEM to PFX: %w", err)
	}

	// 导入证书
	// REF: https://learn.microsoft.com/en-us/rest/api/keyvault/certificates/import-certificate/import-certificate
	importCertificateParams := azcertificates.ImportCertificateParameters{
		Base64EncodedCertificate: to.Ptr(base64.StdEncoding.EncodeToString(certPFX)),
		CertificatePolicy: &azcertificates.CertificatePolicy{
			SecretProperties: &azcertificates.SecretProperties{
				ContentType: to.Ptr("application/x-pkcs12"),
			},
		},
		Tags: map[string]*string{
			TAG_CERTCN: to.Ptr(certCN),
			TAG_CERTSN: to.Ptr(certSN),
		},
	}
	importCertificateResp, err := m.sdkClient.ImportCertificate(context.TODO(), certName, importCertificateParams, nil)
	m.logger.Debug("sdk request 'keyvault.ImportCertificate'", slog.String("request.certificateName", certName), slog.Any("request.parameters", importCertificateParams), slog.Any("response", importCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'keyvault.ImportCertificate': %w", err)
	}

	return &core.SSLManageUploadResult{
		CertId:   string(*importCertificateResp.ID),
		CertName: certName,
	}, nil
}

func createSDKClient(tenantId, clientId, clientSecret, cloudName, keyvaultName string) (*azcertificates.Client, error) {
	env, err := azenv.GetCloudEnvConfiguration(cloudName)
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
	if azenv.IsUSGovernmentEnv(cloudName) {
		endpoint = fmt.Sprintf("https://%s.vault.usgovcloudapi.net", keyvaultName)
	} else if azenv.IsChinaEnv(cloudName) {
		endpoint = fmt.Sprintf("https://%s.vault.azure.cn", keyvaultName)
	}

	client, err := azcertificates.NewClient(endpoint, credential, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}
