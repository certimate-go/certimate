package huaweicloudelb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	hwelbmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/model"
	hwelbregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/region"
	"github.com/samber/lo"

	hwelb "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3"

	"github.com/certimate-go/certimate/pkg/core"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider    = core.Purger
	PurgeResult = core.PurgerPurgeResult
)

type PurgerConfig struct {
	// 华为云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 华为云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 华为云企业项目 ID。
	EnterpriseProjectId string `json:"enterpriseProjectId,omitempty"`
	// 华为云区域。
	Region string `json:"region"`
}

type Purger struct {
	config    *PurgerConfig
	logger    *slog.Logger
	sdkClient *hwelb.ElbClient
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

	// 查询证书列表
	// REF: https://support.huaweicloud.com/api-elb/ListCertificates.html
	listCertificatesMarker := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &hwelbmodel.ListCertificatesRequest{
			Marker: listCertificatesMarker,
			Limit:  lo.ToPtr(int32(2000)),
			Type:   lo.ToPtr([]string{"server", "server_sm"}),
		}
		listCertificatesResp, err := p.sdkClient.ListCertificates(listCertificatesReq)
		p.logger.Debug("sdk request 'elb.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'elb.ListCertificates': %w", err)
		}

		for _, certItem := range *listCertificatesResp.Certificates {
			certNotAfter, err := time.Parse(time.RFC3339, certItem.ExpireTime)
			if err != nil {
				continue
			}

			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, certItem.Id)
			}
		}

		if len(*listCertificatesResp.Certificates) == 0 || listCertificatesResp.PageInfo.NextMarker == nil {
			break
		}

		listCertificatesMarker = listCertificatesResp.PageInfo.NextMarker
	}

	// 删除证书
	// REF: https://support.huaweicloud.com/api-elb/DeleteCertificate.html
	purgedCount := 0
	if len(purgingCertIds) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certIds", purgingCertIds))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertIds, func(ctx context.Context, certId string, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			deleteCertificateReq := &hwelbmodel.DeleteCertificateRequest{
				CertificateId: certId,
			}
			deleteCertificateResp, err := p.sdkClient.DeleteCertificate(deleteCertificateReq)
			p.logger.Debug("sdk request 'elb.DeleteCertificate'", slog.Any("request", deleteCertificateReq), slog.Any("response", deleteCertificateResp))
			if err != nil {
				if deleteCertificateResp != nil && (deleteCertificateResp.HttpStatusCode == 404 || deleteCertificateResp.HttpStatusCode == 409) {
					return nil
				}

				return fmt.Errorf("failed to execute sdk request 'elb.DeleteCertificate': %w", err)
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

func createSDKClient(accessKeyId, secretAccessKey, region string) (*hwelb.ElbClient, error) {
	if region == "" {
		region = "cn-north-4" // ELB 服务默认区域：华北北京四
	}

	auth, err := basic.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	hcRegion, err := hwelbregion.SafeValueOf(region)
	if err != nil {
		return nil, err
	}

	hcClient, err := hwelb.ElbClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	client := hwelb.NewElbClient(hcClient)
	return client, nil
}
