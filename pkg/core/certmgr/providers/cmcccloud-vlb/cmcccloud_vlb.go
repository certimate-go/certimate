package cmcccloudvlb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/config"
	"gitlab.ecloud.com/ecloud/ecloudsdkvlb/model"

	"github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/gitlab.ecloud.com/ecloud/ecloudsdkvlb"

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// 移动云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 移动云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 移动云资源池 ID。
	PoolId string `json:"poolId"`
	// 是否是 SNI 证书。
	IsSNI bool `json:"isSni,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *ecloudsdkvlb.Client
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.PoolId)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Certmgr{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (c *Certmgr) SetLogger(logger *slog.Logger) {
	if logger == nil {
		c.logger = slog.New(slog.DiscardHandler)
	} else {
		c.logger = logger
	}
}

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*UploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 查询弹性负载均衡证书列表，避免重复上传
	// REF: https://ecloud.10086.cn/op-help-center/doc/article/97613
	// REF: https://ecloud.10086.cn/op-help-center/doc/article/97399
	listLoadbalanceCertificationResp, err := c.sdkClient.ListLoadbalanceCertificationResp()
	c.logger.Debug("sdk request 'vlb.ListLoadbalanceCertification'", slog.Any("response", listLoadbalanceCertificationResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'vlb.ListLoadbalanceCertification': %w", err)
	} else {
		if listLoadbalanceCertificationResp.Body != nil && listLoadbalanceCertificationResp.Body.Content != nil {
			for _, certItem := range *listLoadbalanceCertificationResp.Body.Content {
				// 对比证书有效期
				oldCertNotAfter, _ := time.Parse(time.DateTime, lo.FromPtr(certItem.ExpirationTime))
				if !certX509.NotAfter.Equal(oldCertNotAfter) {
					continue
				}

				// 对比证书内容
				getLoadbalanceCertificationDetailReq := &model.GetLoadbalanceCertificationDetailRespRequest{
					&model.GetLoadbalanceCertificationDetailRespPath{
						ContainerUuid: certItem.Id,
					},
				}
				getLoadbalanceCertificationDetailResp, err := c.sdkClient.GetLoadbalanceCertificationDetailResp(getLoadbalanceCertificationDetailReq)
				c.logger.Debug("sdk request 'vlb.GetLoadbalanceCertificationDetail'", slog.Any("request", getLoadbalanceCertificationDetailReq), slog.Any("response", getLoadbalanceCertificationDetailResp))
				if err != nil {
					return nil, fmt.Errorf("failed to execute sdk request 'vlb.GetLoadbalanceCertificationDetail': %w", err)
				} else {
					if !xcert.EqualCertificatesFromPEM(certPEM, lo.FromPtr(getLoadbalanceCertificationDetailResp.Body.PublicKey)) {
						continue
					}
				}

				// 如果以上信息都一致，则视为已存在相同证书，直接返回
				c.logger.Info("ssl certificate already exists")
				return &UploadResult{
					CertId:   lo.FromPtr(certItem.Id),
					CertName: lo.FromPtr(certItem.Name),
				}, nil
			}
		}
	}

	// 生成新证书名（需符合移动云命名规则）
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 创建弹性负载均衡证书
	// REF: https://ecloud.10086.cn/op-help-center/doc/article/97397
	createLoadbalanceCertificationReq := &model.CreateLoadbalanceCertificationRequest{
		&model.CreateLoadbalanceCertificationBody{
			Name:        lo.ToPtr(certName),
			Type:        lo.ToPtr(lo.Ternary(c.config.IsSNI, model.CreateLoadbalanceCertificationBodyTypeEnumSni, model.CreateLoadbalanceCertificationBodyTypeEnumServer)),
			Description: lo.ToPtr("upload from Certimate"),
			PublicKey:   lo.ToPtr(certPEM),
			PrivateKey:  lo.ToPtr(privkeyPEM),
		},
	}
	createLoadbalanceCertificationResp, err := c.sdkClient.CreateLoadbalanceCertification(createLoadbalanceCertificationReq)
	c.logger.Debug("sdk request 'vlb.CreateLoadbalanceCertification'", slog.Any("request", createLoadbalanceCertificationReq), slog.Any("response", createLoadbalanceCertificationResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'vlb.CreateLoadbalanceCertification': %w", err)
	}

	return &UploadResult{
		CertId:   lo.FromPtr(createLoadbalanceCertificationResp.Body),
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}

func createSDKClient(accessKeyId, accessKeySecret, poolId string) (*ecloudsdkvlb.Client, error) {
	client := ecloudsdkvlb.NewClient(&config.Config{
		AccessKey: &accessKeyId,
		SecretKey: &accessKeySecret,
		RegionId:  &poolId,
	})

	return client, nil
}
