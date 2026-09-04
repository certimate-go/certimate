package jdcloudssl

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	jdcore "github.com/jdcloud-api/jdcloud-sdk-go/core"
	jdsslapis "github.com/jdcloud-api/jdcloud-sdk-go/services/ssl/apis"

	jdssl "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/jdcloud-api/jdcloud-sdk-go/services/ssl/client"

	"github.com/certimate-go/certimate/pkg/core"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider    = core.Purger
	PurgeResult = core.PurgerPurgeResult
)

type PurgerConfig struct {
	// 京东云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 京东云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
}

type Purger struct {
	config    *PurgerConfig
	logger    *slog.Logger
	sdkClient *jdssl.SslClient
}

var _ Provider = (*Purger)(nil)

func NewPurger(config *PurgerConfig) (*Purger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the purger provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
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

	// 查看证书列表
	// REF: https://docs.jdcloud.com/cn/ssl-certificate/api/describecerts
	describeCertsPageNumber := 1
	describeCertsPageSize := 10
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeCertsReq := jdsslapis.NewDescribeCertsRequestWithoutParam()
		describeCertsReq.SetPageNumber(describeCertsPageNumber)
		describeCertsReq.SetPageSize(describeCertsPageSize)
		describeCertsResp, err := p.sdkClient.DescribeCerts(describeCertsReq)
		p.logger.Debug("sdk request 'ssl.DescribeCerts'", slog.Any("request", describeCertsReq), slog.Any("response", describeCertsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'ssl.DescribeCerts': %w", err)
		}

		for _, certItem := range describeCertsResp.Result.CertListDetails {
			certNotAfter, err := time.Parse(time.RFC3339, certItem.EndTime)
			if err != nil {
				continue
			}

			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, certItem.CertId)
			}
		}

		if len(describeCertsResp.Result.CertListDetails) < int(describeCertsPageSize) {
			break
		}

		describeCertsPageNumber++
	}

	// 删除证书
	// REF: https://docs.jdcloud.com/cn/ssl-certificate/api/deletecerts
	purgedCount := 0
	if len(purgingCertIds) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certIds", purgingCertIds))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertIds, func(ctx context.Context, certId string, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			deleteCertsReq := jdsslapis.NewDeleteCertsRequestWithoutParam()
			deleteCertsReq.SetCertId(certId)
			deleteCertResp, err := p.sdkClient.DeleteCerts(deleteCertsReq)
			p.logger.Debug("sdk request 'ssl.DeleteCerts'", slog.Any("request", deleteCertsReq), slog.Any("response", deleteCertResp))
			if err != nil {
				if deleteCertResp != nil && deleteCertResp.Error.Code == 404 {
					return nil
				}

				return fmt.Errorf("failed to execute sdk request 'ssl.DeleteCerts': %w", err)
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

func createSDKClient(accessKeyId, accessKeySecret string) (*jdssl.SslClient, error) {
	clientCredentials := jdcore.NewCredentials(accessKeyId, accessKeySecret)
	client := jdssl.NewSslClient(clientCredentials)
	client.DisableLogger()
	return client, nil
}
