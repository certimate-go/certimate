package linodelos

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core"
	linodesdk "github.com/certimate-go/certimate/pkg/sdk3rd/linode"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// Linode AccessToken。
	AccessToken string `json:"accessToken"`
	// 对象存储区域 ID。
	RegionId string `json:"regionId"`
	// 对象存储桶名。
	Bucket string `json:"bucket"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *linodesdk.Client
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessToken)
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	if d.config.RegionId == "" {
		return nil, fmt.Errorf("config `regionId` is required")
	}
	if d.config.Bucket == "" {
		return nil, fmt.Errorf("config `bucket` is required")
	}

	// Get an Object Storage TLS/SSL certificate
	// REF: https://techdocs.akamai.com/linode-api/reference/get-object-storage-ssl
	getObjectStorageSSLResp, err := d.sdkClient.GetObjectStorageSSLWithContext(ctx, d.config.RegionId, d.config.Bucket)
	d.logger.Debug("sdk request 'objectstorage.GetSSL'", slog.String("params.regionId", d.config.RegionId), slog.String("params.bucket", d.config.Bucket), slog.Any("response", getObjectStorageSSLResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'objectstorage.GetSSL': %w", err)
	}

	// Delete an Object Storage TLS/SSL certificate
	// REF: https://techdocs.akamai.com/linode-api/reference/delete-object-storage-ssl
	if getObjectStorageSSLResp.SSL {
		deleteObjectStorageSSLResp, err := d.sdkClient.DeleteObjectStorageSSLWithContext(ctx, d.config.RegionId, d.config.Bucket)
		d.logger.Debug("sdk request 'objectstorage.DeleteSSL'", slog.String("params.regionId", d.config.RegionId), slog.String("params.bucket", d.config.Bucket), slog.Any("response", deleteObjectStorageSSLResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'objectstorage.DeleteSSL': %w", err)
		}
	}

	// Upload an Object Storage TLS/SSL certificate
	// REF: https://techdocs.akamai.com/linode-api/reference/post-object-storage-ssl
	uploadObjectStorageSSLReq := &linodesdk.UploadObjectStorageSSLRequest{
		Certificate: certPEM,
		PrivateKey:  privkeyPEM,
	}
	uploadObjectStorageSSLResp, err := d.sdkClient.UploadObjectStorageSSLWithContext(ctx, d.config.RegionId, d.config.Bucket, uploadObjectStorageSSLReq)
	d.logger.Debug("sdk request 'objectstorage.UploadSSL'", slog.String("params.regionId", d.config.RegionId), slog.String("params.bucket", d.config.Bucket), slog.Any("request", uploadObjectStorageSSLReq), slog.Any("response", uploadObjectStorageSSLResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'objectstorage.UploadSSL': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(accessToken string) (*linodesdk.Client, error) {
	client, err := linodesdk.NewClient(
		linodesdk.WithAccessToken(accessToken),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
