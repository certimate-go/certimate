package rainyunsslcenter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	rainyunsdk "github.com/certimate-go/certimate/pkg/sdk3rd/rainyun"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 雨云 API 密钥。
	ApiKey string `json:"ApiKey"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *rainyunsdk.Client
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.ApiKey)
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
	// 避免重复上传
	if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM); err != nil {
		return nil, err
	} else if upok {
		c.logger.Info("ssl certificate already exists")
		return upres, nil
	}

	// SSL 证书上传
	// REF: https://apifox.com/apidoc/shared/a4595cc8-44c5-4678-a2a3-eed7738dab03/api-69943046
	sslCenterCreateReq := &rainyunsdk.SslCenterCreateRequest{
		Cert: certPEM,
		Key:  privkeyPEM,
	}
	sslCenterCreateResp, err := c.sdkClient.SslCenterCreateWithContext(ctx, sslCenterCreateReq)
	c.logger.Debug("sdk request 'sslcenter.Create'", slog.Any("request", sslCenterCreateReq), slog.Any("response", sslCenterCreateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'sslcenter.Create': %w", err)
	}

	// 获取刚刚上传证书 ID
	if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM); err != nil {
		return nil, err
	} else if !upok {
		return nil, errors.New("could not find ssl certificate, may be upload failed")
	} else {
		return upres, nil
	}
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	certId, err := strconv.ParseInt(certIdOrName, 10, 64)
	if err != nil {
		return nil, err
	}

	// SSL 证书替换操作
	// REF: https://s.apifox.cn/a4595cc8-44c5-4678-a2a3-eed7738dab03/api-69943049
	sslCenterUpdateReq := &rainyunsdk.SslCenterUpdateRequest{
		Cert: certPEM,
		Key:  privkeyPEM,
	}
	sslCenterUpdateResp, err := c.sdkClient.SslCenterUpdateWithContext(ctx, certId, sslCenterUpdateReq)
	c.logger.Debug("sdk request 'sslcenter.Update'", slog.Int64("certId", certId), slog.Any("request", sslCenterUpdateReq), slog.Any("response", sslCenterUpdateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'sslcenter.Update': %w", err)
	}

	return &certmgr.OperateResult{}, nil
}

func (c *Certmgr) tryGetResultIfCertExists(ctx context.Context, certPEM string) (*certmgr.UploadResult, bool, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, false, err
	}

	// 获取 SSL 证书列表
	// REF: https://apifox.com/apidoc/shared/a4595cc8-44c5-4678-a2a3-eed7738dab03/api-69943046
	// REF: https://apifox.com/apidoc/shared/a4595cc8-44c5-4678-a2a3-eed7738dab03/api-69943048
	sslCenterListPage := 1
	sslCenterListPerPage := 100
	for {
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		default:
		}

		sslCenterListReq := &rainyunsdk.SslCenterListRequest{
			Filters: &rainyunsdk.SslCenterListFilters{
				Domain: &certX509.Subject.CommonName,
			},
			Page:    lo.ToPtr(int32(sslCenterListPage)),
			PerPage: lo.ToPtr(int32(sslCenterListPerPage)),
		}
		sslCenterListResp, err := c.sdkClient.SslCenterListWithContext(ctx, sslCenterListReq)
		c.logger.Debug("sdk request 'sslcenter.List'", slog.Any("request", sslCenterListReq), slog.Any("response", sslCenterListResp))
		if err != nil {
			return nil, false, fmt.Errorf("failed to execute sdk request 'sslcenter.List': %w", err)
		}

		if sslCenterListResp.Data == nil {
			break
		}

		for _, sslItem := range sslCenterListResp.Data.Records {
			// 对比证书的多域名
			if sslItem.Domain != strings.Join(certX509.DNSNames, ", ") {
				continue
			}

			// 对比证书的有效期
			if sslItem.StartDate != certX509.NotBefore.Unix() || sslItem.ExpireDate != certX509.NotAfter.Unix() {
				continue
			}

			// 对比证书内容
			sslCenterGetResp, err := c.sdkClient.SslCenterGetWithContext(ctx, sslItem.ID)
			if err != nil {
				return nil, false, fmt.Errorf("failed to execute sdk request 'sslcenter.Get': %w", err)
			} else {
				if !xcert.EqualCertificatesFromPEM(certPEM, sslCenterGetResp.Data.Cert) {
					continue
				}
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			return &certmgr.UploadResult{
				CertId: fmt.Sprintf("%d", sslItem.ID),
			}, true, nil
		}

		if len(sslCenterListResp.Data.Records) < sslCenterListPerPage {
			break
		}

		sslCenterListPage++
	}

	return nil, false, nil
}

func createSDKClient(apiKey string) (*rainyunsdk.Client, error) {
	return rainyunsdk.NewClient(apiKey)
}
