package baiducloudcert

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/baidubce/bce-sdk-go/bce"

	"github.com/certimate-go/certimate/pkg/core"
	baiducert "github.com/certimate-go/certimate/pkg/sdk3rd/baiducloud/cert"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// 百度智能云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 百度智能云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *baiducert.Client
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
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

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*UploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 查看证书列表
	// REF: https://cloud.baidu.com/doc/Reference/s/Gjwvz27xu#34-%E6%9F%A5%E7%9C%8B%E8%AF%81%E4%B9%A6%E5%88%97%E8%A1%A8
	// REF: https://cloud.baidu.com/doc/Reference/s/Gjwvz27xu#35-%E6%9F%A5%E7%9C%8B%E8%AF%81%E4%B9%A6%E5%88%97%E8%A1%A8%E8%AF%A6%E6%83%85
	listCertDetail, err := c.sdkClient.ListCertDetail()
	c.logger.Debug("sdk request 'cert.ListCertDetail'", slog.Any("response", listCertDetail))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cert.ListCertDetail': %w", err)
	} else {
		for _, certItem := range listCertDetail.Certs {
			// 对比证书通用名称
			if !strings.EqualFold(certX509.Subject.CommonName, certItem.CertCommonName) {
				continue
			}

			// 对比证书备用名称
			if certItem.CertDNSNames != strings.Join(certX509.DNSNames, ",") {
				continue
			}

			// 对比证书有效期
			newCertNotBefore := certX509.NotBefore
			newCertNotAfter := certX509.NotAfter
			oldCertNotBefore, _ := time.Parse("2006-01-02T15:04:05Z", certItem.CertStartTime)
			oldCertNotAfter, _ := time.Parse("2006-01-02T15:04:05Z", certItem.CertStopTime)
			if !newCertNotBefore.Equal(oldCertNotBefore) || !newCertNotAfter.Equal(oldCertNotAfter) {
				continue
			}

			// 对比证书内容
			getCertDetailResp, err := c.sdkClient.GetCertRawData(certItem.CertId)
			c.logger.Debug("sdk request 'cert.GetCertRawData'", slog.String("params.certId", certItem.CertId), slog.Any("response", getCertDetailResp))
			if err != nil {
				if sdkErr, ok := err.(*bce.BceServiceError); ok {
					if sdkErrCode := sdkErr.Code; sdkErrCode == "ResourceNotFoundException" {
						continue
					}
				}

				return nil, fmt.Errorf("failed to execute sdk request 'cert.GetCertRawData': %w", err)
			} else {
				if !xcert.EqualCertificatesFromPEM(certPEM, getCertDetailResp.CertServerData) {
					continue
				}
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &UploadResult{
				CertId:   certItem.CertId,
				CertName: certItem.CertName,
			}, nil
		}
	}

	// 创建证书
	// REF: https://cloud.baidu.com/doc/Reference/s/Gjwvz27xu#31-%E5%88%9B%E5%BB%BA%E8%AF%81%E4%B9%A6
	createCertReq := &baiducert.CreateCertArgs{}
	createCertReq.CertName = fmt.Sprintf("certimate-%d", time.Now().UnixMilli())
	createCertReq.CertServerData = certPEM
	createCertReq.CertPrivateData = privkeyPEM
	createCertResp, err := c.sdkClient.CreateCert(createCertReq)
	c.logger.Debug("sdk request 'cert.CreateCert'", slog.Any("request", createCertReq), slog.Any("response", createCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cert.CreateCert': %w", err)
	}

	return &UploadResult{
		CertId:   createCertResp.CertId,
		CertName: createCertResp.CertName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}

func createSDKClient(accessKeyId, secretAccessKey string) (*baiducert.Client, error) {
	client, err := baiducert.NewClient(accessKeyId, secretAccessKey, "")
	if err != nil {
		return nil, err
	}

	return client, nil
}
