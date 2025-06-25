package ucloudussl

import (
	"context"
	"crypto/md5"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	ucloudauth "github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/certimate-go/certimate/pkg/core"
	usslsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/ussl"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type SSLManagerProviderConfig struct {
	// 优刻得 API 私钥。
	PrivateKey string `json:"privateKey"`
	// 优刻得 API 公钥。
	PublicKey string `json:"publicKey"`
	// 优刻得项目 ID。
	ProjectId string `json:"projectId,omitempty"`
}

type SSLManagerProvider struct {
	config    *SSLManagerProviderConfig
	logger    *slog.Logger
	sdkClient *usslsdk.USSLClient
}

var _ core.SSLManager = (*SSLManagerProvider)(nil)

func NewSSLManagerProvider(config *SSLManagerProviderConfig) (*SSLManagerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl manager provider is nil")
	}

	client, err := createSDKClient(config.PrivateKey, config.PublicKey)
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
	// 生成新证书名（需符合优刻得命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 生成优刻得所需的证书参数
	certPEMBase64 := base64.StdEncoding.EncodeToString([]byte(certPEM))
	privkeyPEMBase64 := base64.StdEncoding.EncodeToString([]byte(privkeyPEM))
	certMd5 := md5.Sum([]byte(certPEMBase64 + privkeyPEMBase64))
	certMd5Hex := hex.EncodeToString(certMd5[:])

	// 上传托管证书
	// REF: https://docs.ucloud.cn/api/usslcertificate-api/upload_normal_certificate
	uploadNormalCertificateReq := m.sdkClient.NewUploadNormalCertificateRequest()
	uploadNormalCertificateReq.CertificateName = ucloud.String(certName)
	uploadNormalCertificateReq.SslPublicKey = ucloud.String(certPEMBase64)
	uploadNormalCertificateReq.SslPrivateKey = ucloud.String(privkeyPEMBase64)
	uploadNormalCertificateReq.SslMD5 = ucloud.String(certMd5Hex)
	if m.config.ProjectId != "" {
		uploadNormalCertificateReq.ProjectId = ucloud.String(m.config.ProjectId)
	}
	uploadNormalCertificateResp, err := m.sdkClient.UploadNormalCertificate(uploadNormalCertificateReq)
	m.logger.Debug("sdk request 'ussl.UploadNormalCertificate'", slog.Any("request", uploadNormalCertificateReq), slog.Any("response", uploadNormalCertificateResp))
	if err != nil {
		if uploadNormalCertificateResp != nil && uploadNormalCertificateResp.GetRetCode() == 80035 {
			if res, err := m.findCertIfExists(ctx, certPEM); err != nil {
				return nil, err
			} else if res == nil {
				return nil, errors.New("ucloud ssl: no certificate found")
			} else {
				m.logger.Info("ssl certificate already exists")
				return res, nil
			}
		}

		return nil, fmt.Errorf("failed to execute sdk request 'ussl.UploadNormalCertificate': %w", err)
	}

	return &core.SSLManageUploadResult{
		CertId:   fmt.Sprintf("%d", uploadNormalCertificateResp.CertificateID),
		CertName: certName,
		ExtendedData: map[string]any{
			"resourceId": uploadNormalCertificateResp.LongResourceID,
		},
	}, nil
}

