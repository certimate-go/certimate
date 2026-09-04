package awsiam

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	awscred "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/smithy-go"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider    = core.Purger
	PurgeResult = core.PurgerPurgeResult
)

type PurgerConfig struct {
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
	// IAM 证书路径。
	// 选填。
	CertificatePath string `json:"certificatePath,omitempty"`
}

type Purger struct {
	config    *PurgerConfig
	logger    *slog.Logger
	sdkClient *iam.Client
}

var _ Provider = (*Purger)(nil)

func NewPurger(config *PurgerConfig) (*Purger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the purger provider is nil")
	}

	client, err := createSDKClient(config.AuthMethod, config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Purger{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (p *Purger) SetLogger(logger *slog.Logger) {
	if logger == nil {
		p.logger = slog.New(slog.DiscardHandler)
	} else {
		p.logger = logger
	}
}

func (p *Purger) Purge(ctx context.Context, expiry time.Duration) (*PurgeResult, error) {
	purgingCertNames := make([]string, 0)

	// 获取服务器证书列表
	// REF: https://docs.aws.amazon.com/IAM/latest/APIReference/API_ListServerCertificates.html
	listServerCertificatesMarker := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listServerCertificatesReq := &iam.ListServerCertificatesInput{
			PathPrefix: lo.EmptyableToPtr(p.config.CertificatePath),
			Marker:     listServerCertificatesMarker,
			MaxItems:   aws.Int32(1000),
		}
		listServerCertificatesResp, err := p.sdkClient.ListServerCertificates(ctx, listServerCertificatesReq)
		p.logger.Debug("sdk request 'iam.ListServerCertificates'", slog.Any("request", listServerCertificatesReq), slog.Any("response", listServerCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'iam.ListServerCertificates': %w", err)
		}

		for _, certItem := range listServerCertificatesResp.ServerCertificateMetadataList {
			if certItem.Expiration == nil {
				continue
			}

			certNotAfter := lo.FromPtr(certItem.Expiration)
			if time.Since(certNotAfter) >= expiry {
				purgingCertNames = append(purgingCertNames, *certItem.ServerCertificateName)
			}
		}

		if len(listServerCertificatesResp.ServerCertificateMetadataList) == 0 || listServerCertificatesResp.Marker == nil {
			break
		}

		listServerCertificatesMarker = listServerCertificatesResp.Marker
	}

	// 删除服务器证书
	// REF: https://docs.aws.amazon.com/IAM/latest/APIReference/API_DeleteServerCertificate.html
	purgedCount := 0
	if len(purgingCertNames) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certNames", purgingCertNames))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertNames, func(ctx context.Context, certName string, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			deleteServerCertificateReq := &iam.DeleteServerCertificateInput{
				ServerCertificateName: aws.String(certName),
			}
			deleteServerCertificateResp, err := p.sdkClient.DeleteServerCertificate(ctx, deleteServerCertificateReq)
			p.logger.Debug("sdk request 'keyvault.DeleteCertificate'", slog.Any("request", deleteServerCertificateReq), slog.Any("response", deleteServerCertificateResp))
			if err != nil {
				var sdkErr smithy.APIError
				if errors.As(err, &sdkErr) {
					if sdkErrCode := sdkErr.ErrorCode(); sdkErrCode == "NoSuchEntity" || sdkErrCode == "DeleteConflict" {
						return nil
					}
				}

				return fmt.Errorf("failed to execute sdk request 'keyvault.DeleteCertificate': %w", err)
			}

			purgedCount++
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return &PurgeResult{
		Count: purgedCount,
	}, nil
}

func createSDKClient(authMethod, accessKeyId, secretAccessKey, region string) (*iam.Client, error) {
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
