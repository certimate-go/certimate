package ctcccloudlvdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	ctyunlvdn "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/lvdn"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 天翼云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 天翼云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *ctyunlvdn.Client
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
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 查询证书列表，避免重复上传
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=125&api=11452&data=183&isNormal=1&vid=261
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=125&api=11449&data=183&isNormal=1&vid=261
	queryCertListPage := 1
	queryCertListPerPage := 1000
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		queryCertListReq := &ctyunlvdn.QueryCertListRequest{
			Page:      lo.ToPtr(int32(queryCertListPage)),
			PerPage:   lo.ToPtr(int32(queryCertListPerPage)),
			UsageMode: lo.ToPtr(int32(0)),
		}
		queryCertListResp, err := c.sdkClient.QueryCertListWithContext(ctx, queryCertListReq)
		c.logger.Debug("sdk request 'lvdn.QueryCertList'", slog.Any("request", queryCertListReq), slog.Any("response", queryCertListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'lvdn.QueryCertList': %w", err)
		}

		if queryCertListResp.ReturnObj == nil {
			break
		}

		for _, certItem := range queryCertListResp.ReturnObj.Results {
			// 对比证书通用名称
			if !strings.EqualFold(certX509.Subject.CommonName, certItem.CN) {
				continue
			}

			// 对比证书扩展名称
			if !slices.Equal(certX509.DNSNames, certItem.SANs) {
				continue
			}

			// 对比证书有效期
			if !certX509.NotBefore.Equal(time.Unix(certItem.IssueTime, 0).UTC()) {
				continue
			} else if !certX509.NotAfter.Equal(time.Unix(certItem.ExpiresTime, 0).UTC()) {
				continue
			}

			// 对比证书内容
			queryCertDetailReq := &ctyunlvdn.QueryCertDetailRequest{
				Id: lo.ToPtr(certItem.Id),
			}
			queryCertDetailResp, err := c.sdkClient.QueryCertDetailWithContext(ctx, queryCertDetailReq)
			c.logger.Debug("sdk request 'lvdn.QueryCertDetail'", slog.Any("request", queryCertDetailReq), slog.Any("response", queryCertDetailResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'lvdn.QueryCertDetail': %w", err)
			} else if queryCertDetailResp.ReturnObj != nil && queryCertDetailResp.ReturnObj.Result != nil {
				if !xcert.EqualCertificatesFromPEM(certPEM, queryCertDetailResp.ReturnObj.Result.Certs) {
					continue
				}
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &certmgr.UploadResult{
				CertId:   fmt.Sprintf("%d", queryCertDetailResp.ReturnObj.Result.Id),
				CertName: queryCertDetailResp.ReturnObj.Result.Name,
			}, nil
		}

		if len(queryCertListResp.ReturnObj.Results) < queryCertListPerPage {
			break
		}

		queryCertListPage++
	}

	// 生成新证书名（需符合天翼云命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 创建证书
	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=125&api=11436&data=183&isNormal=1&vid=261
	createCertReq := &ctyunlvdn.CreateCertRequest{
		Name:  lo.ToPtr(certName),
		Certs: lo.ToPtr(certPEM),
		Key:   lo.ToPtr(privkeyPEM),
	}
	createCertResp, err := c.sdkClient.CreateCertWithContext(ctx, createCertReq)
	c.logger.Debug("sdk request 'lvdn.CreateCert'", slog.Any("request", createCertReq), slog.Any("response", createCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'lvdn.CreateCert': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   fmt.Sprintf("%d", createCertResp.ReturnObj.Id),
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
}

func createSDKClient(accessKeyId, secretAccessKey string) (*ctyunlvdn.Client, error) {
	return ctyunlvdn.NewClient(accessKeyId, secretAccessKey)
}
