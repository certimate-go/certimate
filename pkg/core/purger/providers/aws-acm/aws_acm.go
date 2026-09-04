package awsacm

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
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
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
}

type Purger struct {
	config    *PurgerConfig
	logger    *slog.Logger
	sdkClient *acm.Client
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
	purgingCertArns := make([]string, 0)

	// 获取证书列表，避免重复上传
	// REF: https://docs.aws.amazon.com/acm/latest/APIReference/API_ListCertificates.html
	listCertificatesNextToken := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &acm.ListCertificatesInput{
			NextToken: listCertificatesNextToken,
			MaxItems:  aws.Int32(1000),
		}
		listCertificatesResp, err := p.sdkClient.ListCertificates(ctx, listCertificatesReq)
		p.logger.Debug("sdk request 'acm.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'acm.ListCertificates': %w", err)
		}

		for _, certItem := range listCertificatesResp.CertificateSummaryList {
			if certItem.CertificateKeyPairOrigin == types.CertificateKeyPairOriginAcme {
				continue
			}
			if certItem.NotAfter == nil {
				continue
			}

			certNotAfter := lo.FromPtr(certItem.NotAfter)
			if time.Since(certNotAfter) >= expiry {
				purgingCertArns = append(purgingCertArns, *certItem.CertificateArn)
			}
		}

		if len(listCertificatesResp.CertificateSummaryList) == 0 || listCertificatesResp.NextToken == nil {
			break
		}

		listCertificatesNextToken = listCertificatesResp.NextToken
	}

	// 删除服务器证书
	// REF: https://docs.aws.amazon.com/acm/latest/APIReference/API_DeleteCertificate.html
	purgedCount := 0
	if len(purgingCertArns) > 0 {
		p.logger.Info("found certificates to purge", slog.Any("certArns", purgingCertArns))

		if err := xloop.ForRangeAllWithContext(ctx, purgingCertArns, func(ctx context.Context, certArn string, i int) error {
			if i > 0 {
				if err := xwait.DelayWithContext(ctx, 1*time.Second); err != nil {
					return err
				}
			}

			deleteServerCertificateReq := &acm.DeleteCertificateInput{
				CertificateArn: aws.String(certArn),
			}
			deleteServerCertificateResp, err := p.sdkClient.DeleteCertificate(ctx, deleteServerCertificateReq)
			p.logger.Debug("sdk request 'acm.DeleteCertificate'", slog.Any("request", deleteServerCertificateReq), slog.Any("response", deleteServerCertificateResp))
			if err != nil {
				var sdkErr smithy.APIError
				if errors.As(err, &sdkErr) {
					if sdkErrCode := sdkErr.ErrorCode(); sdkErrCode == "ResourceNotFoundException" || sdkErrCode == "ResourceInUseException" {
						return nil
					}
				}

				return fmt.Errorf("failed to execute sdk request 'acm.DeleteCertificate': %w", err)
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

func createSDKClient(authMethod, accessKeyId, secretAccessKey, region string) (*acm.Client, error) {
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
