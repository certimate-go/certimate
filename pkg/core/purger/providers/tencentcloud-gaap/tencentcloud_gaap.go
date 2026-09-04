package tencentcloudgaap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	tcgaap "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/gaap/v20180529"

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
}

type Purger struct {
	config    *PurgerConfig
	logger    *slog.Logger
	sdkClient *tcgaap.Client
}

var _ Provider = (*Purger)(nil)

func NewPurger(config *PurgerConfig) (*Purger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the purger provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
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
	purgingCertIds := make([]string, 0)

	// 查询服务器证书列表
	// REF: https://cloud.tencent.com/document/api/608/36977
	describeCertificatesOffset := 0
	describeCertificatesLimit := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeCertificatesReq := tcgaap.NewDescribeCertificatesRequest()
		describeCertificatesReq.CertificateType = common.Int64Ptr(2)
		describeCertificatesReq.Offset = common.Uint64Ptr(uint64(describeCertificatesOffset))
		describeCertificatesReq.Limit = common.Uint64Ptr(uint64(describeCertificatesLimit))
		describeCertificatesResp, err := p.sdkClient.DescribeCertificatesWithContext(ctx, describeCertificatesReq)
		p.logger.Debug("sdk request 'gaap.DescribeCertificates'", slog.Any("request", describeCertificatesReq), slog.Any("response", describeCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'gaap.DescribeCertificates': %w", err)
		}

		for _, certItem := range describeCertificatesResp.Response.CertificateSet {
			certNotAfter := time.Unix(int64(lo.FromPtr(certItem.EndTime)), 0)
			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, lo.FromPtr(certItem.CertificateId))
			}
		}

		if len(describeCertificatesResp.Response.CertificateSet) < describeCertificatesLimit {
			break
		}

		describeCertificatesOffset += describeCertificatesLimit
	}

	// 删除证书
	// REF: https://cloud.tencent.com/document/api/608/36979
	purgedCount := 0
	if len(purgingCertIds) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certIds", purgingCertIds))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertIds, func(ctx context.Context, certId string, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			deleteCertificateReq := tcgaap.NewDeleteCertificateRequest()
			deleteCertificateReq.CertificateId = common.StringPtr(certId)
			deleteCertificateResp, err := p.sdkClient.DeleteCertificateWithContext(ctx, deleteCertificateReq)
			p.logger.Debug("sdk request 'gaap.DeleteCertificate'", slog.Any("request", deleteCertificateReq), slog.Any("response", deleteCertificateResp))
			if err != nil {
				if sdkErr, ok := err.(*tcerrors.TencentCloudSDKError); ok {
					if sdkErrCode := sdkErr.Code; sdkErrCode == "ResourceNotFound" || sdkErrCode == "FailedOperation.CertificateIsUsing" {
						return nil
					}
				}

				return fmt.Errorf("failed to execute sdk request 'gaap.DeleteCertificate': %w", err)
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

func createSDKClient(secretId, secretKey, endpoint string) (*tcgaap.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcgaap.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
