package onepanel

import (
	"cmp"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/1panel"
	onepanelsdk "github.com/certimate-go/certimate/pkg/sdk3rd/1panel"
	onepanelsdk2 "github.com/certimate-go/certimate/pkg/sdk3rd/1panel/v2"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 1Panel 服务地址。
	ServerUrl string `json:"serverUrl"`
	// 1Panel 版本。
	// 可取值 "v1"、"v2"。
	ApiVersion string `json:"apiVersion"`
	// 1Panel 接口密钥。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 子节点名称。
	// 选填。
	NodeName string `json:"nodeName,omitempty"`
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 域名匹配模式。
	// 零值时默认值 [WEBSITE_MATCH_PATTERN_SPECIFIED]。
	WebsiteMatchPattern string `json:"websiteMatchPattern,omitempty"`
	// 网站 ID。
	// 部署目标为 [DEPLOY_TARGET_WEBSITE]、且匹配模式非 [WEBSITE_MATCH_PATTERN_CERTSAN] 时必填。
	WebsiteId int64 `json:"websiteId,omitempty"`
	// 证书 ID。
	// 部署目标为 [DEPLOY_TARGET_CERTIFICATE] 时必填。
	CertificateId int64 `json:"certificateId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  any
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiVersion, config.ApiKey, config.AllowInsecureConnections, config.NodeName)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		ServerUrl:                config.ServerUrl,
		ApiVersion:               config.ApiVersion,
		ApiKey:                   config.ApiKey,
		AllowInsecureConnections: config.AllowInsecureConnections,
		NodeName:                 config.NodeName,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create certmgr: %w", err)
	}

	return &Deployer{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
		sdkCertmgr: pcertmgr,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}

	d.sdkCertmgr.SetLogger(logger)
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	// 根据部署目标决定业务流程
	switch d.config.DeployTarget {
	case DEPLOY_TARGET_WEBSITE:
		if err := d.deployToWebsite(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case DEPLOY_TARGET_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToWebsite(ctx context.Context, certPEM, privkeyPEM string) error {
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取待部署的网站列表
	var websiteIds []int64
	switch d.config.WebsiteMatchPattern {
	case "", WEBSITE_MATCH_PATTERN_SPECIFIED:
		{
			if d.config.WebsiteId == 0 {
				return fmt.Errorf("config `websiteId` is required")
			}

			websiteIds = []int64{d.config.WebsiteId}
		}

	case WEBSITE_MATCH_PATTERN_CERTSAN:
		{
			websiteIdCandidates, err := d.getMatchedWebsiteIdsByCertificate(ctx, certPEM)
			if err != nil {
				return err
			}

			websiteIds = websiteIdCandidates
		}

	default:
		return fmt.Errorf("unsupported website match pattern: '%s'", d.config.WebsiteMatchPattern)
	}

	// 遍历更新网站证书
	if len(websiteIds) == 0 {
		d.logger.Info("no websites to deploy")
	} else {
		d.logger.Info("found websites to deploy", slog.Any("websiteIds", websiteIds))
		var errs []error

		websiteSSLId, _ := strconv.ParseInt(upres.CertId, 10, 64)
		for i, websiteId := range websiteIds {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateWebsiteCertificate(ctx, websiteId, websiteSSLId); err != nil {
					errs = append(errs, err)
				} else if i < len(websiteIds)-1 {
					xwait.DelayWithContext(ctx, 5*time.Second)
				}
			}
		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}
	}

	return nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == 0 {
		return fmt.Errorf("config `certificateId` is required")
	}

	// 替换证书
	rplres, err := d.sdkCertmgr.Replace(ctx, strconv.FormatInt(d.config.CertificateId, 10), certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to replace certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate replaced", slog.Any("result", rplres))
	}

	return nil
}

func (d *Deployer) getMatchedWebsiteIdsByCertificate(ctx context.Context, certPEM string) ([]int64, error) {
	var websiteIds []int64

	switch sdkClient := d.sdkClient.(type) {
	case *onepanelsdk.Client:
		{
			websiteSearchPage := 1
			websiteSearchPageSize := 100
			for {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
				}
				websiteSearchReq := &onepanelsdk.WebsiteSearchRequest{
					Order:    "ascending",
					OrderBy:  "primary_domain",
					Page:     int32(websiteSearchPage),
					PageSize: int32(websiteSearchPageSize),
				}
				websiteSearchResp, err := sdkClient.WebsiteSearchWithContext(ctx, websiteSearchReq)
				d.logger.Debug("sdk request 'WebsiteSearch'", slog.Any("request", websiteSearchReq), slog.Any("response", websiteSearchResp))
				if err != nil {
					return nil, fmt.Errorf("failed to execute sdk request 'WebsiteSearch': %w", err)
				}

				if websiteSearchResp.Data == nil {
					break
				}

				for _, websiteItem := range websiteSearchResp.Data.Items {
					if !xcerthostname.IsMatchByCertificatePEM(certPEM, websiteItem.PrimaryDomain) {
						continue
					}

					websiteGetResp, err := sdkClient.WebsiteGetWithContext(ctx, websiteItem.ID)
					d.logger.Debug("sdk request 'WebsiteGet'", slog.Int64("params.websiteId", websiteItem.ID), slog.Any("response", websiteGetResp))
					if err != nil {
						return nil, fmt.Errorf("failed to execute sdk request 'WebsiteGet': %w", err)
					}

					for _, domainInfo := range websiteGetResp.Data.Domains {
						if domainInfo.SSL || xcerthostname.IsMatchByCertificatePEM(certPEM, domainInfo.Domain) {
							websiteIds = append(websiteIds, websiteItem.ID)
							break
						}
					}
				}

				if len(websiteSearchResp.Data.Items) < websiteSearchPageSize {
					break
				}

				websiteSearchPage++
			}
		}

	case *onepanelsdk2.Client:
		{
			websiteSearchPage := 1
			websiteSearchPageSize := 100
			for {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
				}

				websiteSearchReq := &onepanelsdk2.WebsiteSearchRequest{
					Order:    "ascending",
					OrderBy:  "primary_domain",
					Page:     int32(websiteSearchPage),
					PageSize: int32(websiteSearchPageSize),
				}
				websiteSearchResp, err := sdkClient.WebsiteSearchWithContext(ctx, websiteSearchReq)
				d.logger.Debug("sdk request 'WebsiteSearch'", slog.Any("request", websiteSearchReq), slog.Any("response", websiteSearchResp))
				if err != nil {
					return nil, fmt.Errorf("failed to execute sdk request 'WebsiteSearch': %w", err)
				}

				if websiteSearchResp.Data == nil {
					break
				}

				for _, websiteItem := range websiteSearchResp.Data.Items {
					if !xcerthostname.IsMatchByCertificatePEM(certPEM, websiteItem.PrimaryDomain) {
						continue
					}

					websiteGetResp, err := sdkClient.WebsiteGetWithContext(ctx, websiteItem.ID)
					d.logger.Debug("sdk request 'WebsiteGet'", slog.Int64("params.websiteId", websiteItem.ID), slog.Any("response", websiteGetResp))
					if err != nil {
						return nil, fmt.Errorf("failed to execute sdk request 'WebsiteGet': %w", err)
					}

					for _, domainInfo := range websiteGetResp.Data.Domains {
						if domainInfo.SSL || xcerthostname.IsMatchByCertificatePEM(certPEM, domainInfo.Domain) {
							websiteIds = append(websiteIds, websiteItem.ID)
							break
						}
					}
				}

				if len(websiteSearchResp.Data.Items) < websiteSearchPageSize {
					break
				}

				websiteSearchPage++
			}
		}

	default:
		panic("unreachable")
	}

	if len(websiteIds) == 0 {
		return nil, fmt.Errorf("could not find any websites matched by certificate")
	}

	return websiteIds, nil
}

func (d *Deployer) updateWebsiteCertificate(ctx context.Context, websiteId int64, websiteSSLId int64) error {
	switch sdkClient := d.sdkClient.(type) {
	case *onepanelsdk.Client:
		{
			// 获取网站 HTTPS 配置
			websiteHttpsGetResp, err := sdkClient.WebsiteHttpsGetWithContext(ctx, websiteId)
			d.logger.Debug("sdk request 'WebsiteHttpsGet'", slog.Int64("params.websiteId", websiteId), slog.Any("response", websiteHttpsGetResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'WebsiteHttpsGet': %w", err)
			} else {
				if websiteHttpsGetResp.Data.Enable && websiteHttpsGetResp.Data.WebsiteSSLID == websiteSSLId {
					return nil
				}
			}

			// 修改网站 HTTPS 配置
			websiteHttpsPostReq := &onepanelsdk.WebsiteHttpsPostRequest{
				WebsiteID:    websiteId,
				Type:         "existed",
				WebsiteSSLID: websiteSSLId,
				Enable:       true,
				HttpConfig:   cmp.Or(websiteHttpsGetResp.Data.HttpConfig, "HTTPToHTTPS"),
				SSLProtocol:  websiteHttpsGetResp.Data.SSLProtocol,
				Algorithm:    websiteHttpsGetResp.Data.Algorithm,
				Hsts:         websiteHttpsGetResp.Data.Hsts,
			}
			websiteHttpsPostResp, err := sdkClient.WebsiteHttpsPostWithContext(ctx, websiteId, websiteHttpsPostReq)
			d.logger.Debug("sdk request 'WebsiteHttpsPost'", slog.Int64("params.websiteId", websiteId), slog.Any("request", websiteHttpsPostReq), slog.Any("response", websiteHttpsPostResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'WebsiteHttpsPost': %w", err)
			}
		}

	case *onepanelsdk2.Client:
		{
			// 获取网站 HTTPS 配置
			websiteHttpsGetResp, err := sdkClient.WebsiteHttpsGetWithContext(ctx, websiteId)
			d.logger.Debug("sdk request 'WebsiteHttpsGet'", slog.Int64("params.websiteId", websiteId), slog.Any("response", websiteHttpsGetResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'WebsiteHttpsGet': %w", err)
			} else {
				if websiteHttpsGetResp.Data.Enable && websiteHttpsGetResp.Data.WebsiteSSLID == websiteSSLId {
					return nil
				}
			}

			// 修改网站 HTTPS 配置
			websiteHttpsPostReq := &onepanelsdk2.WebsiteHttpsPostRequest{
				WebsiteID:    websiteId,
				Type:         "existed",
				WebsiteSSLID: websiteSSLId,
				Enable:       true,
				HttpConfig:   cmp.Or(websiteHttpsGetResp.Data.HttpConfig, "HTTPToHTTPS"),
				SSLProtocol:  websiteHttpsGetResp.Data.SSLProtocol,
				Algorithm:    websiteHttpsGetResp.Data.Algorithm,
				Hsts:         websiteHttpsGetResp.Data.Hsts,
				Http3:        websiteHttpsGetResp.Data.Http3,
			}
			websiteHttpsPostResp, err := sdkClient.WebsiteHttpsPostWithContext(ctx, websiteId, websiteHttpsPostReq)
			d.logger.Debug("sdk request 'WebsiteHttpsPost'", slog.Int64("params.websiteId", websiteId), slog.Any("request", websiteHttpsPostReq), slog.Any("response", websiteHttpsPostResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'WebsiteHttpsPost': %w", err)
			}
		}

	default:
		panic("unreachable")
	}

	return nil
}

const (
	sdkVersionV1 = "v1"
	sdkVersionV2 = "v2"
)

func createSDKClient(serverUrl, apiVersion, apiKey string, skipTlsVerify bool, nodeName string) (any, error) {
	if apiVersion == sdkVersionV1 {
		client, err := onepanelsdk.NewClient(serverUrl,
			onepanelsdk.WithApiKey(apiKey),
		)
		if err != nil {
			return nil, err
		}

		if skipTlsVerify {
			client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
		}

		return client, nil
	} else if apiVersion == sdkVersionV2 {
		client, err := onepanelsdk2.NewClient(serverUrl,
			onepanelsdk2.WithApiKey(apiKey),
			onepanelsdk2.WithNode(nodeName),
		)
		if err != nil {
			return nil, err
		}

		if skipTlsVerify {
			client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
		}

		return client, nil
	}

	return nil, fmt.Errorf("1panel: invalid api version")
}
