package huaweicloudwaf

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	hwwafmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1/model"
	hwwafregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1/region"
	"github.com/samber/lo"

	hwwaf "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1"

	"github.com/certimate-go/certimate/pkg/core"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xhuaweicloud "github.com/certimate-go/certimate/pkg/utils/third-party/huaweicloud"
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
	sdkClient *hwwaf.WafClient
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
	// REF: https://support.huaweicloud.com/api-waf/ListCertificates.html
	listCertificatesPage := 1
	listCertificatesPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &hwwafmodel.ListCertificatesRequest{
			EnterpriseProjectId: lo.EmptyableToPtr(p.config.EnterpriseProjectId),
			ExpStatus:           lo.ToPtr(hwwafmodel.GetListCertificatesRequestExpStatusEnum().E_1),
			Page:                lo.ToPtr(int32(listCertificatesPage)),
			Pagesize:            lo.ToPtr(int32(listCertificatesPageSize)),
		}
		listCertificatesResp, err := p.sdkClient.ListCertificates(listCertificatesReq)
		p.logger.Debug("sdk request 'waf.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'waf.ListCertificates': %w", err)
		}

		for _, certItem := range *listCertificatesResp.Items {
			certNotAfter := time.UnixMilli(lo.FromPtr(certItem.ExpireTime))
			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, certItem.Id)
			}
		}

		if len(*listCertificatesResp.Items) < listCertificatesPageSize {
			break
		}

		listCertificatesPage++
	}

	// 删除证书
	// REF: https://support.huaweicloud.com/api-waf/DeleteCertificate.html
	purgedCount := 0
	if len(purgingCertIds) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certIds", purgingCertIds))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertIds, func(ctx context.Context, certId string, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			deleteCertificateReq := &hwwafmodel.DeleteCertificateRequest{
				EnterpriseProjectId: lo.EmptyableToPtr(p.config.EnterpriseProjectId),
				CertificateId:       certId,
			}
			deleteCertificateResp, err := p.sdkClient.DeleteCertificate(deleteCertificateReq)
			p.logger.Debug("sdk request 'waf.DeleteCertificate'", slog.Any("request", deleteCertificateReq), slog.Any("response", deleteCertificateResp))
			if err != nil {
				if deleteCertificateResp != nil && deleteCertificateResp.HttpStatusCode == 404 {
					return nil
				}

				return fmt.Errorf("failed to execute sdk request 'waf.DeleteCertificate': %w", err)
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

func createSDKClient(accessKeyId, secretAccessKey, region string) (*hwwaf.WafClient, error) {
	projectId, err := getSDKProjectId(accessKeyId, secretAccessKey, region)
	if err != nil {
		return nil, err
	}

	auth, err := basic.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		WithProjectId(projectId).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	hcRegion, err := hwwafregion.SafeValueOf(region)
	if err != nil {
		return nil, err
	}

	hcClient, err := hwwaf.WafClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	client := hwwaf.NewWafClient(hcClient)
	return client, nil
}

func getSDKProjectId(accessKeyId, secretAccessKey, region string) (string, error) {
	auth, err := global.NewCredentialsBuilder().
		WithAk(accessKeyId).
		WithSk(secretAccessKey).
		SafeBuild()
	if err != nil {
		return "", err
	}

	return xhuaweicloud.GetKeystoneProjectIDWithRegion(auth, region)
}
