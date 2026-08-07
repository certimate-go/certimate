package ucloudussl

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/certimate-go/certimate/pkg/core"
	ucloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/ussl"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// 优刻得 API 私钥。
	PrivateKey string `json:"privateKey"`
	// 优刻得 API 公钥。
	PublicKey string `json:"publicKey"`
	// 优刻得项目 ID。
	ProjectId string `json:"projectId,omitempty"`
	// 优刻得接口端点。
	Endpoint string `json:"endpoint,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *ucloudsdk.USSLClient
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.PrivateKey, config.PublicKey, config.ProjectId, config.Endpoint)
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
	// 生成新证书名（需符合优刻得命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 生成优刻得所需的证书参数
	certPEMBase64 := base64.StdEncoding.EncodeToString([]byte(certPEM))
	privkeyPEMBase64 := base64.StdEncoding.EncodeToString([]byte(privkeyPEM))
	certMd5 := md5.Sum([]byte(certPEMBase64 + privkeyPEMBase64))
	certMd5Hex := hex.EncodeToString(certMd5[:])

	// 上传托管证书
	// REF: https://docs.ucloud.cn/api/usslcertificate-api/upload_normal_certificate
	uploadNormalCertificateReq := c.sdkClient.NewUploadNormalCertificateRequest()
	uploadNormalCertificateReq.CertificateName = ucloud.String(certName)
	uploadNormalCertificateReq.SslPublicKey = ucloud.String(certPEMBase64)
	uploadNormalCertificateReq.SslPrivateKey = ucloud.String(privkeyPEMBase64)
	uploadNormalCertificateReq.SslMD5 = ucloud.String(certMd5Hex)
	uploadNormalCertificateResp, err := c.sdkClient.UploadNormalCertificate(uploadNormalCertificateReq)
	c.logger.Debug("sdk request 'ussl.UploadNormalCertificate'", slog.Any("request", uploadNormalCertificateReq), slog.Any("response", uploadNormalCertificateResp))
	if err != nil {
		if uploadNormalCertificateResp != nil && uploadNormalCertificateResp.GetRetCode() == 80035 {
			if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM); err != nil {
				return nil, err
			} else if !upok {
				return nil, fmt.Errorf("could not find ssl certificate, may be upload failed")
			} else {
				c.logger.Info("ssl certificate already exists")
				return upres, nil
			}
		}

		return nil, fmt.Errorf("failed to execute sdk request 'ussl.UploadNormalCertificate': %w", err)
	}

	return &UploadResult{
		CertId:   fmt.Sprintf("%d", uploadNormalCertificateResp.CertificateID),
		CertName: certName,
		ExtendedData: map[string]any{
			"ResourceId": uploadNormalCertificateResp.LongResourceID,
		},
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}

func (c *Certmgr) tryGetResultIfCertExists(ctx context.Context, certPEM string) (*UploadResult, bool, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, false, err
	}

	// 查询用户证书列表
	// REF: https://docs.ucloud.cn/api/usslcertificate-api/get_certificate_list
	// REF: https://docs.ucloud.cn/api/usslcertificate-api/get_certificate_detail_info
	// REF: https://docs.ucloud.cn/api/usslcertificate-api/download_certificate
	getCertificateListPage := 1
	getCertificateListPageSize := 1000
	for {
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		default:
		}

		getCertificateListReq := c.sdkClient.NewGetCertificateListRequest()
		getCertificateListReq.Mode = ucloud.String("trust")
		getCertificateListReq.Domain = ucloud.String(certX509.Subject.CommonName)
		getCertificateListReq.Sort = ucloud.String("2")
		getCertificateListReq.Page = ucloud.Int(getCertificateListPage)
		getCertificateListReq.PageSize = ucloud.Int(getCertificateListPageSize)
		getCertificateListResp, err := c.sdkClient.GetCertificateList(getCertificateListReq)
		c.logger.Debug("sdk request 'ussl.GetCertificateList'", slog.Any("request", getCertificateListReq), slog.Any("response", getCertificateListResp))
		if err != nil {
			return nil, false, fmt.Errorf("failed to execute sdk request 'ussl.GetCertificateList': %w", err)
		}

		for _, certItem := range getCertificateListResp.CertificateList {
			// 对比证书备用名称
			if len(certX509.DNSNames) == 0 || certItem.Domains != strings.Join(certX509.DNSNames, ",") {
				continue
			}

			// 对比证书有效期
			if certX509.NotBefore.UnixMilli() != certItem.NotBefore {
				continue
			} else if certX509.NotAfter.UnixMilli() != certItem.NotAfter {
				continue
			}

			// 对比证书签名算法
			getCertificateDetailInfoReq := c.sdkClient.NewGetCertificateDetailInfoRequest()
			getCertificateDetailInfoReq.CertificateID = ucloud.Int(certItem.CertificateID)
			getCertificateDetailInfoResp, err := c.sdkClient.GetCertificateDetailInfo(getCertificateDetailInfoReq)
			if err != nil {
				return nil, false, fmt.Errorf("failed to execute sdk request 'ussl.GetCertificateDetailInfo': %w", err)
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
				continue
			}

			// 对比证书内容
			downloadCertificateReq := c.sdkClient.NewDownloadCertificateRequest()
			downloadCertificateReq.CertificateID = ucloud.Int(certItem.CertificateID)
			downloadCertificateResp, err := c.sdkClient.DownloadCertificate(downloadCertificateReq)
			c.logger.Debug("sdk request 'ussl.DownloadCertificate'", slog.Any("request", downloadCertificateReq), slog.Any("response", downloadCertificateResp))
			if err != nil {
				return nil, false, fmt.Errorf("failed to execute sdk request 'ussl.DownloadCertificate': %w", err)
			} else {
				oldCertPEM, err := downloadCertificateFromUCloud(downloadCertificateResp.CertificateUrl)
				if err != nil {
					c.logger.Warn("could not download certificate from ucloud", slog.String("url", downloadCertificateResp.CertificateUrl), slog.Any("error", err))
				}
				if oldCertPEM != nil && !xcert.EqualCertificatesFromPEM(certPEM, string(oldCertPEM)) {
					continue
				}
			}

			return &UploadResult{
				CertId:   fmt.Sprintf("%d", certItem.CertificateID),
				CertName: certItem.Name,
				ExtendedData: map[string]any{
					"ResourceId": certItem.CertificateSN,
				},
			}, true, nil
		}

		if len(getCertificateListResp.CertificateList) < getCertificateListPageSize ||
			getCertificateListPage*getCertificateListPageSize >= getCertificateListResp.TotalCount {
			break
		}

		getCertificateListPage++
	}

	return nil, false, nil
}

func createSDKClient(privateKey, publicKey, projectId, endpoint string) (*ucloudsdk.USSLClient, error) {
	if privateKey == "" {
		return nil, fmt.Errorf("ucloud: invalid private key")
	}
	if publicKey == "" {
		return nil, fmt.Errorf("ucloud: invalid public key")
	}

	cfg := ucloud.NewConfig()
	if projectId != "" {
		cfg.ProjectId = projectId
	}
	if endpoint != "" {
		if strings.Contains(endpoint, "://") {
			cfg.BaseUrl = endpoint
		} else {
			cfg.BaseUrl = "https://" + endpoint
		}
	}

	credential := auth.NewCredential()
	credential.PrivateKey = privateKey
	credential.PublicKey = publicKey

	client := ucloudsdk.NewClient(&cfg, &credential)
	return client, nil
}

func downloadCertificateFromUCloud(url string) ([]byte, error) {
	url = strings.ReplaceAll(url, "\\u0026", "&")

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/zip") {
		return nil, fmt.Errorf("unexpected content type: %s", contentType)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	const targetFileName = "Nginx/public.pem"
	var pemData []byte

	for _, f := range reader.File {
		if strings.EqualFold(f.Name, "ALL/public.crt") || strings.EqualFold(f.Name, "Nginx/public.pem") {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			pemData, err = io.ReadAll(rc)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	if len(pemData) == 0 {
		return nil, fmt.Errorf("could not find the certificate file in the zip archive")
	}

	return pemData, nil
}
