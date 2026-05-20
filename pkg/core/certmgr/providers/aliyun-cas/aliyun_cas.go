package aliyuncas

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	alicas "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/alibabacloud-go/cas-20200407/v4/client"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 阿里云地域。
	Region string `json:"region"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *alicas.Client
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
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
	// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-listusercertificateorder
	// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-getusercertificatedetail
	listUserCertificateOrderPage := 1
	listUserCertificateOrderLimit := 50
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listUserCertificateOrderReq := &alicas.ListUserCertificateOrderRequest{
			ResourceGroupId: lo.EmptyableToPtr(c.config.ResourceGroupId),
			CurrentPage:     tea.Int64(int64(listUserCertificateOrderPage)),
			ShowSize:        tea.Int64(int64(listUserCertificateOrderLimit)),
			OrderType:       tea.String("CERT"),
		}
		listUserCertificateOrderResp, err := c.sdkClient.ListUserCertificateOrderWithContext(ctx, listUserCertificateOrderReq, &dara.RuntimeOptions{})
		c.logger.Debug("sdk request 'cas.ListUserCertificateOrder'", slog.Any("request", listUserCertificateOrderReq), slog.Any("response", listUserCertificateOrderResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cas.ListUserCertificateOrder': %w", err)
		}

		if listUserCertificateOrderResp.Body == nil {
			break
		}

		for _, certItem := range listUserCertificateOrderResp.Body.CertificateOrderList {
			// 对比证书通用名称
			if !strings.EqualFold(certX509.Subject.CommonName, tea.StringValue(certItem.CommonName)) {
				continue
			}

			// 对比证书序列号
			// 注意阿里云 CAS 会在序列号前补零，需去除后再比较
			oldCertSN := strings.TrimPrefix(tea.StringValue(certItem.SerialNo), "0")
			newCertSN := strings.TrimPrefix(certX509.SerialNumber.Text(16), "0")
			if !strings.EqualFold(newCertSN, oldCertSN) {
				continue
			}

			// 对比证书内容
			getUserCertificateDetailReq := &alicas.GetUserCertificateDetailRequest{
				CertId: certItem.CertificateId,
			}
			getUserCertificateDetailResp, err := c.sdkClient.GetUserCertificateDetailWithContext(ctx, getUserCertificateDetailReq, &dara.RuntimeOptions{})
			c.logger.Debug("sdk request 'cas.GetUserCertificateDetail'", slog.Any("request", getUserCertificateDetailReq), slog.Any("response", getUserCertificateDetailResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'cas.GetUserCertificateDetail': %w", err)
			} else {
				if !xcert.EqualCertificatesFromPEM(certPEM, tea.StringValue(getUserCertificateDetailResp.Body.Cert)) {
					continue
				}
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			c.logger.Info("ssl certificate already exists")
			return &certmgr.UploadResult{
				CertId:   fmt.Sprintf("%d", tea.Int64Value(certItem.CertificateId)),
				CertName: tea.StringValue(certItem.Name),
				ExtendedData: map[string]any{
					"InstanceId":     tea.StringValue(getUserCertificateDetailResp.Body.InstanceId),
					"CertIdentifier": tea.StringValue(getUserCertificateDetailResp.Body.CertIdentifier),
				},
			}, nil
		}

		if len(listUserCertificateOrderResp.Body.CertificateOrderList) < listUserCertificateOrderLimit {
			break
		}

		listUserCertificateOrderPage++
	}

	// 生成新证书名（需符合阿里云命名规则）
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 上传新证书
	// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-uploadusercertificate
	uploadUserCertificateReq := &alicas.UploadUserCertificateRequest{
		ResourceGroupId: lo.EmptyableToPtr(c.config.ResourceGroupId),
		Name:            tea.String(certName),
		Cert:            tea.String(certPEM),
		Key:             tea.String(privkeyPEM),
	}
	uploadUserCertificateResp, err := c.sdkClient.UploadUserCertificateWithContext(ctx, uploadUserCertificateReq, &dara.RuntimeOptions{})
	c.logger.Debug("sdk request 'cas.UploadUserCertificate'", slog.Any("request", uploadUserCertificateReq), slog.Any("response", uploadUserCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cas.UploadUserCertificate': %w", err)
	}

	// 获取证书详情
	// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-getusercertificatedetail
	getUserCertificateDetailReq := &alicas.GetUserCertificateDetailRequest{
		CertId:     uploadUserCertificateResp.Body.CertId,
		CertFilter: tea.Bool(true),
	}
	getUserCertificateDetailResp, err := c.sdkClient.GetUserCertificateDetailWithContext(ctx, getUserCertificateDetailReq, &dara.RuntimeOptions{})
	c.logger.Debug("sdk request 'cas.GetUserCertificateDetail'", slog.Any("request", getUserCertificateDetailReq), slog.Any("response", getUserCertificateDetailResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cas.GetUserCertificateDetail': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   fmt.Sprintf("%d", tea.Int64Value(getUserCertificateDetailResp.Body.Id)),
		CertName: certName,
		ExtendedData: map[string]any{
			"InstanceId":     tea.StringValue(getUserCertificateDetailResp.Body.InstanceId),
			"CertIdentifier": tea.StringValue(getUserCertificateDetailResp.Body.CertIdentifier),
		},
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
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
