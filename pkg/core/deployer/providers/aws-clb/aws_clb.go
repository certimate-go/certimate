package awsclb

import (
	"context"
	"fmt"
	"log/slog"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	awscred "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimplacm "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aws-acm"
	cmgrimpliam "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aws-iam"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// AWS AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// AWS SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// AWS 区域。
	Region string `json:"region"`
	// AWS CLB 负载均衡器名称。
	LoadbalancerName string `json:"loadbalancerName"`
	// AWS CLB 负载均衡器端口。
	LoadbalancerPort int32 `json:"loadbalancerPort"`
	// AWS CLB 证书来源。
	// 可取值 "ACM"、"IAM"。
	CertificateSource string `json:"certificateSource"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *elasticloadbalancing.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	var pcertmgr core.Certmgr
	switch config.CertificateSource {
	case CERTIFICATE_SOURCE_ACM:
		pcertmgr, err = cmgrimplacm.NewCertmgr(&cmgrimplacm.CertmgrConfig{
			AccessKeyId:     config.AccessKeyId,
			SecretAccessKey: config.SecretAccessKey,
			Region:          config.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("could not create certmgr: %w", err)
		}

	case CERTIFICATE_SOURCE_IAM:
		pcertmgr, err = cmgrimpliam.NewCertmgr(&cmgrimpliam.CertmgrConfig{
			AccessKeyId:     config.AccessKeyId,
			SecretAccessKey: config.SecretAccessKey,
			Region:          config.Region,
			CertificatePath: "/elb/",
		})
		if err != nil {
			return nil, fmt.Errorf("could not create certmgr: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported certificate source: '%s'", config.CertificateSource)
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
	if d.config.LoadbalancerName == "" {
		return nil, fmt.Errorf("config `loadbalancerName` is required")
	}
	if d.config.LoadbalancerPort == 0 {
		return nil, fmt.Errorf("config `loadbalancerPort` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 替换 HTTPS 侦听器 SSL 证书
	// REF: https://docs.aws.amazon.com/elasticloadbalancing/2012-06-01/APIReference/API_SetLoadBalancerListenerSSLCertificate.html
	setLoadBalancerListenerSSLCertificateReq := &elasticloadbalancing.SetLoadBalancerListenerSSLCertificateInput{
		LoadBalancerName: aws.String(d.config.LoadbalancerName),
		LoadBalancerPort: d.config.LoadbalancerPort,
		SSLCertificateId: aws.String(upres.ExtendedData["Arn"].(string)),
	}
	setLoadBalancerListenerSSLCertificateResp, err := d.sdkClient.SetLoadBalancerListenerSSLCertificate(ctx, setLoadBalancerListenerSSLCertificateReq)
	d.logger.Debug("sdk request 'elasticloadbalancing.SetLoadBalancerListenerSSLCertificate'", slog.Any("request", setLoadBalancerListenerSSLCertificateReq), slog.Any("response", setLoadBalancerListenerSSLCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'elasticloadbalancing.SetLoadBalancerListenerSSLCertificate': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*elasticloadbalancing.Client, error) {
	cfg, err := awscfg.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := elasticloadbalancing.NewFromConfig(cfg, func(o *elasticloadbalancing.Options) {
		o.Region = region
		o.Credentials = aws.NewCredentialsCache(awscred.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, ""))
	})
	return client, nil
}
