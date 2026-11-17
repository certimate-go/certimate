package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	awsstd "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awscred "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	smithy "github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const (
	authMethodAccessKey = "accesskey"
	authMethodIMDS      = "imds"
	awsRegion           = "us-east-1"
	defaultSTSHost      = "sts.us-east-1.amazonaws.com"
	defaultIMDSHost     = "169.254.169.254"
)

type checker struct {
	stsEndpoint  string
	imdsEndpoint string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeAWS)
}

func (c *checker) Scope() []string {
	return []string{core.AccessCheckItemConnect, core.AccessCheckItemAuth}
}

func (c *checker) Check(ctx context.Context, config map[string]any) (*CheckResult, error) {
	ctx, cancel := accesschecker.WithBudget(ctx)
	defer cancel()

	cfg, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	diagnostics := []string{stsHost(c.stsEndpoint)}
	var status int
	var code string
	var callErr error
	secrets := []string{cfg.AccessKeyId, cfg.SecretAccessKey}
	switch cfg.AuthMethod {
	case "", authMethodAccessKey:
		status, code, callErr = c.getCallerIdentity(ctx, cfg.AccessKeyId, cfg.SecretAccessKey, "")
	case authMethodIMDS:
		diagnostics = []string{imdsHost(c.imdsEndpoint), stsHost(c.stsEndpoint)}
		status, code, callErr, secrets = c.identityFromIMDS(ctx, secrets)
	default:
		return nil, fmt.Errorf("%w: authMethod is invalid", core.ErrAccessCheckConfigInvalid)
	}
	return accesschecker.Finish(callErr, status, code, diagnostics, secrets...), nil
}

type checkConfig struct {
	AuthMethod      string
	AccessKeyId     string
	SecretAccessKey string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForAWS
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		AuthMethod:      strings.ToLower(strings.TrimSpace(wire.AuthMethod)),
		AccessKeyId:     strings.TrimSpace(wire.AccessKeyId),
		SecretAccessKey: wire.SecretAccessKey,
	}
	switch cfg.AuthMethod {
	case "", authMethodAccessKey:
		cfg.AuthMethod = authMethodAccessKey
		if cfg.AccessKeyId == "" {
			return nil, fmt.Errorf("%w: accessKeyId is required", core.ErrAccessCheckConfigInvalid)
		}
		if strings.TrimSpace(cfg.SecretAccessKey) == "" {
			return nil, fmt.Errorf("%w: secretAccessKey is required", core.ErrAccessCheckConfigInvalid)
		}
	case authMethodIMDS:
	default:
		return nil, fmt.Errorf("%w: authMethod is invalid", core.ErrAccessCheckConfigInvalid)
	}
	return cfg, nil
}

func (c *checker) identityFromIMDS(ctx context.Context, secrets []string) (int, string, error, []string) {
	client := imds.New(imds.Options{
		Endpoint:              c.imdsEndpoint,
		ClientEnableState:     imds.ClientEnabled,
		DisableDefaultTimeout: true,
		Retryer:               awsstd.NopRetryer{},
	})
	roleOut, err := client.GetMetadata(ctx, &imds.GetMetadataInput{Path: "iam/security-credentials/"})
	if err != nil {
		status, code := awsFault(err)
		if status == 0 && code == "" {
			return 0, "", err, secrets
		}
		if status == 401 || status == 403 || status == 404 {
			return status, "InvalidCredential", err, secrets
		}
		return status, code, err, secrets
	}
	role, err := readLimited(roleOut.Content)
	if err != nil {
		return 0, "", err, secrets
	}
	role = strings.TrimSpace(role)
	if role == "" || strings.ContainsAny(role, "\r\n") {
		return 401, "InvalidCredential", errors.New("imds returned no iam role"), secrets
	}

	credOut, err := client.GetMetadata(ctx, &imds.GetMetadataInput{Path: "iam/security-credentials/" + role})
	if err != nil {
		status, code := awsFault(err)
		if status == 401 || status == 403 || status == 404 {
			return status, "InvalidCredential", err, secrets
		}
		return status, code, err, secrets
	}
	raw, err := readLimited(credOut.Content)
	if err != nil {
		return 0, "", err, secrets
	}
	var creds struct {
		Code            string
		AccessKeyId     string
		SecretAccessKey string
		Token           string
	}
	if err := json.Unmarshal([]byte(raw), &creds); err != nil || creds.Code != "Success" || creds.AccessKeyId == "" || creds.SecretAccessKey == "" {
		return 401, "InvalidCredential", errors.New("imds credentials were not accepted"), secrets
	}
	secrets = append(secrets, creds.AccessKeyId, creds.SecretAccessKey, creds.Token)
	status, code, err := c.getCallerIdentity(ctx, creds.AccessKeyId, creds.SecretAccessKey, creds.Token)
	return status, code, err, secrets
}

func (c *checker) getCallerIdentity(ctx context.Context, accessKeyId, secretAccessKey, sessionToken string) (int, string, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(awsRegion),
		awsconfig.WithCredentialsProvider(awscred.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, sessionToken)),
		awsconfig.WithRetryer(func() awsstd.Retryer { return awsstd.NopRetryer{} }),
	)
	if err != nil {
		return 0, "", err
	}
	client := sts.NewFromConfig(awsCfg, func(o *sts.Options) {
		if c.stsEndpoint != "" {
			o.BaseEndpoint = awsstd.String(c.stsEndpoint)
		}
	})
	_, err = client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		status, code := awsFault(err)
		return status, code, err
	}
	return 200, "", nil
}

func awsFault(err error) (int, string) {
	status := 0
	code := ""
	var respErr *smithyhttp.ResponseError
	if errors.As(err, &respErr) && respErr != nil {
		status = respErr.HTTPStatusCode()
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr != nil {
		code = apiErr.ErrorCode()
	}
	return status, code
}

func readLimited(body io.ReadCloser) (string, error) {
	defer body.Close()
	data, err := io.ReadAll(io.LimitReader(body, 1<<20))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func stsHost(endpoint string) string {
	if endpoint == "" {
		return defaultSTSHost
	}
	return accesschecker.HostOf(endpoint)
}

func imdsHost(endpoint string) string {
	if endpoint == "" {
		return defaultIMDSHost
	}
	return accesschecker.HostOf(endpoint)
}
