package aliyuncas

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	alicas "github.com/alibabacloud-go/cas-20200407/v3/client"
	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xtypes "github.com/certimate-go/certimate/pkg/utils/types"
)

type SSLManagerProviderConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 阿里云地域。
	Region string `json:"region"`
}

type SSLManagerProvider struct {
	config    *SSLManagerProviderConfig
	logger    *slog.Logger
	sdkClient *alicas.Client
}

var _ core.SSLManager = (*SSLManagerProvider)(nil)

func NewSSLManagerProvider(config *SSLManagerProviderConfig) (*SSLManagerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl manager provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	return &SSLManagerProvider{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (m *SSLManagerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		m.logger = slog.New(slog.DiscardHandler)
	} else {
		m.logger = logger
	}
}

func (m *SSLManagerProvider) Upload(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLManageUploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 查询证书列表，避免重复上传
	// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-listusercertificateorder
	// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-getusercertificatedetail
	listUserCertificateOrderPage := int64(1)
	listUserCertificateOrderLimit := int64(50)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listUserCertificateOrderReq := &alicas.ListUserCertificateOrderRequest{
			ResourceGroupId: xtypes.ToPtrOrZeroNil(m.config.ResourceGroupId),
			CurrentPage:     tea.Int64(listUserCertificateOrderPage),
			ShowSize:        tea.Int64(listUserCertificateOrderLimit),
			OrderType:       tea.String("CERT"),
		}
		listUserCertificateOrderResp, err := m.sdkClient.ListUserCertificateOrder(listUserCertificateOrderReq)
		m.logger.Debug("sdk request 'cas.ListUserCertificateOrder'", slog.Any("request", listUserCertificateOrderReq), slog.Any("response", listUserCertificateOrderResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cas.ListUserCertificateOrder': %w", err)
		}

		if listUserCertificateOrderResp.Body.CertificateOrderList != nil {
			for _, certOrder := range listUserCertificateOrderResp.Body.CertificateOrderList {
				// 先对比证书通用名称
				if !strings.EqualFold(certX509.Subject.CommonName, tea.StringValue(certOrder.CommonName)) {
					continue
				}

				// 再对比证书序列号
				// 注意阿里云 CAS 会在序列号前补零，需去除后再比较
				oldCertSN := strings.TrimLeft(tea.StringValue(certOrder.SerialNo), "0")
				newCertSN := strings.TrimLeft(certX509.SerialNumber.Text(16), "0")
				if !strings.EqualFold(newCertSN, oldCertSN) {
					continue
				}

				// 最后对比证书内容
				getUserCertificateDetailReq := &alicas.GetUserCertificateDetailRequest{
					CertId: certOrder.CertificateId,
				}
				getUserCertificateDetailResp, err := m.sdkClient.GetUserCertificateDetail(getUserCertificateDetailReq)
				m.logger.Debug("sdk request 'cas.GetUserCertificateDetail'", slog.Any("request", getUserCertificateDetailReq), slog.Any("response", getUserCertificateDetailResp))
				if err != nil {
					return nil, fmt.Errorf("failed to execute sdk request 'cas.GetUserCertificateDetail': %w", err)
				}

				var isSameCert bool
				if *getUserCertificateDetailResp.Body.Cert == certPEM {
					isSameCert = true
				} else {
					oldCertX509, err := xcert.ParseCertificateFromPEM(*getUserCertificateDetailResp.Body.Cert)
					if err != nil {
						continue
					}

					isSameCert = xcert.EqualCertificate(certX509, oldCertX509)
				}

				// 如果已存在相同证书，直接返回
				if isSameCert {
					m.logger.Info("ssl certificate already exists")
					return &core.SSLManageUploadResult{
						CertId:   fmt.Sprintf("%d", tea.Int64Value(certOrder.CertificateId)),
						CertName: *certOrder.Name,
						ExtendedData: map[string]any{
							"instanceId":     tea.StringValue(getUserCertificateDetailResp.Body.InstanceId),
							"certIdentifier": tea.StringValue(getUserCertificateDetailResp.Body.CertIdentifier),
						},
					}, nil
				}
			}
		}

		if listUserCertificateOrderResp.Body.CertificateOrderList == nil || len(listUserCertificateOrderResp.Body.CertificateOrderList) < int(listUserCertificateOrderLimit) {
			break
		} else {
			listUserCertificateOrderPage++
		}
	}

	// 生成新证书名（需符合阿里云命名规则）
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 上传新证书
	// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-uploadusercertificate
	uploadUserCertificateReq := &alicas.UploadUserCertificateRequest{
		ResourceGroupId: xtypes.ToPtrOrZeroNil(m.config.ResourceGroupId),
		Name:            tea.String(certName),
		Cert:            tea.String(certPEM),
		Key:             tea.String(privkeyPEM),
	}
	uploadUserCertificateResp, err := m.sdkClient.UploadUserCertificate(uploadUserCertificateReq)
	m.logger.Debug("sdk request 'cas.UploadUserCertificate'", slog.Any("request", uploadUserCertificateReq), slog.Any("response", uploadUserCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cas.UploadUserCertificate': %w", err)
	}

	// 获取证书详情
	// REF: https://help.aliyun.com/zh/ssl-certificate/developer-reference/api-cas-2020-04-07-getusercertificatedetail
	getUserCertificateDetailReq := &alicas.GetUserCertificateDetailRequest{
		CertId:     uploadUserCertificateResp.Body.CertId,
		CertFilter: tea.Bool(true),
	}
	getUserCertificateDetailResp, err := m.sdkClient.GetUserCertificateDetail(getUserCertificateDetailReq)
	m.logger.Debug("sdk request 'cas.GetUserCertificateDetail'", slog.Any("request", getUserCertificateDetailReq), slog.Any("response", getUserCertificateDetailResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'cas.GetUserCertificateDetail': %w", err)
	}

	return &core.SSLManageUploadResult{
		CertId:   fmt.Sprintf("%d", tea.Int64Value(getUserCertificateDetailResp.Body.Id)),
		CertName: certName,
		ExtendedData: map[string]any{
			"instanceId":     tea.StringValue(getUserCertificateDetailResp.Body.InstanceId),
			"certIdentifier": tea.StringValue(getUserCertificateDetailResp.Body.CertIdentifier),
		},
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
