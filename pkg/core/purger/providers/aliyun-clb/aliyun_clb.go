package aliyunclb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	alislb "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/alibabacloud-go/slb-20140515/v4/client"

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
	sdkClient *alislb.Client
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
	// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-describeservercertificates
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		describeServerCertificatesReq := &alislb.DescribeServerCertificatesRequest{
			ResourceGroupId: lo.EmptyableToPtr(p.config.ResourceGroupId),
			RegionId:        tea.String(p.config.Region),
		}
		describeServerCertificatesResp, err := p.sdkClient.DescribeServerCertificatesWithContext(ctx, describeServerCertificatesReq, &dara.RuntimeOptions{})
		p.logger.Debug("sdk request 'slb.DescribeServerCertificates'", slog.Any("request", describeServerCertificatesReq), slog.Any("response", describeServerCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'slb.DescribeServerCertificates': %w", err)
		}

		for _, certItem := range describeServerCertificatesResp.Body.ServerCertificates.ServerCertificate {
			certNotAfter := time.Unix(tea.Int64Value(certItem.ExpireTimeStamp), 0)
			if certNotAfter.IsZero() {
				continue
			}

			if time.Since(certNotAfter) >= expiry {
				purgingCertIds = append(purgingCertIds, tea.StringValue(certItem.ServerCertificateId))
			}
		}

		break
	}

	// 删除服务器证书
	// REF: https://help.aliyun.com/zh/slb/classic-load-balancer/developer-reference/api-slb-2014-05-15-deleteservercertificate
	purgedCount := 0
	if len(purgingCertIds) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certIds", purgingCertIds))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertIds, func(ctx context.Context, certId string, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			deleteServerCertificateReq := &alislb.DeleteServerCertificateRequest{
				RegionId:            tea.String(p.config.Region),
				ServerCertificateId: tea.String(certId),
			}
			deleteServerCertificateResp, err := p.sdkClient.DeleteServerCertificateWithContext(ctx, deleteServerCertificateReq, &dara.RuntimeOptions{})
			p.logger.Debug("sdk request 'slb.DeleteServerCertificate'", slog.Any("request", deleteServerCertificateReq), slog.Any("response", deleteServerCertificateResp))
			if err != nil {
				if sdkErr, ok := err.(*tea.SDKError); ok {
					if sdkErrCode := tea.StringValue(sdkErr.Code); sdkErrCode == "ServerCertificateId.NotFound" {
						return nil
					}
				}

				return fmt.Errorf("failed to execute sdk request 'slb.DeleteServerCertificate': %w", err)
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

func createSDKClient(accessKeyId, accessKeySecret, region string) (*alislb.Client, error) {
	// 接入点一览 https://api.aliyun.com/product/Slb
	var endpoint string
	switch region {
	case "",
		"cn-hangzhou",
		"cn-hangzhou-finance",
		"cn-shanghai-finance-1",
		"cn-shenzhen-finance-1":
		endpoint = "slb.aliyuncs.com"
	default:
		endpoint = fmt.Sprintf("slb.%s.aliyuncs.com", region)
	}

	config := &aliopen.Config{
		Endpoint:        tea.String(endpoint),
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
	}

	client, err := alislb.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
