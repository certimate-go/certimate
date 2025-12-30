package baotawaf

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	btwafsdk "github.com/certimate-go/certimate/pkg/sdk3rd/btwaf"
)

type DeployerConfig struct {
	// 堡塔云 WAF 服务地址。
	ServerUrl string `json:"serverUrl"`
	// 堡塔云 WAF 接口密钥。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 网站名称。
	SiteNames []string `json:"siteNames"`
	// 网站 SSL 端口。
	// 零值时默认值 443。
	SitePort int32 `json:"sitePort,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *btwafsdk.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiKey, config.AllowInsecureConnections)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Deployer{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
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
	if len(d.config.SiteNames) == 0 {
		return nil, errors.New("config `siteNames` is required")
	}

	// 遍历更新站点证书
	var errs []error
	for i, siteName := range d.config.SiteNames {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if err := d.updateSiteCertificate(ctx, siteName, d.config.SitePort, certPEM, privkeyPEM); err != nil {
				errs = append(errs, err)
			}
			if i < len(d.config.SiteNames)-1 {
				time.Sleep(time.Second * 5)
			}
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) findSiteByName(ctx context.Context, siteName string) (*btwafsdk.SiteRecord, error) {
	// 查询网站列表
	getSiteListPage := 1
	getSiteListPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		getSiteListReq := &btwafsdk.GetSiteListRequest{
			SiteName: lo.ToPtr(siteName),
			Page:     lo.ToPtr(int32(getSiteListPage)),
			PageSize: lo.ToPtr(int32(getSiteListPageSize)),
		}
		getSiteListResp, err := d.sdkClient.GetSiteList(getSiteListReq)
		d.logger.Debug("sdk request 'bt.GetSiteList'", slog.Any("request", getSiteListReq), slog.Any("response", getSiteListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'bt.GetSiteList': %w", err)
		}

		if getSiteListResp.Result == nil {
			break
		}

		for _, siteItem := range getSiteListResp.Result.List {
			if siteItem.SiteName == siteName {
				return siteItem, nil
			}
		}

		if len(getSiteListResp.Result.List) < getSiteListPageSize {
			break
		}

		getSiteListPage++
	}

	return nil, fmt.Errorf("could not find site '%s'", siteName)
}

func (d *Deployer) updateSiteCertificate(ctx context.Context, siteName string, sitePort int32, certPEM, privkeyPEM string) error {
	if sitePort == 0 {
		sitePort = 443
	}

	// 获取网站配置
	siteData, err := d.findSiteByName(ctx, siteName)
	if err != nil {
		return err
	}

	// 修改网站配置
	modifySiteReq := &btwafsdk.ModifySiteRequest{
		SiteId: lo.ToPtr(siteData.SiteId),
		Type:   lo.ToPtr("openCert"),
		Server: &btwafsdk.SiteServerInfoMod{
			ListenSSLPorts: lo.ToPtr([]string{fmt.Sprintf("%d", d.config.SitePort)}),
			SSL: &btwafsdk.SiteServerSSLInfo{
				IsSSL:      lo.ToPtr(int32(1)),
				FullChain:  lo.ToPtr(certPEM),
				PrivateKey: lo.ToPtr(privkeyPEM),
			},
		},
	}
	modifySiteResp, err := d.sdkClient.ModifySite(modifySiteReq)
	d.logger.Debug("sdk request 'bt.ModifySite'", slog.Any("request", modifySiteReq), slog.Any("response", modifySiteResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'bt.ModifySite': %w", err)
	}

	return nil
}

func createSDKClient(serverUrl, apiKey string, skipTlsVerify bool) (*btwafsdk.Client, error) {
	client, err := btwafsdk.NewClient(serverUrl, apiKey)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
