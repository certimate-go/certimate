package huaweicloudobs

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/certimate-go/certimate/pkg/core/deployer"
)

type DeployerConfig struct {
	// 华为云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 华为云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 华为云区域。
	Region string `json:"region"`
	// 存储桶名。
	Bucket string `json:"bucket"`
	// 自定义域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config *DeployerConfig
	logger *slog.Logger
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	return &Deployer{
		config: config,
		logger: slog.Default(),
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	if d.config.Region == "" {
		return nil, fmt.Errorf("config `region` is required")
	}
	if d.config.Bucket == "" {
		return nil, fmt.Errorf("config `bucket` is required")
	}
	if d.config.Domain == "" {
		return nil, fmt.Errorf("config `domain` is required")
	}

	// REF: https://support.huaweicloud.com/usermanual-obs/obs_06_3200.html
	// REF: https://support.huaweicloud.com/api-obs/obs_04_0059.html
	url := fmt.Sprintf("https://%s.obs.%s.myhuaweicloud.com/?customdomain=%s", d.config.Bucket, d.config.Region, d.config.Domain)
	bodyXML := fmt.Sprintf(`
<CustomDomainConfiguration>
	<Name>%s</Name>
	<Certificate>%s</Certificate>
	<CertificateChain>%s</CertificateChain>
	<PrivateKey>%s</PrivateKey>
</CustomDomainConfiguration>`,
		d.config.Bucket+"_"+d.config.Domain, certPEM, certPEM, privkeyPEM,
	)

	// 计算 Content-MD5（Base64 编码）
	md5sum := md5.Sum([]byte(bodyXML))
	md5sumEncoded := base64.StdEncoding.EncodeToString(md5sum[:])

	// 构造签名字符串
	date := time.Now().UTC().Format(http.TimeFormat)
	method := "PUT"
	contentType := "application/xml"
	canonicalizedResource := fmt.Sprintf("/%s/?customdomain=%s", d.config.Bucket, d.config.Domain)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", method, md5sumEncoded, contentType, date, canonicalizedResource)

	// HMAC-SHA1 签名
	h := hmac.New(sha1.New, []byte(d.config.SecretAccessKey))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Authorization
	authHeader := fmt.Sprintf("OBS %s:%s", d.config.AccessKeyId, signature)

	// 创建请求
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(bodyXML)))
	if err != nil {
		return nil, fmt.Errorf("huaweicloud obs api error: %w", err)
	}
	req.Header.Set("Date", date)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-MD5", md5sumEncoded)
	req.Header.Set("Content-Type", contentType)

	// 请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("huaweicloud obs api error: %w", err)
	}
	defer resp.Body.Close()

	// 响应
	if resp.StatusCode != http.StatusOK {
		body := &bytes.Buffer{}
		body.ReadFrom(resp.Body)
		return nil, fmt.Errorf("huaweicloud obs api error: unexpected status code: %d (resp: %s)", resp.StatusCode, body.String())
	}

	return &deployer.DeployResult{}, nil
}
