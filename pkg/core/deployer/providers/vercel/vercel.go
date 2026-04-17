package vercel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	vercelsdk "github.com/certimate-go/certimate/pkg/sdk3rd/vercel"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type DeployerConfig struct {
	// Vercel API AccessToken。
	ApiAccessToken string `json:"apiAccessToken"`
	// Vercel Team ID。
	TeamId string `json:"teamId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *vercelsdk.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ApiAccessToken, config.TeamId)
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
	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 上传证书
	// REF: https://vercel.com/docs/rest-api/certs/upload-a-cert
	uploadCertReq := &vercelsdk.UploadCertParams{
		CA:             intermediaCertPEM,
		Cert:           serverCertPEM,
		Key:            privkeyPEM,
		SkipValidation: true,
	}
	uploadCertResp, err := d.sdkClient.UploadCertWithContext(ctx, uploadCertReq)
	d.logger.Debug("sdk request 'vercel.UploadCert'", slog.Any("request", uploadCertReq), slog.Any("response", uploadCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'vercel.UploadCert': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(apiToken, teamId string) (*vercelsdk.Client, error) {
	return vercelsdk.NewClientWithTeam(apiToken, teamId)
}
