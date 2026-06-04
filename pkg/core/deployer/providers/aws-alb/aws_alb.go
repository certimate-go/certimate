package awsalb

import (
	"context"
	"fmt"
	"log/slog"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	awscred "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"

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
	// AWS ALB 负载均衡器 ARN。
	LoadbalancerArn string `json:"loadbalancerArn"`
	// AWS ALB 侦听器 ARN。
	ListenerArn string `json:"listenerArn"`
	// AWS ALB 证书来源。
	// 可取值 "ACM"、"IAM"。
	CertificateSource string `json:"certificateSource"`
	// 是否设为默认证书。
	IsDefault bool `json:"isDefault,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *elasticloadbalancingv2.Client
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
	if d.config.LoadbalancerArn == "" {
		return nil, fmt.Errorf("config `loadbalancerArn` is required")
	}
	if d.config.ListenerArn == "" {
		return nil, fmt.Errorf("config `listenerArn` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 查询负载均衡器
	// REF: https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_DescribeLoadBalancers.html
	describeLoadBalancersReq := &elasticloadbalancingv2.DescribeLoadBalancersInput{
		LoadBalancerArns: []string{d.config.LoadbalancerArn},
		PageSize:         aws.Int32(1),
	}
	describeLoadBalancersResp, err := d.sdkClient.DescribeLoadBalancers(ctx, describeLoadBalancersReq)
	d.logger.Debug("sdk request 'elasticloadbalancingv2.DescribeLoadBalancers'", slog.Any("request", describeLoadBalancersReq), slog.Any("response", describeLoadBalancersResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'elasticloadbalancingv2.DescribeLoadBalancers': %w", err)
	} else if len(describeLoadBalancersResp.LoadBalancers) == 0 || describeLoadBalancersResp.LoadBalancers[0].Type != types.LoadBalancerTypeEnumApplication {
		return nil, fmt.Errorf("could not find alb instance '%s'", d.config.LoadbalancerArn)
	}

	// 查询侦听器
	// REF: https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_DescribeListeners.html
	describeListenersReq := &elasticloadbalancingv2.DescribeListenersInput{
		LoadBalancerArn: aws.String(d.config.LoadbalancerArn),
		ListenerArns:    []string{d.config.ListenerArn},
		PageSize:        aws.Int32(1),
	}
	describeListenersResp, err := d.sdkClient.DescribeListeners(ctx, describeListenersReq)
	d.logger.Debug("sdk request 'elasticloadbalancingv2.DescribeListeners'", slog.Any("request", describeListenersReq), slog.Any("response", describeListenersResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'elasticloadbalancingv2.DescribeListeners': %w", err)
	} else if len(describeListenersResp.Listeners) == 0 {
		return nil, fmt.Errorf("could not find alb listener '%s'", d.config.ListenerArn)
	}

	if d.config.IsDefault {
		if describeListenersResp.Listeners[0].Certificates != nil {
			for _, cert := range describeListenersResp.Listeners[0].Certificates {
				if aws.ToString(cert.CertificateArn) == upres.ExtendedData["Arn"].(string) {
					d.logger.Info("no need to update alb listener default certificate")
					return &DeployResult{}, nil
				}
			}
		}

		// 更新 HTTPS 侦听器
		// REF: https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_ModifyListener.html
		modifyListenerReq := &elasticloadbalancingv2.ModifyListenerInput{
			ListenerArn: aws.String(d.config.ListenerArn),
			Certificates: []types.Certificate{
				{
					CertificateArn: aws.String(upres.ExtendedData["Arn"].(string)),
				},
			},
		}
		modifyListenerResp, err := d.sdkClient.ModifyListener(ctx, modifyListenerReq)
		d.logger.Debug("sdk request 'elasticloadbalancingv2.ModifyListener'", slog.Any("request", modifyListenerReq), slog.Any("response", modifyListenerResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'elasticloadbalancingv2.ModifyListener': %w", err)
		}
	} else {
		// 将证书添加到证书列表
		// REF: https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_AddListenerCertificates.html
		addListenerCertificatesReq := &elasticloadbalancingv2.AddListenerCertificatesInput{
			ListenerArn: aws.String(d.config.ListenerArn),
			Certificates: []types.Certificate{
				{
					CertificateArn: aws.String(upres.ExtendedData["Arn"].(string)),
				},
			},
		}
		addListenerCertificatesResp, err := d.sdkClient.AddListenerCertificates(ctx, addListenerCertificatesReq)
		d.logger.Debug("sdk request 'elasticloadbalancingv2.AddListenerCertificates'", slog.Any("request", addListenerCertificatesReq), slog.Any("response", addListenerCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'elasticloadbalancingv2.AddListenerCertificates': %w", err)
		}
	}

	return &DeployResult{}, nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*elasticloadbalancingv2.Client, error) {
	cfg, err := awscfg.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := elasticloadbalancingv2.NewFromConfig(cfg, func(o *elasticloadbalancingv2.Options) {
		o.Region = region
		o.Credentials = aws.NewCredentialsCache(awscred.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, ""))
	})
	return client, nil
}
