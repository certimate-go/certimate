package volcenginelive

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	velive "github.com/volcengine/volc-sdk-golang/service/live/v20230101"
	ve "github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *velive.Live
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client := velive.NewInstance()
	client.SetAccessKey(config.AccessKeyId)
	client.SetSecretKey(config.AccessKeySecret)

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
	// REF: https://www.volcengine.com/docs/6469/1186278#%E6%9F%A5%E8%AF%A2%E8%AF%81%E4%B9%A6%E5%88%97%E8%A1%A8
	listCertReq := &velive.ListCertV2Body{}
	listCertResp, err := c.sdkClient.ListCertV2(ctx, listCertReq)
	c.logger.Debug("sdk request 'live.ListCertV2'", slog.Any("request", listCertReq), slog.Any("response", listCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'live.ListCertV2': %w", err)
	}
	if listCertResp.Result.CertList != nil {
		for _, certItem := range listCertResp.Result.CertList {
			// 查询证书详细信息
			// REF: https://www.volcengine.com/docs/6469/1186278#%E6%9F%A5%E7%9C%8B%E8%AF%81%E4%B9%A6%E8%AF%A6%E6%83%85
			describeCertDetailSecretReq := &velive.DescribeCertDetailSecretV2Body{
				ChainID: ve.String(certItem.ChainID),
			}
			describeCertDetailSecretResp, err := c.sdkClient.DescribeCertDetailSecretV2(ctx, describeCertDetailSecretReq)
			c.logger.Debug("sdk request 'live.DescribeCertDetailSecretV2'", slog.Any("request", describeCertDetailSecretReq), slog.Any("response", describeCertDetailSecretResp))
			if err != nil {
				continue
			}

			// 如果已存在相同证书，直接返回
			oldCertPEM := strings.Join(describeCertDetailSecretResp.Result.SSL.Chain, "\n\n")
			if xcert.EqualCertificatesFromPEM(certPEM, oldCertPEM) {
				c.logger.Info("ssl certificate already exists")
				return &certmgr.UploadResult{
					CertId:   certItem.ChainID,
					CertName: certItem.CertName,
				}, nil
			}
		}
	}

	// 生成新证书名（需符合火山引擎命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 上传新证书
	// REF: https://www.volcengine.com/docs/6469/1186278#%E6%B7%BB%E5%8A%A0%E8%AF%81%E4%B9%A6
	createCertReq := &velive.CreateCertBody{
		CertName: ve.String(certName),
		Rsa: velive.CreateCertBodyRsa{
			Prikey: privkeyPEM,
			Pubkey: certPEM,
		},
		UseWay: "https",
	}
	createCertResp, err := c.sdkClient.CreateCert(ctx, createCertReq)
	c.logger.Debug("sdk request 'live.CreateCert'", slog.Any("request", createCertReq), slog.Any("response", createCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'live.CreateCert': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   *createCertResp.Result.ChainID,
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.ReplaceResult, error) {
	return nil, certmgr.ErrUnsupported
}
