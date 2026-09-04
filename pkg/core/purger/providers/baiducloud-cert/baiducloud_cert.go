package baiducloudcert

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/baidubce/bce-sdk-go/bce"

	"github.com/certimate-go/certimate/pkg/core"
	baiducert "github.com/certimate-go/certimate/pkg/sdk3rd/baiducloud/cert"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider    = core.Purger
	PurgeResult = core.PurgerPurgeResult
)

type PurgerConfig struct {
	// 百度智能云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 百度智能云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
}

type Purger struct {
	config    *PurgerConfig
	logger    *slog.Logger
	sdkClient *baiducert.Client
}

var _ Provider = (*Purger)(nil)

func NewPurger(config *PurgerConfig) (*Purger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the purger provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey)
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
	// REF: https://cloud.baidu.com/doc/Reference/s/Gjwvz27xu#34-%E6%9F%A5%E7%9C%8B%E8%AF%81%E4%B9%A6%E5%88%97%E8%A1%A8
	listCertDetail, err := p.sdkClient.ListCertDetail()
	p.logger.Debug("sdk request 'cert.ListCertDetail'", slog.Any("response", listCertDetail))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cert.ListCertDetail': %w", err)
	} else {
		for _, certItem := range listCertDetail.Certs {
			certNotAfter, err := time.Parse(time.RFC3339, certItem.CertStopTime)
			if err != nil {
				continue
			}

			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, certItem.CertId)
			}
		}
	}

	// 删除证书
	// REF: https://cloud.baidu.com/doc/Reference/s/Gjwvz27xu#37-%E5%88%A0%E9%99%A4%E8%AF%81%E4%B9%A6
	purgedCount := 0
	if len(purgingCertIds) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certIds", purgingCertIds))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertIds, func(ctx context.Context, certId string, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			err := p.sdkClient.DeleteCert(certId)
			p.logger.Debug("sdk request 'cert.DeleteCert'", slog.String("params.certId", certId))
			if err != nil {
				if sdkErr, ok := err.(*bce.BceServiceError); ok {
					if sdkErrCode := sdkErr.Code; sdkErrCode == "ResourceNotFoundException" || sdkErrCode == "OperationNotAllowedException" {
						return nil
					}
				}

				return fmt.Errorf("failed to execute sdk request 'cert.DeleteCert': %w", err)
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

func createSDKClient(accessKeyId, secretAccessKey string) (*baiducert.Client, error) {
	client, err := baiducert.NewClient(accessKeyId, secretAccessKey, "")
	if err != nil {
		return nil, err
	}

	return client, nil
}
