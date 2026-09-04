package ucloudussl

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/certimate-go/certimate/pkg/core"
	ucloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/ussl"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider    = core.Purger
	PurgeResult = core.PurgerPurgeResult
)

type PurgerConfig struct {
	// 优刻得 API 私钥。
	PrivateKey string `json:"privateKey"`
	// 优刻得 API 公钥。
	PublicKey string `json:"publicKey"`
	// 优刻得项目 ID。
	ProjectId string `json:"projectId,omitempty"`
	// 优刻得接口端点。
	Endpoint string `json:"endpoint,omitempty"`
}

type Purger struct {
	config    *PurgerConfig
	logger    *slog.Logger
	sdkClient *ucloudsdk.USSLClient
}

var _ Provider = (*Purger)(nil)

func NewPurger(config *PurgerConfig) (*Purger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the purger provider is nil")
	}

	client, err := createSDKClient(config.PrivateKey, config.PublicKey, config.ProjectId, config.Endpoint)
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
	purgingCertIds := make([]int, 0)
	purgingCertId2ModeMap := make(map[int]string)

	// 查询用户托管证书列表
	// REF: https://docs.ucloud.cn/api/usslcertificate-api/get_certificate_list
	getCertificateListPage := 1
	getCertificateListPageSize := 1000
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		getCertificateListReq := p.sdkClient.NewGetCertificateListRequest()
		getCertificateListReq.Mode = ucloud.String("trust")
		getCertificateListReq.Sort = ucloud.String("2")
		getCertificateListReq.Page = ucloud.Int(getCertificateListPage)
		getCertificateListReq.PageSize = ucloud.Int(getCertificateListPageSize)
		getCertificateListResp, err := p.sdkClient.GetCertificateList(getCertificateListReq)
		p.logger.Debug("sdk request 'ussl.GetCertificateList'", slog.Any("request", getCertificateListReq), slog.Any("response", getCertificateListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'ussl.GetCertificateList': %w", err)
		}

		for _, certItem := range getCertificateListResp.CertificateList {
			certNotAfter := time.UnixMilli(certItem.NotAfter)

			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, certItem.CertificateID)
				purgingCertId2ModeMap[certItem.CertificateID] = certItem.Mode
			}
		}

		if len(getCertificateListResp.CertificateList) < getCertificateListPageSize ||
			getCertificateListPage*getCertificateListPageSize >= getCertificateListResp.TotalCount {
			break
		}

		getCertificateListPage++
	}

	// 查询用户购买证书列表
	// REF: https://docs.ucloud.cn/api/usslcertificate-api/get_certificate_list
	getCertificateListPage = 1
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		getCertificateListReq := p.sdkClient.NewGetCertificateListRequest()
		getCertificateListReq.Mode = ucloud.String("purchase")
		getCertificateListReq.Sort = ucloud.String("2")
		getCertificateListReq.Page = ucloud.Int(getCertificateListPage)
		getCertificateListReq.PageSize = ucloud.Int(getCertificateListPageSize)
		getCertificateListResp, err := p.sdkClient.GetCertificateList(getCertificateListReq)
		p.logger.Debug("sdk request 'ussl.GetCertificateList'", slog.Any("request", getCertificateListReq), slog.Any("response", getCertificateListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'ussl.GetCertificateList': %w", err)
		}

		for _, certItem := range getCertificateListResp.CertificateList {
			certNotAfter := time.UnixMilli(certItem.NotAfter)

			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, certItem.CertificateID)
				purgingCertId2ModeMap[certItem.CertificateID] = certItem.Mode
			}
		}

		if len(getCertificateListResp.CertificateList) < getCertificateListPageSize ||
			getCertificateListPage*getCertificateListPageSize >= getCertificateListResp.TotalCount {
			break
		}

		getCertificateListPage++
	}

	// 删除证书
	// REF: https://docs.ucloud.cn/api/usslcertificate-api/delete_ssl_certificate
	purgedCount := 0
	if len(purgingCertIds) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certIds", purgingCertIds))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertIds, func(ctx context.Context, certId int, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			deleteSSLCertificateReq := p.sdkClient.NewDeleteSSLCertificateRequest()
			deleteSSLCertificateReq.CertificateMode = ucloud.String(purgingCertId2ModeMap[certId])
			deleteSSLCertificateReq.CertificateID = ucloud.Int(certId)
			deleteSSLCertificateResp, err := p.sdkClient.DeleteSSLCertificate(deleteSSLCertificateReq)
			p.logger.Debug("sdk request 'ussl.DeleteSSLCertificate'", slog.Any("request", deleteSSLCertificateReq), slog.Any("response", deleteSSLCertificateResp))
			if err != nil {
				if deleteSSLCertificateResp != nil && deleteSSLCertificateResp.GetRetCode() == 80011 {
					return nil
				}

				return fmt.Errorf("failed to execute sdk request 'ussl.DeleteSSLCertificate': %w", err)
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

func createSDKClient(privateKey, publicKey, projectId, endpoint string) (*ucloudsdk.USSLClient, error) {
	if privateKey == "" {
		return nil, fmt.Errorf("ucloud: invalid private key")
	}
	if publicKey == "" {
		return nil, fmt.Errorf("ucloud: invalid public key")
	}

	cfg := ucloud.NewConfig()
	if projectId != "" {
		cfg.ProjectId = projectId
	}
	if endpoint != "" {
		if strings.Contains(endpoint, "://") {
			cfg.BaseUrl = endpoint
		} else {
			cfg.BaseUrl = "https://" + endpoint
		}
	}

	credential := auth.NewCredential()
	credential.PrivateKey = privateKey
	credential.PublicKey = publicKey

	client := ucloudsdk.NewClient(&cfg, &credential)
	return client, nil
}
