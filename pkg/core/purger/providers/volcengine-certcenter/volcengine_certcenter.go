package volcenginecertcenter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"
	ve "github.com/volcengine/volcengine-go-sdk/volcengine"
	vesession "github.com/volcengine/volcengine-go-sdk/volcengine/session"
	veerr "github.com/volcengine/volcengine-go-sdk/volcengine/volcengineerr"

	vecertificateservice "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/volcengine/volcengine-go-sdk/service/certificateservice"

	"github.com/certimate-go/certimate/pkg/core"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider    = core.Purger
	PurgeResult = core.PurgerPurgeResult
)

type PurgerConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 火山引擎项目名称。
	ProjectName string `json:"projectName,omitempty"`
	// 火山引擎地域。
	Region string `json:"region"`
}

type Purger struct {
	config    *PurgerConfig
	logger    *slog.Logger
	sdkClient *vecertificateservice.CERTIFICATESERVICE
}

var _ Provider = (*Purger)(nil)

func NewPurger(config *PurgerConfig) (*Purger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the purger provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
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

	// 获取证书实例列表
	// REF: https://docs.volcengine.com/docs/6638/1413343
	certificateGetInstanceListPageNumber := 1
	certificateGetInstanceListPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		certificateGetInstanceListReq := &vecertificateservice.CertificateGetInstanceListInput{
			ProjectName: lo.EmptyableToPtr(p.config.ProjectName),
			PageNumber:  lo.ToPtr(int32(certificateGetInstanceListPageNumber)),
			PageSize:    lo.ToPtr(int32(certificateGetInstanceListPageSize)),
		}
		certificateGetInstanceListResp, err := p.sdkClient.CertificateGetInstanceListWithContext(ctx, certificateGetInstanceListReq)
		p.logger.Debug("sdk request 'certificateservice.CertificateGetInstanceList'", slog.Any("request", certificateGetInstanceListReq), slog.Any("response", certificateGetInstanceListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'certificateservice.CertificateGetInstanceList': %w", err)
		}

		for _, certItem := range certificateGetInstanceListResp.Instances {
			certNotAfter, err := time.ParseInLocation(time.DateTime, lo.FromPtr(certItem.NotAfter), time.FixedZone("CST", 8*60*60))
			if err != nil {
				continue
			}

			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, lo.FromPtr(certItem.InstanceId))
			}
		}

		if len(certificateGetInstanceListResp.Instances) < certificateGetInstanceListPageSize ||
			certificateGetInstanceListPageNumber*certificateGetInstanceListPageSize >= int(lo.FromPtr(certificateGetInstanceListResp.TotalCount)) {
			break
		}

		certificateGetInstanceListPageNumber++
	}

	// 删除证书实例
	// REF: https://docs.volcengine.com/docs/6638/1412931
	purgedCount := 0
	if len(purgingCertIds) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certIds", purgingCertIds))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertIds, func(ctx context.Context, certId string, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			certificateDeleteInstanceReq := &vecertificateservice.CertificateDeleteInstanceInput{
				InstanceId: lo.ToPtr(certId),
			}
			certificateDeleteInstanceResp, err := p.sdkClient.CertificateDeleteInstanceWithContext(ctx, certificateDeleteInstanceReq)
			p.logger.Debug("sdk request 'certificateservice.CertificateDeleteInstance'", slog.Any("request", certificateDeleteInstanceReq), slog.Any("response", certificateDeleteInstanceResp))
			if err != nil {
				var sdkErr veerr.RequestFailure
				if errors.As(err, &sdkErr) {
					if sdkErrCode := sdkErr.Code(); sdkErrCode == "NotFound.CertificateInstance" || sdkErrCode == "Delete.Certificate" {
						return nil
					}
				}

				return fmt.Errorf("failed to execute sdk request 'certificateservice.CertificateDeleteInstance': %w", err)
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

func createSDKClient(accessKeyId, secretAccessKey, region string) (*vecertificateservice.CERTIFICATESERVICE, error) {
	if region == "" {
		region = "cn-beijing" // 证书中心默认区域：北京
	}

	config := ve.NewConfig().
		WithAkSk(accessKeyId, secretAccessKey).
		WithRegion(region)

	session, err := vesession.NewSession(config)
	if err != nil {
		return nil, err
	}

	client := vecertificateservice.New(session, config)
	return client, nil
}
