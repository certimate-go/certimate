package tencentcloudtse

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	tctse "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tse/v20201207"

	"github.com/certimate-go/certimate/pkg/core"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider    = core.Purger
	PurgeResult = core.PurgerPurgeResult
)

type PurgerConfig struct {
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
}

type Purger struct {
	config    *PurgerConfig
	logger    *slog.Logger
	sdkClient *tctse.Client
}

var _ Provider = (*Purger)(nil)

func NewPurger(config *PurgerConfig) (*Purger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the purger provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint, config.Region)
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
	switch p.config.ServiceType {
	case SERVICE_TYPE_CLOUDNATIVE:
		return p.uploadToCloudNative(ctx, expiry)

	default:
		return nil, fmt.Errorf("unsupported service type '%s'", p.config.ServiceType)
	}
}

func (p *Purger) uploadToCloudNative(ctx context.Context, expiry time.Duration) (*PurgeResult, error) {
	purgingCertIds := make([]string, 0)

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
		describeCloudNativeAPIGatewayCertificatesReq.GatewayId = common.StringPtr(p.config.GatewayId)
		describeCloudNativeAPIGatewayCertificatesReq.Offset = common.Int64Ptr(int64(describeCloudNativeAPIGatewayCertificatesOffset))
		describeCloudNativeAPIGatewayCertificatesReq.Limit = common.Int64Ptr(int64(describeCloudNativeAPIGatewayCertificatesLimit))
		describeCloudNativeAPIGatewayCertificatesResp, err := p.sdkClient.DescribeCloudNativeAPIGatewayCertificatesWithContext(ctx, describeCloudNativeAPIGatewayCertificatesReq)
		p.logger.Debug("sdk request 'tse.DescribeCloudNativeAPIGatewayCertificates'", slog.Any("request", describeCloudNativeAPIGatewayCertificatesReq), slog.Any("response", describeCloudNativeAPIGatewayCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'tse.DescribeCloudNativeAPIGatewayCertificates': %w", err)
		}

		for _, certItem := range describeCloudNativeAPIGatewayCertificatesResp.Response.Result.CertificatesList {
			certNotAfter, err := time.ParseInLocation(time.DateTime, lo.FromPtr(certItem.ExpireTime), time.FixedZone("CST", 8*60*60))
			if err != nil {
				continue
			}

			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, lo.FromPtr(certItem.CertId))
			}
		}

		if len(describeCloudNativeAPIGatewayCertificatesResp.Response.Result.CertificatesList) < describeCloudNativeAPIGatewayCertificatesLimit {
			break
		}

		describeCloudNativeAPIGatewayCertificatesOffset += describeCloudNativeAPIGatewayCertificatesLimit
	}

	// 删除证书
	// REF: https://cloud.tencent.com/document/api/400/41675
	purgedCount := 0
	if len(purgingCertIds) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certIds", purgingCertIds))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertIds, func(ctx context.Context, certId string, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			deleteCloudNativeAPIGatewayCertificateReq := tctse.NewDeleteCloudNativeAPIGatewayCertificateRequest()
			deleteCloudNativeAPIGatewayCertificateReq.GatewayId = common.StringPtr(p.config.GatewayId)
			deleteCloudNativeAPIGatewayCertificateReq.Id = common.StringPtr(certId)
			deleteCloudNativeAPIGatewayCertificateResp, err := p.sdkClient.DeleteCloudNativeAPIGatewayCertificateWithContext(ctx, deleteCloudNativeAPIGatewayCertificateReq)
			p.logger.Debug("sdk request 'tse.DeleteCloudNativeAPIGatewayCertificate'", slog.Any("request", deleteCloudNativeAPIGatewayCertificateReq), slog.Any("response", deleteCloudNativeAPIGatewayCertificateResp))
			if err != nil {
				if sdkErr, ok := err.(*tcerrors.TencentCloudSDKError); ok {
					if sdkErrCode := sdkErr.Code; sdkErrCode == "ResourceNotFound.ResourceNotFound" {
						return nil
					}
				}

				return fmt.Errorf("failed to execute sdk request 'tse.DeleteCloudNativeAPIGatewayCertificate': %w", err)
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

func createSDKClient(secretId, secretKey, endpoint, region string) (*tctse.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tctse.NewClient(credential, region, cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
