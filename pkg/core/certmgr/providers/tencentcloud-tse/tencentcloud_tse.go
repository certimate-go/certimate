package tencentcloudtse

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	tcssl "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
	tctse "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tse/v20201207"

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云项目 ID。
	ProjectId int64 `json:"projectId,omitempty"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 腾讯云地域。
	Region string `json:"region"`
	// 服务类型。
	ServiceType string `json:"serviceType"`
	// 云原生网关 ID。
	// 服务类型为 [SERVICE_TYPE_CLOUDNATIVE] 时必填。
	GatewayId string `json:"gatewayId,omitempty"`
	// 云原生网关绑定的域名。
	// 服务类型为 [SERVICE_TYPE_CLOUDNATIVE] 时选填。
	// 零值时根据证书内容自动识别。
	Domains []string `json:"domains,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *wSDKClients
}

var _ Provider = (*Certmgr)(nil)

type wSDKClients struct {
	SSL *tcssl.Client
	TSE *tctse.Client
}

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClients(config.SecretId, config.SecretKey, config.Endpoint, config.Region)
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
	switch c.config.ServiceType {
	case SERVICE_TYPE_CLOUDNATIVE:
		return c.uploadToCloudNative(ctx, certPEM, privkeyPEM)

	default:
		return nil, fmt.Errorf("unsupported service type '%s'", c.config.ServiceType)
	}
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	switch c.config.ServiceType {
	case SERVICE_TYPE_CLOUDNATIVE:
		return c.replaceToCloudNative(ctx, certIdOrName, certPEM, privkeyPEM)

	default:
		return nil, fmt.Errorf("unsupported service type '%s'", c.config.ServiceType)
	}
}

