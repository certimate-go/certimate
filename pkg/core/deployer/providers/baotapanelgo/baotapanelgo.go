package baotapanelgo

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	btsdk "github.com/certimate-go/certimate/pkg/sdk3rd/btpanelgo"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type DeployerConfig struct {
	// 宝塔面板服务地址。
	ServerUrl string `json:"serverUrl"`
	// 宝塔面板接口密钥。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 网站类型。
	SiteType string `json:"siteType"`
	// 网站名称。
	SiteNames []string `json:"siteNames,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *btsdk.Client
}

var _ deployer.Provider = (*Deployer)(nil)

var (
	btProjectTypes      = []string{"php", "java", "asp", "go", "python", "nodejs", "proxy", "general"}
	btProjectTypesInIIS = []string{"php", "asp", "aspx"}
)

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
			if err := d.updatePHPSiteCertificate(ctx, d.config.SiteType, siteName, certPEM, privkeyPEM); err != nil {
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

func (d *Deployer) findSiteByName(ctx context.Context, siteType, siteName string) (*btsdk.SiteData, error) {
	if siteType == "" || lo.Contains(btProjectTypesInIIS, siteType) {
		// 查询网站列表
		datalistGetDataListPage := 1
		datalistGetDataListLimit := 10
		for {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			datalistGetDataListReq := &btsdk.DatalistGetDataListRequest{
				Table:        lo.ToPtr("sites"),
				SearchString: lo.ToPtr(siteName),
				Page:         lo.ToPtr(int32(datalistGetDataListPage)),
				Limit:        lo.ToPtr(int32(datalistGetDataListLimit)),
			}
			datalistGetDataListResp, err := d.sdkClient.DatalistGetDataList(datalistGetDataListReq)
			d.logger.Debug("sdk request 'bt.DatalistGetDataList'", slog.Any("request", datalistGetDataListReq), slog.Any("response", datalistGetDataListResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'bt.DatalistGetDataList': %w", err)
			}

			for _, siteItem := range datalistGetDataListResp.Data {
				if strings.EqualFold(siteItem.Name, siteName) {
					return siteItem, nil
				}
			}

			if len(datalistGetDataListResp.Data) < datalistGetDataListLimit {
				break
			}

			datalistGetDataListPage++
		}
	} else {
		// 查询网站列表
		siteGetProjectListPage := 1
		siteGetProjectListLimit := 10
		for {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			siteGetProjectListReq := &btsdk.SiteGetProjectListRequest{
				SearchType:   lo.ToPtr(siteType),
				SearchString: lo.ToPtr(siteName),
				Page:         lo.ToPtr(int32(siteGetProjectListPage)),
				Limit:        lo.ToPtr(int32(siteGetProjectListLimit)),
			}
			siteGetProjectListResp, err := d.sdkClient.SiteGetProjectList(siteGetProjectListReq)
			d.logger.Debug("sdk request 'bt.SiteGetProjectList'", slog.Any("request", siteGetProjectListReq), slog.Any("response", siteGetProjectListResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'bt.SiteGetProjectList': %w", err)
			}

			for _, siteItem := range siteGetProjectListResp.Data {
				if strings.EqualFold(siteItem.Name, siteName) {
					return siteItem, nil
				}
			}

			if len(siteGetProjectListResp.Data) < siteGetProjectListLimit {
				break
			}

			siteGetProjectListPage++
		}
	}

	return nil, fmt.Errorf("could not find site '%s'", siteName)
}

func (d *Deployer) updatePHPSiteCertificate(ctx context.Context, siteType, siteName string, certPEM, privkeyPEM string) error {
	// 获取面板配置
	panelGetConfigReq := &btsdk.PanelGetConfigRequest{}
	panelGetConfigResp, err := d.sdkClient.PanelGetConfig(panelGetConfigReq)
	d.logger.Debug("sdk request 'bt.PanelGetConfig'", slog.Any("request", panelGetConfigReq), slog.Any("response", panelGetConfigResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'bt.PanelGetConfig': %w", err)
	}

	// 获取网站
	siteData, err := d.findSiteByName(ctx, siteType, siteName)
	if err != nil {
		return err
	}

	// 根据服务器类型部署证书
	pfxRequried := panelGetConfigResp.Site != nil && strings.EqualFold(panelGetConfigResp.Site.WebServer, "iis")
	if pfxRequried {
		// 转换证书格式
		certPFXPassword := "certimate"
		certPFX, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, certPFXPassword)
		if err != nil {
			return fmt.Errorf("failed to transform certificate from PEM to PFX: %w", err)
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
			return fmt.Errorf("failed to execute sdk request 'bt.FilesUpload': %w", err)
		}

		// 服务器为 IIS，设置网站 SSL
		siteSetSitePFXSSLReq := &btsdk.SiteSetSitePFXSSLRequest{
			SiteId:   lo.ToPtr(siteData.Id),
			PFX:      lo.ToPtr(fmt.Sprintf("%s/%s", certPFXPath, certPFXFileName)),
			Password: lo.ToPtr(certPFXPassword),
		}
		siteSetSitePFXSSLResp, err := d.sdkClient.SiteSetSitePFXSSL(siteSetSitePFXSSLReq)
		d.logger.Debug("sdk request 'bt.SiteSetSitePFXSSL'", slog.Any("request", siteSetSitePFXSSLReq), slog.Any("response", siteSetSitePFXSSLResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'bt.SiteSetSitePFXSSL': %w", err)
		}
	} else {
		// 服务器非 IIS，设置网站 SSL
		siteSetSiteSSLReq := &btsdk.SiteSetSiteSSLRequest{
			SiteId: lo.ToPtr(siteData.Id),
			Status: lo.ToPtr(true),
			Key:    lo.ToPtr(privkeyPEM),
			Cert:   lo.ToPtr(certPEM),
		}
		siteSetSiteSSLResp, err := d.sdkClient.SiteSetSiteSSL(siteSetSiteSSLReq)
		d.logger.Debug("sdk request 'bt.SiteSetSiteSSL'", slog.Any("request", siteSetSiteSSLReq), slog.Any("response", siteSetSiteSSLResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'bt.SiteSetSiteSSL': %w", err)
		}
	}

	return nil
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
