package tencentcloudssl

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	tcssl "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"

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
	sdkClient *tcssl.Client
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

	// 获取证书列表
	// REF: https://cloud.tencent.com/document/api/400/41671
	listCertificatesOffset := 0
	listCertificatesLimit := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeCertificatesReq := tcssl.NewDescribeCertificatesRequest()
		describeCertificatesReq.ProjectId = lo.EmptyableToPtr(uint64(p.config.ProjectId))
		describeCertificatesReq.CertificateType = common.StringPtr("SVR")
		describeCertificatesReq.CertificateStatus = common.Uint64Ptrs([]uint64{3})
		describeCertificatesReq.IsSM = common.Int64Ptr(0)
		describeCertificatesReq.Offset = common.Uint64Ptr(uint64(listCertificatesOffset))
		describeCertificatesReq.Limit = common.Uint64Ptr(uint64(listCertificatesLimit))
		describeCertificatesResp, err := p.sdkClient.DescribeCertificatesWithContext(ctx, describeCertificatesReq)
		p.logger.Debug("sdk request 'ssl.DescribeCertificates'", slog.Any("request", describeCertificatesReq), slog.Any("response", describeCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'ssl.DescribeCertificates': %w", err)
		}

		for _, certItem := range describeCertificatesResp.Response.Certificates {
			certNotAfter, err := time.ParseInLocation(time.DateTime, lo.FromPtr(certItem.CertEndTime), time.FixedZone("CST", 8*60*60))
			if err != nil {
				continue
			}

			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, lo.FromPtr(certItem.CertificateId))
			}
		}

		if len(describeCertificatesResp.Response.Certificates) < listCertificatesLimit {
			break
		}

		listCertificatesOffset += listCertificatesLimit
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

			deleteCertificateReq := tcssl.NewDeleteCertificateRequest()
			deleteCertificateReq.CertificateId = common.StringPtr(certId)
			deleteCertificateResp, err := p.sdkClient.DeleteCertificateWithContext(ctx, deleteCertificateReq)
			p.logger.Debug("sdk request 'ssl.DeleteCertificate'", slog.Any("request", deleteCertificateReq), slog.Any("response", deleteCertificateResp))
			if err != nil {
				if sdkErr, ok := err.(*tcerrors.TencentCloudSDKError); ok {
					if sdkErrCode := sdkErr.Code; sdkErrCode == "FailedOperation.CertificateNotFound" || sdkErrCode == "FailedOperation.BoundResources" || sdkErrCode == "FailedOperation.DeleteResourceFailed" {
						return nil
					}
				}

				return fmt.Errorf("failed to execute sdk request 'ssl.DeleteCertificate': %w", err)
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

func createSDKClient(secretId, secretKey, endpoint string) (*tcssl.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcssl.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