func (c *Certmgr) uploadToCloudNative(ctx context.Context, certPEM, privkeyPEM string) (*UploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 先上传证书到 SSL
	// REF: https://cloud.tencent.com/document/api/400/41665
	uploadCertificateReq := tcssl.NewUploadCertificateRequest()
	uploadCertificateReq.ProjectId = lo.EmptyableToPtr(uint64(c.config.ProjectId))
	uploadCertificateReq.CertificatePublicKey = common.StringPtr(certPEM)
	uploadCertificateReq.CertificatePrivateKey = common.StringPtr(privkeyPEM)
	uploadCertificateReq.Repeatable = common.BoolPtr(false)
	uploadCertificateResp, err := c.sdkClient.SSL.UploadCertificateWithContext(ctx, uploadCertificateReq)
	c.logger.Debug("sdk request 'ssl.UploadCertificate'", slog.Any("request", uploadCertificateReq), slog.Any("response", uploadCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'ssl.UploadCertificate': %w", err)
	}

	// 查询云原生网关证书列表，避免重复上传
	// REF: https://cloud.tencent.com/document/api/1364/98588
	describeCloudNativeAPIGatewayCertificatesOffset := 0
	describeCloudNativeAPIGatewayCertificatesLimit := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeCloudNativeAPIGatewayCertificatesReq := tctse.NewDescribeCloudNativeAPIGatewayCertificatesRequest()
		describeCloudNativeAPIGatewayCertificatesReq.GatewayId = common.StringPtr(c.config.GatewayId)
		describeCloudNativeAPIGatewayCertificatesReq.Offset = common.Int64Ptr(int64(describeCloudNativeAPIGatewayCertificatesOffset))
		describeCloudNativeAPIGatewayCertificatesReq.Limit = common.Int64Ptr(int64(describeCloudNativeAPIGatewayCertificatesLimit))
		describeCloudNativeAPIGatewayCertificatesResp, err := c.sdkClient.TSE.DescribeCloudNativeAPIGatewayCertificatesWithContext(ctx, describeCloudNativeAPIGatewayCertificatesReq)
		c.logger.Debug("sdk request 'tse.DescribeCloudNativeAPIGatewayCertificates'", slog.Any("request", describeCloudNativeAPIGatewayCertificatesReq), slog.Any("response", describeCloudNativeAPIGatewayCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'tse.DescribeCloudNativeAPIGatewayCertificates': %w", err)
		}

		for _, certItem := range describeCloudNativeAPIGatewayCertificatesResp.Response.Result.CertificatesList {
			if lo.FromPtr(uploadCertificateResp.Response.CertificateId) == lo.FromPtr(certItem.CertId) ||
				xcert.EqualCertificatesFromPEM(certPEM, lo.FromPtr(certItem.Crt)) {
				// 如果已存在相同证书，直接返回
				c.logger.Info("ssl certificate already exists")
				return &UploadResult{
					CertId:   lo.FromPtr(certItem.CertId),
					CertName: lo.FromPtr(certItem.Name),
				}, nil
			}
		}

		if len(describeCloudNativeAPIGatewayCertificatesResp.Response.Result.CertificatesList) < describeCloudNativeAPIGatewayCertificatesLimit {
			break
		}

		describeCloudNativeAPIGatewayCertificatesOffset += describeCloudNativeAPIGatewayCertificatesLimit
	}

	// 生成新证书名（需符合腾讯云命名规则）
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 创建云原生网关证书
	// REF: https://cloud.tencent.com/document/api/1364/98591
	createCloudNativeAPIGatewayCertificateReq := tctse.NewCreateCloudNativeAPIGatewayCertificateRequest()
	createCloudNativeAPIGatewayCertificateReq.GatewayId = common.StringPtr(c.config.GatewayId)
	createCloudNativeAPIGatewayCertificateReq.Name = common.StringPtr(certName)
	createCloudNativeAPIGatewayCertificateReq.CertId = uploadCertificateResp.Response.CertificateId
	createCloudNativeAPIGatewayCertificateReq.BindDomains = common.StringPtrs(lo.Ternary(len(c.config.Domains) != 0, c.config.Domains, certX509.DNSNames))
	createCloudNativeAPIGatewayCertificateResp, err := c.sdkClient.TSE.CreateCloudNativeAPIGatewayCertificateWithContext(ctx, createCloudNativeAPIGatewayCertificateReq)
	c.logger.Debug("sdk request 'tse.CreateCloudNativeAPIGatewayCertificate'", slog.Any("request", createCloudNativeAPIGatewayCertificateReq), slog.Any("response", createCloudNativeAPIGatewayCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'tse.CreateCloudNativeAPIGatewayCertificate': %w", err)
	}

	return &UploadResult{
		CertId:   lo.FromPtr(createCloudNativeAPIGatewayCertificateResp.Response.Result.Id),
		CertName: certName,
	}, nil
}

func (c *Certmgr) replaceToCloudNative(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	// 更新云原生网关证书
	// REF: https://cloud.tencent.com/document/api/1364/100199
	modifyCloudNativeAPIGatewayCertificateReq := tctse.NewModifyCloudNativeAPIGatewayCertificateRequest()
	modifyCloudNativeAPIGatewayCertificateReq.GatewayId = common.StringPtr(c.config.GatewayId)
	modifyCloudNativeAPIGatewayCertificateReq.Id = common.StringPtr(certIdOrName)
	modifyCloudNativeAPIGatewayCertificateReq.Crt = common.StringPtr(certPEM)
	modifyCloudNativeAPIGatewayCertificateReq.Key = common.StringPtr(privkeyPEM)
	modifyCloudNativeAPIGatewayCertificateReq.CertSource = common.StringPtr("native")
	modifyCloudNativeAPIGatewayCertificateResp, err := c.sdkClient.TSE.ModifyCloudNativeAPIGatewayCertificateWithContext(ctx, modifyCloudNativeAPIGatewayCertificateReq)
	c.logger.Debug("sdk request 'tse.ModifyCloudNativeAPIGatewayCertificate'", slog.Any("request", modifyCloudNativeAPIGatewayCertificateReq), slog.Any("response", modifyCloudNativeAPIGatewayCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'tse.ModifyCloudNativeAPIGatewayCertificate': %w", err)
	}

	return &ReplaceResult{}, nil
}

func createSDKClients(secretId, secretKey, endpoint, region string) (*wSDKClients, error) {
	wsdk := &wSDKClients{}

	{
		credential := common.NewCredential(secretId, secretKey)

		cpf := profile.NewClientProfile()

		client, err := tcssl.NewClient(credential, "", cpf)
		if err != nil {
			return nil, err
		}

		wsdk.SSL = client
	}

	{
		credential := common.NewCredential(secretId, secretKey)

		cpf := profile.NewClientProfile()
		if endpoint != "" {
			cpf.HttpProfile.Endpoint = endpoint
		}

		client, err := tctse.NewClient(credential, region, cpf)
		if err != nil {
			return nil, err
		}

		wsdk.TSE = client
	}

	return wsdk, nil
}
