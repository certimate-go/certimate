package baotapanelwaf

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core"
	btwafsdk "github.com/certimate-go/certimate/pkg/sdk3rd/btwaf"
	xtypes "github.com/certimate-go/certimate/pkg/utils/types"
)

type SSLDeployerProviderConfig struct {
	// 堡塔云 WAF 服务地址。
	ServerUrl string `json:"serverUrl"`
	// 堡塔云 WAF 接口密钥。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 网站名称。
	SiteName string `json:"siteName"`
	// 网站 SSL 端口。
	// 零值时默认值 443。
	SitePort int32 `json:"sitePort,omitempty"`
}

type SSLDeployerProvider struct {
	config    *SSLDeployerProviderConfig
	logger    *slog.Logger
	sdkClient *btwafsdk.Client
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiKey, config.AllowInsecureConnections)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	return &SSLDeployerProvider{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (d *SSLDeployerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
	if d.config.SiteName == "" {
		return nil, errors.New("config `siteName` is required")
	}
	if d.config.SitePort == 0 {
		d.config.SitePort = 443
	}

	// 遍历获取网站列表，获取网站 ID
	// REF: https://support.huaweicloud.com/api-waf/ListHost.html
	siteId := ""
	getSitListPage := int32(1)
	getSitListPageSize := int32(100)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		getSiteListReq := &btwafsdk.GetSiteListRequest{
			SiteName: xtypes.ToPtr(d.config.SiteName),
			Page:     xtypes.ToPtr(getSitListPage),
			PageSize: xtypes.ToPtr(getSitListPageSize),
		}
		getSiteListResp, err := d.sdkClient.GetSiteList(getSiteListReq)
		d.logger.Debug("sdk request 'bt.GetSiteList'", slog.Any("request", getSiteListReq), slog.Any("response", getSiteListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'bt.GetSiteList': %w", err)
		}

		if getSiteListResp.Result != nil && getSiteListResp.Result.List != nil {
			for _, siteItem := range getSiteListResp.Result.List {
				if siteItem.SiteName == d.config.SiteName {
					siteId = siteItem.SiteId
					break
				}
			}
		}

		if getSiteListResp.Result == nil || len(getSiteListResp.Result.List) < int(getSitListPageSize) {
			break
		} else {
			getSitListPage++
		}
	}
	if siteId == "" {
		return nil, errors.New("site not found")
	}

	// 修改站点配置
	modifySiteReq := &btwafsdk.ModifySiteRequest{
		SiteId: xtypes.ToPtr(siteId),
		Type:   xtypes.ToPtr("openCert"),
		Server: &btwafsdk.SiteServerInfoMod{
			ListenSSLPorts: xtypes.ToPtr([]string{fmt.Sprintf("%d", d.config.SitePort)}),
			SSL: &btwafsdk.SiteServerSSLInfo{
				IsSSL:      xtypes.ToPtr(int32(1)),
				FullChain:  xtypes.ToPtr(certPEM),
				PrivateKey: xtypes.ToPtr(privkeyPEM),
			},
		},
	}
	modifySiteResp, err := d.sdkClient.ModifySite(modifySiteReq)
	d.logger.Debug("sdk request 'bt.ModifySite'", slog.Any("request", modifySiteReq), slog.Any("response", modifySiteResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'bt.ModifySite': %w", err)
	}

	return &core.SSLDeployResult{}, nil
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