func (m *SSLManagerProvider) findCertIfExists(ctx context.Context, certPEM string) (*core.SSLManageUploadResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 遍历获取用户证书列表
	// REF: https://docs.ucloud.cn/api/usslcertificate-api/get_certificate_list
	// REF: https://docs.ucloud.cn/api/usslcertificate-api/download_certificate
	getCertificateListPage := int(1)
	getCertificateListLimit := int(1000)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		getCertificateListReq := m.sdkClient.NewGetCertificateListRequest()
		getCertificateListReq.Mode = ucloud.String("trust")
		getCertificateListReq.Domain = ucloud.String(certX509.Subject.CommonName)
		getCertificateListReq.Sort = ucloud.String("2")
		getCertificateListReq.Page = ucloud.Int(getCertificateListPage)
		getCertificateListReq.PageSize = ucloud.Int(getCertificateListLimit)
		if m.config.ProjectId != "" {
			getCertificateListReq.ProjectId = ucloud.String(m.config.ProjectId)
		}
		getCertificateListResp, err := m.sdkClient.GetCertificateList(getCertificateListReq)
		m.logger.Debug("sdk request 'ussl.GetCertificateList'", slog.Any("request", getCertificateListReq), slog.Any("response", getCertificateListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'ussl.GetCertificateList': %w", err)
		}

		if getCertificateListResp.CertificateList != nil {
			for _, certItem := range getCertificateListResp.CertificateList {
				// 优刻得未提供可唯一标识证书的字段，只能通过多个字段尝试对比来判断是否为同一证书
				// 先分别对比证书的多域名、品牌、有效期，再对比签名算法

				if len(certX509.DNSNames) == 0 || certItem.Domains != strings.Join(certX509.DNSNames, ",") {
					continue
				}

				if len(certX509.Issuer.Organization) == 0 || certItem.Brand != certX509.Issuer.Organization[0] {
					continue
				}

				if int64(certItem.NotBefore) != certX509.NotBefore.UnixMilli() || int64(certItem.NotAfter) != certX509.NotAfter.UnixMilli() {
					continue
				}

				getCertificateDetailInfoReq := m.sdkClient.NewGetCertificateDetailInfoRequest()
				getCertificateDetailInfoReq.CertificateID = ucloud.Int(certItem.CertificateID)
				if m.config.ProjectId != "" {
					getCertificateDetailInfoReq.ProjectId = ucloud.String(m.config.ProjectId)
				}
				getCertificateDetailInfoResp, err := m.sdkClient.GetCertificateDetailInfo(getCertificateDetailInfoReq)
				if err != nil {
					return nil, fmt.Errorf("failed to execute sdk request 'ussl.GetCertificateDetailInfo': %w", err)
				}

				switch certX509.SignatureAlgorithm {
				case x509.SHA256WithRSA:
					if !strings.EqualFold(getCertificateDetailInfoResp.CertificateInfo.Algorithm, "SHA256-RSA") {
						continue
					}
				case x509.SHA384WithRSA:
					if !strings.EqualFold(getCertificateDetailInfoResp.CertificateInfo.Algorithm, "SHA384-RSA") {
						continue
					}
				case x509.SHA512WithRSA:
					if !strings.EqualFold(getCertificateDetailInfoResp.CertificateInfo.Algorithm, "SHA512-RSA") {
						continue
					}
				case x509.SHA256WithRSAPSS:
					if !strings.EqualFold(getCertificateDetailInfoResp.CertificateInfo.Algorithm, "SHA256-RSAPSS") {
						continue
					}
				case x509.SHA384WithRSAPSS:
					if !strings.EqualFold(getCertificateDetailInfoResp.CertificateInfo.Algorithm, "SHA384-RSAPSS") {
						continue
					}
				case x509.SHA512WithRSAPSS:
					if !strings.EqualFold(getCertificateDetailInfoResp.CertificateInfo.Algorithm, "SHA512-RSAPSS") {
						continue
					}
				case x509.ECDSAWithSHA256:
					if !strings.EqualFold(getCertificateDetailInfoResp.CertificateInfo.Algorithm, "ECDSA-SHA256") {
						continue
					}
				case x509.ECDSAWithSHA384:
					if !strings.EqualFold(getCertificateDetailInfoResp.CertificateInfo.Algorithm, "ECDSA-SHA384") {
						continue
					}
				case x509.ECDSAWithSHA512:
					if !strings.EqualFold(getCertificateDetailInfoResp.CertificateInfo.Algorithm, "ECDSA-SHA512") {
						continue
					}
				default:
					// 未知签名算法，跳过
					continue
				}

				return &core.SSLManageUploadResult{
					CertId:   fmt.Sprintf("%d", certItem.CertificateID),
					CertName: certItem.Name,
					ExtendedData: map[string]any{
						"resourceId": certItem.CertificateSN,
					},
				}, nil
			}
		}

		if getCertificateListResp.CertificateList == nil || len(getCertificateListResp.CertificateList) < int(getCertificateListLimit) {
			break
		} else {
			getCertificateListPage++
		}
	}

	return nil, nil
}

func createSDKClient(privateKey, publicKey string) (*usslsdk.USSLClient, error) {
	cfg := ucloud.NewConfig()

	credential := ucloudauth.NewCredential()
	credential.PrivateKey = privateKey
	credential.PublicKey = publicKey

	client := usslsdk.NewClient(&cfg, &credential)
	return client, nil
}
