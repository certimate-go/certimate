package wangsucertificate

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	wangsusdk "github.com/certimate-go/certimate/pkg/sdk3rd/wangsu/certificate"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 网宿云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 网宿云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *wangsusdk.Client
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
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
	// REF: https://www.wangsu.com/document/api-doc/22675?productCode=certificatemanagement
	listCertificatesResp, err := c.sdkClient.ListCertificatesWithContext(ctx)
	c.logger.Debug("sdk request 'certificatemanagement.ListCertificates'", slog.Any("response", listCertificatesResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'certificatemanagement.ListCertificates': %w", err)
	}

	if listCertificatesResp.Certificates != nil {
		for _, certItem := range listCertificatesResp.Certificates {
			// 对比证书序列号
			if !strings.EqualFold(certX509.SerialNumber.Text(16), certItem.Serial) {
				continue
			}

			// 对比证书有效期
			timezoneOfCST := time.FixedZone("CST", 8*60*60)
			oldCertNotBefore, _ := time.ParseInLocation(time.DateTime, certItem.ValidityFrom, timezoneOfCST)
			oldCertNotAfter, _ := time.ParseInLocation(time.DateTime, certItem.ValidityTo, timezoneOfCST)
			if !certX509.NotBefore.Equal(oldCertNotBefore) || !certX509.NotAfter.Equal(oldCertNotAfter) {
				continue
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &certmgr.UploadResult{
				CertId:   certItem.CertificateId,
				CertName: certItem.Name,
			}, nil
		}
	}

	// 生成新证书名（需符合网宿云命名规则）
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 新增证书
	// REF: https://www.wangsu.com/document/api-doc/25199?productCode=certificatemanagement
	createCertificateReq := &wangsusdk.CreateCertificateRequest{
		Name:        lo.ToPtr(certName),
		Certificate: lo.ToPtr(certPEM),
		PrivateKey:  lo.ToPtr(privkeyPEM),
		Comment:     lo.ToPtr("upload from certimate"),
	}
	createCertificateResp, err := c.sdkClient.CreateCertificateWithContext(ctx, createCertificateReq)
	c.logger.Debug("sdk request 'certificatemanagement.CreateCertificate'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'certificatemanagement.CreateCertificate': %w", err)
	}

	// 网宿云证书 URL 中包含证书 ID
	// 格式：
	//    https://open.chinanetcenter.com/api/certificate/100001
	wangsuCertIdMatches := regexp.MustCompile(`/certificate/([0-9]+)`).FindStringSubmatch(createCertificateResp.CertificateLocation)
	if len(wangsuCertIdMatches) == 0 {
		return nil, fmt.Errorf("received empty certificate id")
	}

	return &certmgr.UploadResult{
		CertId:   wangsuCertIdMatches[1],
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	certId := certIdOrName
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 修改证书
	// REF: https://www.wangsu.com/document/api-doc/25568?productCode=certificatemanagement
	updateCertificateReq := &wangsusdk.UpdateCertificateRequest{
		Name:        lo.ToPtr(certName),
		Certificate: lo.ToPtr(certPEM),
		PrivateKey:  lo.ToPtr(privkeyPEM),
		Comment:     lo.ToPtr("upload from certimate"),
	}
	updateCertificateResp, err := c.sdkClient.UpdateCertificateWithContext(ctx, certId, updateCertificateReq)
	c.logger.Debug("sdk request 'certificatemanagement.UpdateCertificate'", slog.Any("request", updateCertificateReq), slog.Any("response", updateCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'certificatemanagement.UpdateCertificate': %w", err)
	}

	return &certmgr.OperateResult{}, nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*wangsusdk.Client, error) {
	return wangsusdk.NewClient(accessKeyId, accessKeySecret)
}
