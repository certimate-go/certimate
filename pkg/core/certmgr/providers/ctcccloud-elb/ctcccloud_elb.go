package ctcccloudelb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	ctyunelb "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/elb"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 天翼云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 天翼云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 天翼云资源池 ID。
	RegionId string `json:"regionId"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *ctyunelb.Client
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey)
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

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*certmgr.UploadResult, error) {
	// 查询证书列表，避免重复上传
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=24&api=5692&data=88&isNormal=1&vid=82
	listCertificatesReq := &ctyunelb.ListCertificatesRequest{
		RegionID: lo.ToPtr(c.config.RegionId),
	}
	listCertificatesResp, err := c.sdkClient.ListCertificatesWithContext(ctx, listCertificatesReq)
	c.logger.Debug("sdk request 'elb.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'elb.ListCertificates': %w", err)
	} else {
		for _, certItem := range listCertificatesResp.ReturnObj {
			// 如果已存在相同证书，直接返回
			if xcert.EqualCertificatesFromPEM(certPEM, certItem.Certificate) {
				c.logger.Info("ssl certificate already exists")
				return &certmgr.UploadResult{
					CertId:   certItem.ID,
					CertName: certItem.Name,
				}, nil
			}
		}
	}

	// 生成新证书名（需符合天翼云命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 创建证书
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=24&api=5685&data=88&isNormal=1&vid=82
	createCertificateReq := &ctyunelb.CreateCertificateRequest{
		ClientToken: lo.ToPtr(security.RandomString(32)),
		RegionID:    lo.ToPtr(c.config.RegionId),
		Name:        lo.ToPtr(certName),
		Description: lo.ToPtr("upload from certimate"),
		Type:        lo.ToPtr("Server"),
		Certificate: lo.ToPtr(certPEM),
		PrivateKey:  lo.ToPtr(privkeyPEM),
	}
	createCertificateResp, err := c.sdkClient.CreateCertificateWithContext(ctx, createCertificateReq)
	c.logger.Debug("sdk request 'elb.CreateCertificate'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'elb.CreateCertificate': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   createCertificateResp.ReturnObj.ID,
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
}

func createSDKClient(accessKeyId, secretAccessKey string) (*ctyunelb.Client, error) {
	return ctyunelb.NewClient(accessKeyId, secretAccessKey)
}
