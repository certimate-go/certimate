package onepanelssl

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	onepanelsdk "github.com/certimate-go/certimate/pkg/sdk3rd/1panel"
	onepanelsdk2 "github.com/certimate-go/certimate/pkg/sdk3rd/1panel/v2"
)

type CertmgrConfig struct {
	// 1Panel 服务地址。
	ServerUrl string `json:"serverUrl"`
	// 1Panel 版本。
	ApiVersion string `json:"apiVersion"`
	// 1Panel 接口密钥。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 子节点名称。
	// 选填。
	NodeName string `json:"nodeName,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient any
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiVersion, config.ApiKey, config.AllowInsecureConnections, config.NodeName)
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
	if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM, privkeyPEM); err != nil {
		return nil, err
	} else if upok {
		c.logger.Info("ssl certificate already exists")
		return upres, nil
	}

	// 生成新证书名（需符合 1Panel 命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 上传证书
	switch sdkClient := c.sdkClient.(type) {
	case *onepanelsdk.Client:
		{
			websiteSSLUploadReq := &onepanelsdk.WebsiteSSLUploadRequest{
				Type:        "paste",
				Description: certName,
				Certificate: certPEM,
				PrivateKey:  privkeyPEM,
			}
			websiteSSLUploadResp, err := sdkClient.WebsiteSSLUpload(websiteSSLUploadReq)
			c.logger.Debug("sdk request '1panel.WebsiteSSLUpload'", slog.Any("request", websiteSSLUploadReq), slog.Any("response", websiteSSLUploadResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request '1panel.WebsiteSSLUpload': %w", err)
			}
		}

	case *onepanelsdk2.Client:
		{
			websiteSSLUploadReq := &onepanelsdk2.WebsiteSSLUploadRequest{
				Type:        "paste",
				Description: certName,
				Certificate: certPEM,
				PrivateKey:  privkeyPEM,
			}
			websiteSSLUploadResp, err := sdkClient.WebsiteSSLUpload(websiteSSLUploadReq)
			c.logger.Debug("sdk request '1panel.WebsiteSSLUpload'", slog.Any("request", websiteSSLUploadReq), slog.Any("response", websiteSSLUploadResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request '1panel.WebsiteSSLUpload': %w", err)
			}
		}

	default:
		panic("unreachable")
	}

	// 获取刚刚上传证书 ID
	if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM, privkeyPEM); err != nil {
		return nil, err
	} else if !upok {
		return nil, fmt.Errorf("could not find ssl certificate, may be upload failed")
	} else {
		return upres, nil
	}
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	sslId, err := strconv.ParseInt(certIdOrName, 10, 64)
	if err != nil {
		return nil, err
	}

	switch sdkClient := c.sdkClient.(type) {
	case *onepanelsdk.Client:
		{
			// 获取证书详情
			websiteSSLGetResp, err := sdkClient.WebsiteSSLGet(sslId)
			c.logger.Debug("sdk request '1panel.WebsiteSSLGet'", slog.Int64("sslId", sslId), slog.Any("response", websiteSSLGetResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request '1panel.WebsiteSSLGet': %w", err)
			}

			// 更新证书
			websiteSSLUploadReq := &onepanelsdk.WebsiteSSLUploadRequest{
				SSLID:       sslId,
				Type:        "paste",
				Description: websiteSSLGetResp.Data.Description,
				Certificate: certPEM,
				PrivateKey:  privkeyPEM,
			}
			websiteSSLUploadResp, err := sdkClient.WebsiteSSLUpload(websiteSSLUploadReq)
			c.logger.Debug("sdk request '1panel.WebsiteSSLUpload'", slog.Any("request", websiteSSLUploadReq), slog.Any("response", websiteSSLUploadResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request '1panel.WebsiteSSLUpload': %w", err)
			}
		}

	case *onepanelsdk2.Client:
		{
			// 获取证书详情
			websiteSSLGetResp, err := sdkClient.WebsiteSSLGet(sslId)
			c.logger.Debug("sdk request '1panel.WebsiteSSLGet'", slog.Any("sslId", sslId), slog.Any("response", websiteSSLGetResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request '1panel.WebsiteSSLGet': %w", err)
			}

			// 更新证书
			websiteSSLUploadReq := &onepanelsdk2.WebsiteSSLUploadRequest{
				SSLID:       sslId,
				Type:        "paste",
				Description: websiteSSLGetResp.Data.Description,
				Certificate: certPEM,
				PrivateKey:  privkeyPEM,
			}
			websiteSSLUploadResp, err := sdkClient.WebsiteSSLUpload(websiteSSLUploadReq)
			c.logger.Debug("sdk request '1panel.WebsiteSSLUpload'", slog.Any("request", websiteSSLUploadReq), slog.Any("response", websiteSSLUploadResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request '1panel.WebsiteSSLUpload': %w", err)
			}
		}

	default:
		panic("unreachable")
	}

	return &certmgr.OperateResult{}, nil
}

func (c *Certmgr) tryGetResultIfCertExists(ctx context.Context, certPEM, privkeyPEM string) (*certmgr.UploadResult, bool, error) {
	switch sdkClient := c.sdkClient.(type) {
	case *onepanelsdk.Client:
		{
			searchWebsiteSSLPage := 1
			searchWebsiteSSLPageSize := 100
			for {
				select {
				case <-ctx.Done():
					return nil, false, ctx.Err()
				default:
				}

				websiteSSLSearchReq := &onepanelsdk.WebsiteSSLSearchRequest{
					Page:     int32(searchWebsiteSSLPage),
					PageSize: int32(searchWebsiteSSLPageSize),
				}
				websiteSSLSearchResp, err := sdkClient.WebsiteSSLSearch(websiteSSLSearchReq)
				c.logger.Debug("sdk request '1panel.WebsiteSSLSearch'", slog.Any("request", websiteSSLSearchReq), slog.Any("response", websiteSSLSearchResp))
				if err != nil {
					return nil, false, fmt.Errorf("failed to execute sdk request '1panel.WebsiteSSLSearch': %w", err)
				}

				if websiteSSLSearchResp.Data == nil {
					break
				}

				for _, sslItem := range websiteSSLSearchResp.Data.Items {
					oldCertPEM := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(sslItem.PEM, "\r", ""), "\n", ""))
					oldPrivkeyPEM := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(sslItem.PrivateKey, "\r", ""), "\n", ""))
					newCertPEM := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(certPEM, "\r", ""), "\n", ""))
					newPrivkeyPEM := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(privkeyPEM, "\r", ""), "\n", ""))
					if oldCertPEM == newCertPEM && oldPrivkeyPEM == newPrivkeyPEM {
						// 如果已存在相同证书，直接返回
						return &certmgr.UploadResult{
							CertId:   fmt.Sprintf("%d", sslItem.ID),
							CertName: sslItem.Description,
						}, true, nil
					}
				}

				if len(websiteSSLSearchResp.Data.Items) < int(websiteSSLSearchResp.Data.Total) {
					break
				}

				searchWebsiteSSLPage++
			}
		}

	case *onepanelsdk2.Client:
		{
			searchWebsiteSSLPage := 1
			searchWebsiteSSLPageSize := 100
			for {
				select {
				case <-ctx.Done():
					return nil, false, ctx.Err()
				default:
				}

				websiteSSLSearchReq := &onepanelsdk2.WebsiteSSLSearchRequest{
					Page:     int32(searchWebsiteSSLPage),
					PageSize: int32(searchWebsiteSSLPageSize),
				}
				websiteSSLSearchResp, err := sdkClient.WebsiteSSLSearch(websiteSSLSearchReq)
				c.logger.Debug("sdk request '1panel.WebsiteSSLSearch'", slog.Any("request", websiteSSLSearchReq), slog.Any("response", websiteSSLSearchResp))
				if err != nil {
					return nil, false, fmt.Errorf("failed to execute sdk request '1panel.WebsiteSSLSearch': %w", err)
				}

				if websiteSSLSearchResp.Data == nil {
					break
				}

				for _, sslItem := range websiteSSLSearchResp.Data.Items {
					oldCertPEM := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(sslItem.PEM, "\r", ""), "\n", ""))
					oldPrivkeyPEM := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(sslItem.PrivateKey, "\r", ""), "\n", ""))
					newCertPEM := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(certPEM, "\r", ""), "\n", ""))
					newPrivkeyPEM := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(privkeyPEM, "\r", ""), "\n", ""))
					if oldCertPEM == newCertPEM && oldPrivkeyPEM == newPrivkeyPEM {
						// 如果已存在相同证书，直接返回
						return &certmgr.UploadResult{
							CertId:   fmt.Sprintf("%d", sslItem.ID),
							CertName: sslItem.Description,
						}, true, nil
					}
				}

				if len(websiteSSLSearchResp.Data.Items) < int(websiteSSLSearchResp.Data.Total) {
					break
				}

				searchWebsiteSSLPage++
			}
		}

	default:
		panic("unreachable")
	}

	return nil, false, nil
}

const (
	sdkVersionV1 = "v1"
	sdkVersionV2 = "v2"
)

func createSDKClient(serverUrl, apiVersion, apiKey string, skipTlsVerify bool, nodeName string) (any, error) {
	if apiVersion == sdkVersionV1 {
		client, err := onepanelsdk.NewClient(serverUrl, apiKey)
		if err != nil {
			return nil, err
		}

		if skipTlsVerify {
			client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
		}

		return client, nil
	} else if apiVersion == sdkVersionV2 {
		var client *onepanelsdk2.Client
		var err error

		if nodeName == "" {
			client, err = onepanelsdk2.NewClient(serverUrl, apiKey)
		} else {
			client, err = onepanelsdk2.NewClientWithNode(serverUrl, apiKey, nodeName)
		}
		if err != nil {
			return nil, err
		}

		if skipTlsVerify {
			client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
		}

		return client, nil
	}

	return nil, errors.New("1panel: invalid api version")
}
