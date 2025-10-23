package baotapanelgosite

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	btsdk "github.com/certimate-go/certimate/pkg/sdk3rd/btpanelgo"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type SSLDeployerProviderConfig struct {
	// 宝塔面板服务地址。
	ServerUrl string `json:"serverUrl"`
	// 宝塔面板接口密钥。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 网站名称。
	SiteName string `json:"siteName,omitempty"`
}

type SSLDeployerProvider struct {
	config    *SSLDeployerProviderConfig
	logger    *slog.Logger
	sdkClient *btsdk.Client
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

	// 设置站点 SSL 证书
	panelGetConfigReq := &btsdk.PanelGetConfigRequest{}
	panelGetConfigResp, err := d.sdkClient.PanelGetConfig(panelGetConfigReq)
	d.logger.Debug("sdk request 'bt.PanelGetConfig'", slog.Any("request", panelGetConfigReq), slog.Any("response", panelGetConfigResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'bt.PanelGetConfig': %w", err)
	}

	// 遍历查询网站列表，获取网站 ID
	var siteId int32
	datalistGetDataListPage := int32(1)
	datalistGetDataListLimit := int32(10)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		datalistGetDataListReq := &btsdk.DatalistGetDataListRequest{
			Table:        lo.ToPtr("sites"),
			SearchString: lo.ToPtr(d.config.SiteName),
			Page:         lo.ToPtr(datalistGetDataListPage),
			Limit:        lo.ToPtr(datalistGetDataListLimit),
		}
		datalistGetDataListResp, err := d.sdkClient.DatalistGetDataList(datalistGetDataListReq)
		d.logger.Debug("sdk request 'bt.DatalistGetDataList'", slog.Any("request", datalistGetDataListReq), slog.Any("response", datalistGetDataListResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'bt.DatalistGetDataList': %w", err)
		}

		for _, siteInfo := range datalistGetDataListResp.Data {
			if strings.EqualFold(siteInfo.Name, d.config.SiteName) {
				siteId = siteInfo.Id
				break
			}
		}

		if siteId != 0 {
			break
		}

		if len(datalistGetDataListResp.Data) < int(datalistGetDataListLimit) {
			break
		} else {
			datalistGetDataListPage++
		}
	}
	if siteId == 0 {
		return nil, errors.New("website not found")
	}

	if panelGetConfigResp.Site != nil && strings.EqualFold(panelGetConfigResp.Site.WebServer, "iis") {
		// 转换证书格式
		certPFXPassword := "certimate"
		certPFX, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, certPFXPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to transform certificate from PEM to PFX: %w", err)
		}

		// 上传证书
		certPFXHash := sha256.Sum256([]byte(certPFX))
		certPFXHashHex := hex.EncodeToString(certPFXHash[:])
		certPFXPath := panelGetConfigResp.Paths.Soft + "/temp/ssl/certimate"
		certPFXFileName := fmt.Sprintf("%s.pfx", certPFXHashHex)
		filesUploadReq := &btsdk.FilesUploadRequest{
			Path:  lo.ToPtr(certPFXPath),
			Name:  lo.ToPtr(certPFXFileName),
			Start: lo.ToPtr(int32(0)),
			Size:  lo.ToPtr(int32(len(certPFX))),
			Blob:  certPFX,
			Force: lo.ToPtr(true),
		}
		filesUploadResp, err := d.sdkClient.FilesUpload(filesUploadReq)
		d.logger.Debug("sdk request 'bt.FilesUpload'", slog.Any("request", filesUploadReq), slog.Any("response", filesUploadResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'bt.FilesUpload': %w", err)
		}

		// 服务器为 IIS，设置网站 SSL
		siteSetSitePFXSSLReq := &btsdk.SiteSetSitePFXSSLRequest{
			SiteId:   lo.ToPtr(siteId),
			PFX:      lo.ToPtr(fmt.Sprintf("%s/%s", certPFXPath, certPFXFileName)),
			Password: lo.ToPtr(certPFXPassword),
		}
		siteSetSitePFXSSLResp, err := d.sdkClient.SiteSetSitePFXSSL(siteSetSitePFXSSLReq)
		d.logger.Debug("sdk request 'bt.SiteSetSitePFXSSL'", slog.Any("request", siteSetSitePFXSSLReq), slog.Any("response", siteSetSitePFXSSLResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'bt.SiteSetSitePFXSSL': %w", err)
		}
	} else {
		// 服务器非 IIS，设置网站 SSL
		siteSetSiteSSLReq := &btsdk.SiteSetSiteSSLRequest{
			SiteId: lo.ToPtr(siteId),
			Status: lo.ToPtr(true),
			Key:    lo.ToPtr(privkeyPEM),
			Cert:   lo.ToPtr(certPEM),
		}
		siteSetSiteSSLResp, err := d.sdkClient.SiteSetSiteSSL(siteSetSiteSSLReq)
		d.logger.Debug("sdk request 'bt.SiteSetSiteSSL'", slog.Any("request", siteSetSiteSSLReq), slog.Any("response", siteSetSiteSSLResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'bt.SiteSetSiteSSL': %w", err)
		}
	}

	return &core.SSLDeployResult{}, nil
}

func createSDKClient(serverUrl, apiKey string, skipTlsVerify bool) (*btsdk.Client, error) {
	client, err := btsdk.NewClient(serverUrl, apiKey)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
