package azurekeyvault

import (
	"context"
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

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
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

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *azcertificates.Client
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.CloudName, config.TenantId, config.ClientId, config.ClientSecret, config.KeyVaultName)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Certmgr{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (c *Certmgr) SetLogger(logger *slog.Logger) {
	if logger == nil {
		c.logger = slog.New(slog.DiscardHandler)
	} else {
		c.logger = logger
	}
}

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*UploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 生成 Azure 业务参数
	certCN := certX509.Subject.CommonName
	certSN := certX509.SerialNumber.Text(16)

	// 获取证书列表，避免重复上传
	// REF: https://learn.microsoft.com/en-us/rest/api/keyvault/certificates/get-certificates/get-certificates
	listCertificatesPager := c.sdkClient.NewListCertificatePropertiesPager(&azcertificates.ListCertificatePropertiesOptions{})
	for listCertificatesPager.More() {
		page, err := listCertificatesPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'keyvault.GetCertificates': %w", err)
		}

		for _, certItem := range page.Value {
			// 对比证书有效期
			if certItem.Attributes == nil {
				continue
			}
			if certItem.Attributes.NotBefore == nil || !certItem.Attributes.NotBefore.Equal(certX509.NotBefore) {
				continue
			}
			if certItem.Attributes.Expires == nil || !certItem.Attributes.Expires.Equal(certX509.NotAfter) {
				continue
			}

			// 对比 Tag 中的通用名称
			if v, ok := certItem.Tags[kvTagCertCN]; !ok || v == nil {
				continue
			} else if *v != certCN {
				continue
			}

			// 对比 Tag 中的序列号
			if v, ok := certItem.Tags[kvTagCertSN]; !ok || v == nil {
				continue
			} else if *v != certSN {
				continue
			}

			// 对比证书内容
			getCertificateResp, err := c.sdkClient.GetCertificate(ctx, certItem.ID.Name(), certItem.ID.Version(), nil)
			c.logger.Debug("sdk request 'keyvault.GetCertificate'", slog.String("request.certificateName", certItem.ID.Name()), slog.String("request.certificateVersion", certItem.ID.Version()), slog.Any("response", getCertificateResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'keyvault.GetCertificate': %w", err)
			} else {
				if !xcert.EqualCertificatesFromPEM(certPEM, string(getCertificateResp.CER)) {
					continue
				}
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &UploadResult{
				CertId:   string(*certItem.ID),
				CertName: certItem.ID.Name(),
			}, nil
		}
	}

	// 生成新证书名（需符合 Azure 命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// Azure Key Vault 不支持导入带有 Certificate Chain 的 PEM 证书。
	// Issue Link: https://github.com/Azure/azure-cli/issues/19017
	// 暂时的解决方法是，将 PEM 证书转换成 PFX 格式，然后再导入。
	certPFX, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, "", nil)
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
			kvTagCertCN: to.Ptr(certCN),
			kvTagCertSN: to.Ptr(certSN),
		},
	}
	importCertificateResp, err := c.sdkClient.ImportCertificate(ctx, certName, importCertificateParams, nil)
	c.logger.Debug("sdk request 'keyvault.ImportCertificate'", slog.String("request.certificateName", certName), slog.Any("request.parameters", importCertificateParams), slog.Any("response", importCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'keyvault.ImportCertificate': %w", err)
	}

	return &UploadResult{
		CertId:   string(*importCertificateResp.ID),
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 转换证书格式
	certPFX, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to transform certificate from PEM to PFX: %w", err)
	}

	// 获取证书
	// REF: https://learn.microsoft.com/en-us/rest/api/keyvault/certificates/get-certificate/get-certificate
	getCertificateResp, err := c.sdkClient.GetCertificate(ctx, certIdOrName, "", nil)
	c.logger.Debug("sdk request 'keyvault.GetCertificate'", slog.String("request.certificateName", certIdOrName), slog.Any("response", getCertificateResp))
	if err != nil {
		var respErr *azcore.ResponseError
		if !errors.As(err, &respErr) || (respErr.ErrorCode != "ResourceNotFound" && respErr.ErrorCode != "CertificateNotFound") {
			return nil, fmt.Errorf("failed to execute sdk request 'keyvault.GetCertificate': %w", err)
		}
	} else {
		// 如果已存在相同证书，直接返回
		if xcert.EqualCertificatesFromPEM(certPEM, string(getCertificateResp.CER)) {
			return &ReplaceResult{}, nil
		}
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
			kvTagCertCN: to.Ptr(certX509.Subject.CommonName),
			kvTagCertSN: to.Ptr(certX509.SerialNumber.Text(16)),
		},
	}
	importCertificateResp, err := c.sdkClient.ImportCertificate(ctx, certIdOrName, importCertificateParams, nil)
	c.logger.Debug("sdk request 'keyvault.ImportCertificate'", slog.String("request.certificateName", certIdOrName), slog.Any("request.parameters", importCertificateParams), slog.Any("response", importCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'keyvault.ImportCertificate': %w", err)
	}

	return &ReplaceResult{}, nil
}

const (
	kvTagCertCN = "certimate/cert-cn"
	kvTagCertSN = "certimate/cert-sn"
)

func createSDKClient(cloudName, tenantId, clientId, clientSecret, keyvaultName string) (*azcertificates.Client, error) {
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
