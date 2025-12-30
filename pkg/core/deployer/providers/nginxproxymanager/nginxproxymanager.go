package nginxproxymanager

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/nginxproxymanager"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	npmsdk "github.com/certimate-go/certimate/pkg/sdk3rd/nginxproxymanager"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type DeployerConfig struct {
	// NPM 服务地址。
	ServerUrl string `json:"serverUrl"`
	// NPM API 认证方式。
	// 可取值 "password"、"token"。
	// 零值时默认值 [AUTH_METHOD_PASSWORD]。
	AuthMethod string `json:"authMethod,omitempty"`
	// NPM 用户名。
	Username string `json:"username,omitempty"`
	// NPM 密码。
	Password string `json:"password,omitempty"`
	// NPM API Token。
	ApiToken string `json:"apiToken,omitempty"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 域名匹配模式。
	// 零值时默认值 [HOST_MATCH_PATTERN_SPECIFIED]。
	HostMatchPattern string `json:"hostMatchPattern,omitempty"`
	// 主机类型。
	// 部署资源类型为 [RESOURCE_TYPE_HOST] 时必填。
	HostType string `json:"hostType,omitempty"`
	// 主机 ID。
	// 部署资源类型为 [RESOURCE_TYPE_HOST]、且匹配模式非 [HOST_MATCH_PATTERN_CERTSAN] 时必填。
	HostId int64 `json:"hostId,omitempty"`
	// 证书 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时必填。
	CertificateId int64 `json:"certificateId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *npmsdk.Client
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.AuthMethod, config.Username, config.Password, config.ApiToken, config.AllowInsecureConnections)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		ServerUrl:                config.ServerUrl,
		AuthMethod:               config.AuthMethod,
		Username:                 config.Username,
		Password:                 config.Password,
		ApiToken:                 config.ApiToken,
		AllowInsecureConnections: config.AllowInsecureConnections,
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_HOST:
		if err := d.deployToHost(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToHost(ctx context.Context, certPEM, privkeyPEM string) error {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return err
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取待部署的主机列表
	var hostIds []int64
	switch d.config.HostMatchPattern {
	case "", HOST_MATCH_PATTERN_SPECIFIED:
		{
			if d.config.HostId == 0 {
				return errors.New("config `hostId` is required")
			}

			hostIds = []int64{d.config.HostId}
		}

	case HOST_MATCH_PATTERN_CERTSAN:
		{
			hostCandidates, err := d.getAllHosts(ctx, d.config.HostType)
			if err != nil {
				return err
			}

			hostIds = lo.Map(
				lo.Filter(hostCandidates, func(hostItem *npmsdk.HostRecord, _ int) bool {
					return len(hostItem.DomainNames) > 0 &&
						lo.EveryBy(hostItem.DomainNames, func(domain string) bool {
							return certX509.VerifyHostname(domain) == nil
						})
				}),
				func(hostItem *npmsdk.HostRecord, _ int) int64 {
					return hostItem.Id
				},
			)
			if len(hostIds) == 0 {
				return errors.New("could not find any hosts matched by certificate")
			}

			// 跳过已部署过的主机
			hostIds = lo.Filter(hostIds, func(hostId int64, _ int) bool {
				hostInfo, _ := lo.Find(hostCandidates, func(hostItem *npmsdk.HostRecord) bool {
					return hostId == hostItem.Id
				})
				if hostInfo != nil {
					return strconv.FormatInt(hostInfo.CertificateId, 10) != upres.CertId
				}

				return true
			})
		}

	default:
		return fmt.Errorf("unsupported host match pattern: '%s'", d.config.HostMatchPattern)
	}

	// 遍历更新主机证书
	if len(hostIds) == 0 {
		d.logger.Info("no hosts to deploy")
	} else {
		d.logger.Info("found hosts to deploy", slog.Any("hostIds", hostIds))
		var errs []error

		certId, _ := strconv.ParseInt(upres.CertId, 10, 64)
		for i, hostId := range hostIds {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateHostCertificate(ctx, d.config.HostType, hostId, certId); err != nil {
					errs = append(errs, err)
				}
				if i < len(hostIds)-1 {
					time.Sleep(time.Second * 5)
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
		return errors.New("config `certificateId` is required")
	}

	// 替换证书
	opres, err := d.sdkCertmgr.Replace(ctx, strconv.FormatInt(d.config.CertificateId, 10), certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to replace certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate replaced", slog.Any("result", opres))
	}

	// 获取默认站点
	settingsGetDefaultSiteReq := &npmsdk.SettingsGetDefaultSiteRequest{}
	settingsGetDefaultSiteResp, err := d.sdkClient.SettingsGetDefaultSite(settingsGetDefaultSiteReq)
	d.logger.Debug("sdk request 'settings.GetDefaultSite'", slog.Any("request", settingsGetDefaultSiteReq), slog.Any("response", settingsGetDefaultSiteResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'settings.GetDefaultSite': %w", err)
	}

	// 更新默认站点，以触发 nginx 重启
	settingsSetDefaultSiteReq := &npmsdk.SettingsSetDefaultSiteRequest{
		Value: settingsGetDefaultSiteResp.Value,
	}
	settingsSetDefaultSiteResp, err := d.sdkClient.SettingsSetDefaultSite(settingsSetDefaultSiteReq)
	d.logger.Debug("sdk request 'settings.SetDefaultSite'", slog.Any("request", settingsSetDefaultSiteReq), slog.Any("response", settingsSetDefaultSiteResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'settings.SetDefaultSite': %w", err)
	}

	return nil
}

func (d *Deployer) getAllHosts(ctx context.Context, cloudHostType string) ([]*npmsdk.HostRecord, error) {
	var hosts []*npmsdk.HostRecord
	switch cloudHostType {
	case HOST_TYPE_PROXY:
		{
			nginxListProxyHostsReq := &npmsdk.NginxListProxyHostsRequest{}
			nginxListProxyHostsResp, err := d.sdkClient.NginxListProxyHosts(nginxListProxyHostsReq)
			d.logger.Debug("sdk request 'nginx.ListProxyHosts'", slog.Any("request", nginxListProxyHostsReq), slog.Any("response", nginxListProxyHostsResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'nginx.ListProxyHosts': %w", err)
			}

			hosts = make([]*npmsdk.HostRecord, 0, len(*nginxListProxyHostsResp))
			for _, hostItem := range *nginxListProxyHostsResp {
				hosts = append(hosts, &hostItem.HostRecord)
			}
		}

	case HOST_TYPE_REDIRECTION:
		{
			nginxListRedirectionHostsReq := &npmsdk.NginxListRedirectionHostsRequest{}
			nginxListRedirectionHostsResp, err := d.sdkClient.NginxListRedirectionHosts(nginxListRedirectionHostsReq)
			d.logger.Debug("sdk request 'nginx.ListRedirectionHosts'", slog.Any("request", nginxListRedirectionHostsReq), slog.Any("response", nginxListRedirectionHostsResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'nginx.ListRedirectionHosts': %w", err)
			}

			hosts = make([]*npmsdk.HostRecord, 0, len(*nginxListRedirectionHostsResp))
			for _, hostItem := range *nginxListRedirectionHostsResp {
				hosts = append(hosts, &hostItem.HostRecord)
			}
		}

	case HOST_TYPE_STREAM:
		{
			nginxListStreamsReq := &npmsdk.NginxListStreamsRequest{}
			nginxListStreamsResp, err := d.sdkClient.NginxListStreams(nginxListStreamsReq)
			d.logger.Debug("sdk request 'nginx.ListStreams'", slog.Any("request", nginxListStreamsReq), slog.Any("response", nginxListStreamsResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'nginx.ListStreams': %w", err)
			}

			hosts = make([]*npmsdk.HostRecord, 0, len(*nginxListStreamsResp))
			for _, hostItem := range *nginxListStreamsResp {
				hosts = append(hosts, &hostItem.HostRecord)
			}
		}

	case HOST_TYPE_DEAD:
		{
			nginxListDeadHostsReq := &npmsdk.NginxListDeadHostsRequest{}
			nginxListDeadHostsResp, err := d.sdkClient.NginxListDeadHosts(nginxListDeadHostsReq)
			d.logger.Debug("sdk request 'nginx.ListDeadHosts'", slog.Any("request", nginxListDeadHostsReq), slog.Any("response", nginxListDeadHostsResp))
			if err != nil {
				return nil, fmt.Errorf("failed to execute sdk request 'nginx.ListDeadHosts': %w", err)
			}

			hosts = make([]*npmsdk.HostRecord, 0, len(*nginxListDeadHostsResp))
			for _, hostItem := range *nginxListDeadHostsResp {
				hosts = append(hosts, &hostItem.HostRecord)
			}
		}

	default:
		return hosts, fmt.Errorf("unsupported host type: '%s'", cloudHostType)
	}

	return hosts, nil
}

func (d *Deployer) updateHostCertificate(ctx context.Context, cloudHostType string, cloudHostId int64, cloudCertId int64) error {
	switch cloudHostType {
	case HOST_TYPE_PROXY:
		{
			nginxUpdateProxyHostReq := &npmsdk.NginxUpdateProxyHostRequest{
				CertificateId: lo.ToPtr(cloudCertId),
			}
			nginxUpdateProxyHostResp, err := d.sdkClient.NginxUpdateProxyHost(cloudHostId, nginxUpdateProxyHostReq)
			d.logger.Debug("sdk request 'nginx.UpdateProxyHost'", slog.Int64("request.hostId", cloudHostId), slog.Any("request", nginxUpdateProxyHostReq), slog.Any("response", nginxUpdateProxyHostResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'nginx.UpdateProxyHost': %w", err)
			}
		}

	case HOST_TYPE_REDIRECTION:
		{
			nginxUpdateRedirectionHostReq := &npmsdk.NginxUpdateRedirectionHostRequest{
				CertificateId: lo.ToPtr(cloudCertId),
			}
			nginxUpdateRedirectionHostResp, err := d.sdkClient.NginxUpdateRedirectionHost(cloudHostId, nginxUpdateRedirectionHostReq)
			d.logger.Debug("sdk request 'nginx.UpdateRedirectionHost'", slog.Int64("request.hostId", cloudHostId), slog.Any("request", nginxUpdateRedirectionHostReq), slog.Any("response", nginxUpdateRedirectionHostResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'nginx.UpdateRedirectionHost': %w", err)
			}
		}

	case HOST_TYPE_STREAM:
		{
			nginxUpdateStreamReq := &npmsdk.NginxUpdateStreamRequest{
				CertificateId: lo.ToPtr(cloudCertId),
			}
			nginxUpdateStreamResp, err := d.sdkClient.NginxUpdateStream(cloudHostId, nginxUpdateStreamReq)
			d.logger.Debug("sdk request 'nginx.UpdateStream'", slog.Int64("request.hostId", cloudHostId), slog.Any("request", nginxUpdateStreamReq), slog.Any("response", nginxUpdateStreamResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'nginx.UpdateStream': %w", err)
			}
		}

	case HOST_TYPE_DEAD:
		{
			nginxUpdateDeadHostReq := &npmsdk.NginxUpdateDeadHostRequest{
				CertificateId: lo.ToPtr(cloudCertId),
			}
			nginxUpdateDeadHostResp, err := d.sdkClient.NginxUpdateDeadHost(cloudHostId, nginxUpdateDeadHostReq)
			d.logger.Debug("sdk request 'nginx.UpdateDeadHost'", slog.Int64("request.hostId", cloudHostId), slog.Any("request", nginxUpdateDeadHostReq), slog.Any("response", nginxUpdateDeadHostResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'nginx.UpdateDeadHost': %w", err)
			}
		}

	default:
		return fmt.Errorf("unsupported host type: '%s'", cloudHostType)
	}

	return nil
}

func createSDKClient(serverUrl, authMethod, username, password, apiToken string, skipTlsVerify bool) (*npmsdk.Client, error) {
	var client *npmsdk.Client
	var err error

	switch authMethod {
	case "", AUTH_METHOD_PASSWORD:
		{
			client, err = npmsdk.NewClient(serverUrl, username, password)
		}

	case AUTH_METHOD_TOKEN:
		{
			client, err = npmsdk.NewClientWithJwtToken(serverUrl, apiToken)
		}
	}

	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
