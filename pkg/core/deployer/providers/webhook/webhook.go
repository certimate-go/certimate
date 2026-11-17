package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcertx509 "github.com/certimate-go/certimate/pkg/utils/cert/x509"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// Webhook URL。
	WebhookUrl string `json:"webhookUrl"`
	// Webhook 回调数据（application/json 或 application/x-www-form-urlencoded 格式）。
	WebhookData string `json:"webhookData,omitempty"`
	// 请求谓词。
	// 零值时默认值 "POST"。
	Method string `json:"method,omitempty"`
	// 请求标头。
	Headers map[string]string `json:"headers,omitempty"`
	// 请求超时（单位：秒）。
	// 零值时默认值 30。
	Timeout int `json:"timeout,omitempty"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	httpClient *resty.Client
}

var _ Provider = (*Deployer)(nil)

const (
	contentTypeJson      = "application/json"
	contentTypeForm      = "application/x-www-form-urlencoded"
	contentTypeMultipart = "multipart/form-data"
)

var allowedContentTypes = map[string]bool{
	contentTypeJson:      true,
	contentTypeForm:      true,
	contentTypeMultipart: true,
}

var allowedMethods = map[string]bool{
	http.MethodGet:    true,
	http.MethodPost:   true,
	http.MethodPut:    true,
	http.MethodPatch:  true,
	http.MethodDelete: true,
}

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client := resty.New().
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).
		AddRetryCondition(func(resp *resty.Response, _ error) bool {
			return resp == nil || resp.StatusCode() >= 500
		})
	if config.Timeout > 0 {
		client.SetTimeout(time.Duration(config.Timeout) * time.Second)
	}
	if config.AllowInsecureConnections {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return &Deployer{
		config:     config,
		logger:     slog.Default(),
		httpClient: client,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x509: %w", err)
	}

	// 提取服务器证书和中间证书
	serverCertPEM, issuerCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 处理 Webhook URL
	webhookUrl, err := url.Parse(d.config.WebhookUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook url: %w", err)
	} else if webhookUrl.Scheme != "http" && webhookUrl.Scheme != "https" {
		return nil, fmt.Errorf("unsupported webhook url scheme '%s'", webhookUrl.Scheme)
	}

	// 处理 Webhook 请求谓词
	webhookMethod := strings.ToUpper(d.config.Method)
	if webhookMethod == "" {
		webhookMethod = http.MethodPost
	} else if !allowedMethods[webhookMethod] {
		return nil, fmt.Errorf("unsupported webhook request method '%s'", webhookMethod)
	}

	// 处理 Webhook 请求标头
	webhookHeaders := make(http.Header)
	for k, v := range d.config.Headers {
		webhookHeaders.Set(k, v)
	}

	// 处理 Webhook 请求内容类型
	webhookContentType := webhookHeaders.Get("Content-Type")
	if webhookContentType == "" {
		webhookContentType = contentTypeJson
		webhookHeaders.Set("Content-Type", contentTypeJson)
	} else if mediaType, _, err := mime.ParseMediaType(webhookContentType); err != nil || !allowedContentTypes[mediaType] {
		return nil, fmt.Errorf("unsupported webhook content type '%s'", webhookContentType)
	}

	// 处理 Webhook 请求数据
	var webhookData any
	if d.config.WebhookData == "" {
		webhookData = map[string]string{
			"name":    strings.Join(xcertx509.GetSubjectAltNames(certX509), ";"),
			"cert":    certPEM,
			"privkey": privkeyPEM,
		}
	} else {
		err = json.Unmarshal([]byte(d.config.WebhookData), &webhookData)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal webhook data: %w", err)
		}

		if webhookMethod == http.MethodGet || webhookContentType == contentTypeForm || webhookContentType == contentTypeMultipart {
			temp := make(map[string]string)
			jsonb, err := json.Marshal(webhookData)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal webhook data: %w", err)
			} else if err := json.Unmarshal(jsonb, &temp); err != nil {
				return nil, fmt.Errorf("failed to unmarshal webhook data: %w", err)
			} else {
				webhookData = temp
			}
		}
	}

	// 替换变量值
	webhookUrl.Path = strings.ReplaceAll(webhookUrl.Path, "${CERTIMATE_DEPLOYER_COMMONNAME}", url.PathEscape(xcertx509.GetSubjectCommonName(certX509)))
	xmaps.DeepReplaceValueUnsafe(webhookData, "${CERTIMATE_DEPLOYER_COMMONNAME}", xcertx509.GetSubjectCommonName(certX509))
	xmaps.DeepReplaceValueUnsafe(webhookData, "${CERTIMATE_DEPLOYER_SUBJECTALTNAMES}", strings.Join(xcertx509.GetSubjectAltNames(certX509), ";"))
	xmaps.DeepReplaceValueUnsafe(webhookData, "${CERTIMATE_DEPLOYER_CERTIFICATE}", certPEM)
	xmaps.DeepReplaceValueUnsafe(webhookData, "${CERTIMATE_DEPLOYER_CERTIFICATE_SERVER}", serverCertPEM)
	xmaps.DeepReplaceValueUnsafe(webhookData, "${CERTIMATE_DEPLOYER_CERTIFICATE_INTERMEDIA}", issuerCertPEM)
	xmaps.DeepReplaceValueUnsafe(webhookData, "${CERTIMATE_DEPLOYER_PRIVATEKEY}", privkeyPEM)

	// 兼容旧版变量
	// TODO: remove in future version
	webhookUrl.Path = strings.ReplaceAll(webhookUrl.Path, "${DOMAIN}", url.PathEscape(certX509.Subject.CommonName))
	xmaps.DeepReplaceValueUnsafe(webhookData, "${DOMAIN}", certX509.Subject.CommonName)
	xmaps.DeepReplaceValueUnsafe(webhookData, "${DOMAINS}", strings.Join(certX509.DNSNames, ";"))
	xmaps.DeepReplaceValueUnsafe(webhookData, "${CERTIFICATE}", certPEM)
	xmaps.DeepReplaceValueUnsafe(webhookData, "${SERVER_CERTIFICATE}", serverCertPEM)
	xmaps.DeepReplaceValueUnsafe(webhookData, "${INTERMEDIA_CERTIFICATE}", issuerCertPEM)
	xmaps.DeepReplaceValueUnsafe(webhookData, "${PRIVATE_KEY}", privkeyPEM)

	// 生成请求
	// 其中 GET 请求需转换为查询参数
	req := d.httpClient.R().SetHeaderMultiValues(webhookHeaders)
	req.URL = webhookUrl.String()
	req.Method = webhookMethod
	if webhookMethod == http.MethodGet {
		req.SetQueryParams(webhookData.(map[string]string))
	} else {
		switch webhookContentType {
		case contentTypeJson:
			req.SetBody(webhookData)
		case contentTypeForm:
			req.SetFormData(webhookData.(map[string]string))
		case contentTypeMultipart:
			req.SetMultipartFormData(webhookData.(map[string]string))
		}
	}

	// 发送请求
	resp, err := req.Send()
	if err != nil {
		return nil, fmt.Errorf("failed to send webhook request: %w", err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("unexpected webhook response status code: %d", resp.StatusCode())
	}

	d.logger.Debug("webhook responded", slog.Any("response", resp.String()))

	return &DeployResult{}, nil
}
