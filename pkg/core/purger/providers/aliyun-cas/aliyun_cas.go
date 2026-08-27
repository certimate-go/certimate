package aliyuncas

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	alicas "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/alibabacloud-go/cas-20200407/v4/client"

	"github.com/certimate-go/certimate/pkg/core"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider    = core.Purger
	PurgeResult = core.PurgerPurgeResult
)

type PurgerConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 阿里云地域。
	Region string `json:"region"`
}

type Purger struct {
	config    *PurgerConfig
	logger    *slog.Logger
	sdkClient *alicas.Client
}

var _ Provider = (*Purger)(nil)

func NewPurger(config *PurgerConfig) (*Purger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the purger provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
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
	// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-listcertificates
	listCertificatesPage := 1
	listCertificatesLimit := 50
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &alicas.ListCertificatesRequest{
			ResourceGroupId:   lo.EmptyableToPtr(p.config.ResourceGroupId),
			CertificateStatus: tea.String("expired"),
			CurrentPage:       tea.Int32(int32(listCertificatesPage)),
			ShowSize:          tea.Int32(int32(listCertificatesLimit)),
		}
		listCertificatesResp, err := p.sdkClient.ListCertificatesWithContext(ctx, listCertificatesReq, &dara.RuntimeOptions{})
		p.logger.Debug("sdk request 'cas.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cas.ListCertificates': %w", err)
		}

		for _, certItem := range listCertificatesResp.Body.CertificateList {
			certNotAfter := time.UnixMilli(tea.Int64Value(certItem.NotAfter))
			if certNotAfter.IsZero() {
				continue
			}

			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, tea.StringValue(certItem.CertificateId))
			}
		}

		if len(listCertificatesResp.Body.CertificateList) < listCertificatesLimit {
			break
		}

		listCertificatesPage++
	}

	// 删除证书
	// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-deleteusercertificate
	purgedCount := 0
	if len(purgingCertIds) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certIds", purgingCertIds))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertIds, func(ctx context.Context, certId string, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			certIdAsInt, err := strconv.ParseInt(certId, 10, 64)
			if err != nil {
				return err
			}

			deleteUserCertificateReq := &alicas.DeleteUserCertificateRequest{
				CertId: tea.Int64(certIdAsInt),
			}
			deleteUserCertificateResp, err := p.sdkClient.DeleteUserCertificateWithContext(ctx, deleteUserCertificateReq, &dara.RuntimeOptions{})
			p.logger.Debug("sdk request 'cas.DeleteUserCertificate'", slog.Any("request", deleteUserCertificateReq), slog.Any("response", deleteUserCertificateResp))
			if err != nil {
				if sdkErr, ok := err.(*tea.SDKError); ok {
					if sdkErrCode := tea.StringValue(sdkErr.Code); strings.HasPrefix(sdkErrCode, "NotFound") {
						return nil
					}
				}

				return fmt.Errorf("failed to execute sdk request 'cas.DeleteUserCertificate': %w", err)
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

func createSDKClient(accessKeyId, accessKeySecret, region string) (*alicas.Client, error) {
	// 接入点一览 https://api.aliyun.com/product/cas
	var endpoint string
	switch region {
	case "", "cn-hangzhou":
		endpoint = "cas.aliyuncs.com"
	default:
		endpoint = fmt.Sprintf("cas.%s.aliyuncs.com", region)
	}

	config := &aliopen.Config{
		Endpoint:        tea.String(endpoint),
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
	}

	client, err := alicas.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
