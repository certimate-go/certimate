package awsalb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	awscred "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/smithy-go"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimplacm "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aws-acm"
	cmgrimpliam "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aws-iam"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// AWS API 认证方式。
	// 可取值 "accesskey"、"imds"。
	// 零值时默认值 [AUTH_METHOD_ACCESSKEY]。
	AuthMethod string `json:"authMethod,omitempty"`
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
	// 是否自动移除同域名的其他证书。
	AutoPrune bool `json:"autoPrune,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClients *wSDKClients
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

type wSDKClients struct {
	ACM *acm.Client
	IAM *iam.Client
	ELB *elasticloadbalancingv2.Client
}

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	clientELB, err := createSDKClientELB(config.AuthMethod, config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	clientACM, err := createSDKClientACM(config.AuthMethod, config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	clientIAM, err := createSDKClientIAM(config.AuthMethod, config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	var pcertmgr core.Certmgr
	switch config.CertificateSource {
	case CERTIFICATE_SOURCE_ACM:
		pcertmgr, err = cmgrimplacm.NewCertmgr(&cmgrimplacm.CertmgrConfig{
			AuthMethod:      config.AuthMethod,
			AccessKeyId:     config.AccessKeyId,
			SecretAccessKey: config.SecretAccessKey,
			Region:          config.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("could not create certmgr: %w", err)
		}

	case CERTIFICATE_SOURCE_IAM:
		pcertmgr, err = cmgrimpliam.NewCertmgr(&cmgrimpliam.CertmgrConfig{
			AuthMethod:      config.AuthMethod,
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
		sdkClients: &wSDKClients{ELB: clientELB, ACM: clientACM, IAM: clientIAM},
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

	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
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
	}
	describeLoadBalancersResp, err := d.sdkClients.ELB.DescribeLoadBalancers(ctx, describeLoadBalancersReq)
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
	}
	describeListenersResp, err := d.sdkClients.ELB.DescribeListeners(ctx, describeListenersReq)
	d.logger.Debug("sdk request 'elasticloadbalancingv2.DescribeListeners'", slog.Any("request", describeListenersReq), slog.Any("response", describeListenersResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'elasticloadbalancingv2.DescribeListeners': %w", err)
	} else if len(describeListenersResp.Listeners) == 0 {
		return nil, fmt.Errorf("could not find alb listener '%s'", d.config.ListenerArn)
	}

	listenerInfo := describeListenersResp.Listeners[0]
	if len(listenerInfo.Certificates) > 0 {
		d.logger.Info("found alb listener certificates in used", slog.Any("certificates", listenerInfo.Certificates))
	}

	if d.config.IsDefault {
		certArn := upres.ExtendedData["Arn"].(string)
		for _, certItem := range listenerInfo.Certificates {
			if aws.ToString(certItem.CertificateArn) == certArn && aws.ToBool(certItem.IsDefault) {
				d.logger.Info("no need to deploy alb listener default certificate")
				return &DeployResult{}, nil
			}
		}

		if err := d.updateListenerDefaultCertificate(ctx, *listenerInfo.ListenerArn, certArn); err != nil {
			return nil, err
		}
	} else {
		certArn := upres.ExtendedData["Arn"].(string)
		for _, certItem := range listenerInfo.Certificates {
			if aws.ToString(certItem.CertificateArn) == certArn && !aws.ToBool(certItem.IsDefault) {
				d.logger.Info("no need to deploy alb listener sni certificate")
				return &DeployResult{}, nil
			}
		}

		if err := d.updateListenerSniCertificate(ctx, *listenerInfo.ListenerArn, certArn); err != nil {
			return nil, err
		}

		if d.config.AutoPrune {
			if err := d.pruneListenerSniCertificates(ctx, *listenerInfo.ListenerArn, certArn, certX509.DNSNames); err != nil {
				return nil, err
			}
		}
	}

	return &DeployResult{}, nil
}

func (d *Deployer) updateListenerDefaultCertificate(ctx context.Context, cloudListenerArn string, cloudCertArn string) error {
	// 更新 HTTPS 侦听器
	// REF: https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_ModifyListener.html
	modifyListenerReq := &elasticloadbalancingv2.ModifyListenerInput{
		ListenerArn: aws.String(cloudListenerArn),
		Certificates: []types.Certificate{
			{
				CertificateArn: aws.String(cloudCertArn),
			},
		},
	}
	modifyListenerResp, err := d.sdkClients.ELB.ModifyListener(ctx, modifyListenerReq)
	d.logger.Debug("sdk request 'elasticloadbalancingv2.ModifyListener'", slog.Any("request", modifyListenerReq), slog.Any("response", modifyListenerResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'elasticloadbalancingv2.ModifyListener': %w", err)
	}

	return nil
}

func (d *Deployer) updateListenerSniCertificate(ctx context.Context, cloudListenerArn string, cloudCertArn string) error {
	// 将证书添加到证书列表
	// REF: https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_AddListenerCertificates.html
	addListenerCertificatesReq := &elasticloadbalancingv2.AddListenerCertificatesInput{
		ListenerArn: aws.String(cloudListenerArn),
		Certificates: []types.Certificate{
			{
				CertificateArn: aws.String(cloudCertArn),
			},
		},
	}
	addListenerCertificatesResp, err := d.sdkClients.ELB.AddListenerCertificates(ctx, addListenerCertificatesReq)
	d.logger.Debug("sdk request 'elasticloadbalancingv2.AddListenerCertificates'", slog.Any("request", addListenerCertificatesReq), slog.Any("response", addListenerCertificatesResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'elasticloadbalancingv2.AddListenerCertificates': %w", err)
	}

	return nil
}

func (d *Deployer) pruneListenerSniCertificates(ctx context.Context, cloudListenerArn string, cloudCertArn string, cloudCertSANs []string) error {
	const acmArnPrefix = "arn:aws:acm:"
	const iamArnPrefix = "arn:aws:iam:"

	certArnsToDelete := make([]string, 0)

	// 查询侦听器证书列表，并找出需要删除绑定的同域名证书
	// REF: https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_DescribeListenerCertificates.html
	// REF: https://docs.aws.amazon.com/IAM/latest/APIReference/API_GetServerCertificate.html
	// REF: https://docs.aws.amazon.com/acm/latest/APIReference/API_GetCertificate.html
	describeListenerCertificatesMarker := (*string)(nil)
	describeListenerCertificatesPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		describeListenerCertificatesReq := &elasticloadbalancingv2.DescribeListenerCertificatesInput{
			ListenerArn: aws.String(cloudListenerArn),
			Marker:      describeListenerCertificatesMarker,
			PageSize:    aws.Int32(int32(describeListenerCertificatesPageSize)),
		}
		describeListenerCertificatesResp, err := d.sdkClients.ELB.DescribeListenerCertificates(ctx, describeListenerCertificatesReq)
		d.logger.Debug("sdk request 'elasticloadbalancingv2.DescribeListenerCertificates'", slog.Any("request", describeListenerCertificatesReq), slog.Any("response", describeListenerCertificatesResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'elasticloadbalancingv2.DescribeListenerCertificates': %w", err)
		}

		for _, certItem := range describeListenerCertificatesResp.Certificates {
			if aws.ToString(certItem.CertificateArn) == cloudCertArn {
				continue
			}
			if aws.ToBool(certItem.IsDefault) {
				continue
			}

			switch {
			case strings.HasPrefix(aws.ToString(certItem.CertificateArn), acmArnPrefix):
				{
					getCertificateReq := &acm.GetCertificateInput{
						CertificateArn: certItem.CertificateArn,
					}
					getCertificateResp, err := d.sdkClients.ACM.GetCertificate(ctx, getCertificateReq)
					if err != nil {
						var sdkerr smithy.APIError
						if errors.As(err, &sdkerr) {
							if sdkErrCode := sdkerr.ErrorCode(); sdkErrCode == "InvalidArnException" || sdkErrCode == "ResourceNotFoundException" {
								continue
							}
						}

						return fmt.Errorf("failed to execute sdk request 'acm.GetCertificate': %w", err)
					}

					certX509InACM, err := xcert.ParseCertificateFromPEM(aws.ToString(getCertificateResp.Certificate))
					if err != nil {
						continue
					}

					if lo.ElementsMatch(cloudCertSANs, certX509InACM.DNSNames) {
						certArnsToDelete = append(certArnsToDelete, aws.ToString(certItem.CertificateArn))
					}
				}

			case strings.HasPrefix(aws.ToString(certItem.CertificateArn), iamArnPrefix):
				{
					getServerCertificateReq := &iam.GetServerCertificateInput{
						ServerCertificateName: certItem.CertificateArn,
					}
					getServerCertificateResp, err := d.sdkClients.IAM.GetServerCertificate(ctx, getServerCertificateReq)
					if err != nil {
						var sdkerr smithy.APIError
						if errors.As(err, &sdkerr) {
							if sdkerr.ErrorCode() == "NoSuchEntity" {
								continue
							}
						}

						return fmt.Errorf("failed to execute sdk request 'iam.GetServerCertificate': %w", err)
					}

					certX509InIAM, err := xcert.ParseCertificateFromPEM(aws.ToString(getServerCertificateResp.ServerCertificate.CertificateBody))
					if err != nil {
						continue
					}

					if lo.ElementsMatch(cloudCertSANs, certX509InIAM.DNSNames) {
						certArnsToDelete = append(certArnsToDelete, aws.ToString(certItem.CertificateArn))
					}
				}
			}
		}

		if len(describeListenerCertificatesResp.Certificates) < describeListenerCertificatesPageSize || describeListenerCertificatesResp.NextMarker == nil {
			break
		}

		describeListenerCertificatesMarker = describeListenerCertificatesResp.NextMarker
	}

	// 从证书列表删除证书
	// REF: https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_RemoveListenerCertificates.html
	if len(certArnsToDelete) > 0 {
		d.logger.Info("found alb listener certificates to delete", slog.Any("certificateArns", certArnsToDelete))

		removeListenerCertificatesReq := &elasticloadbalancingv2.RemoveListenerCertificatesInput{
			ListenerArn: aws.String(d.config.ListenerArn),
			Certificates: lo.Map(certArnsToDelete, func(certArn string, _ int) types.Certificate {
				return types.Certificate{
					CertificateArn: aws.String(certArn),
				}
			}),
		}
		removeListenerCertificatesResp, err := d.sdkClients.ELB.RemoveListenerCertificates(ctx, removeListenerCertificatesReq)
		d.logger.Debug("sdk request 'elasticloadbalancingv2.RemoveListenerCertificates'", slog.Any("request", removeListenerCertificatesReq), slog.Any("response", removeListenerCertificatesResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'elasticloadbalancingv2.RemoveListenerCertificates': %w", err)
		}
	}

	return nil
}

func createSDKClientELB(authMethod, accessKeyId, secretAccessKey, region string) (*elasticloadbalancingv2.Client, error) {
	opts := []func(options *awscfg.LoadOptions) error{
		awscfg.WithRegion(region),
	}

	staticCredsProvider := awscred.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")
	imdsCredsProvider := aws.NewCredentialsCache(ec2rolecreds.New())
	switch authMethod {
	case "":
		if accessKeyId != "" && secretAccessKey != "" {
			opts = append(opts, awscfg.WithCredentialsProvider(staticCredsProvider))
		}
	case AUTH_METHOD_ACCESSKEY:
		opts = append(opts, awscfg.WithCredentialsProvider(staticCredsProvider))
	case AUTH_METHOD_IMDS:
		opts = append(opts, awscfg.WithCredentialsProvider(imdsCredsProvider))
	default:
		return nil, fmt.Errorf("unsupported auth method '%s'", authMethod)
	}

	cfg, err := awscfg.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	client := elasticloadbalancingv2.NewFromConfig(cfg)
	return client, nil
}

func createSDKClientACM(authMethod, accessKeyId, secretAccessKey, region string) (*acm.Client, error) {
	opts := []func(options *awscfg.LoadOptions) error{
		awscfg.WithRegion(region),
	}

	staticCredsProvider := awscred.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")
	imdsCredsProvider := aws.NewCredentialsCache(ec2rolecreds.New())
	switch authMethod {
	case "":
		if accessKeyId != "" && secretAccessKey != "" {
			opts = append(opts, awscfg.WithCredentialsProvider(staticCredsProvider))
		}
	case AUTH_METHOD_ACCESSKEY:
		opts = append(opts, awscfg.WithCredentialsProvider(staticCredsProvider))
	case AUTH_METHOD_IMDS:
		opts = append(opts, awscfg.WithCredentialsProvider(imdsCredsProvider))
	default:
		return nil, fmt.Errorf("unsupported auth method '%s'", authMethod)
	}

	cfg, err := awscfg.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	client := acm.NewFromConfig(cfg)
	return client, nil
}

func createSDKClientIAM(authMethod, accessKeyId, secretAccessKey, region string) (*iam.Client, error) {
	opts := []func(options *awscfg.LoadOptions) error{
		awscfg.WithRegion(region),
	}

	staticCredsProvider := awscred.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")
	imdsCredsProvider := aws.NewCredentialsCache(ec2rolecreds.New())
	switch authMethod {
	case "":
		if accessKeyId != "" && secretAccessKey != "" {
			opts = append(opts, awscfg.WithCredentialsProvider(staticCredsProvider))
		}
	case AUTH_METHOD_ACCESSKEY:
		opts = append(opts, awscfg.WithCredentialsProvider(staticCredsProvider))
	case AUTH_METHOD_IMDS:
		opts = append(opts, awscfg.WithCredentialsProvider(imdsCredsProvider))
	default:
		return nil, fmt.Errorf("unsupported auth method '%s'", authMethod)
	}

	cfg, err := awscfg.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	client := iam.NewFromConfig(cfg)
	return client, nil
}
