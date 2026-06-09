package certacme

import (
	"context"
	"crypto"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/challenge/http01"
	"github.com/go-acme/lego/v5/log"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/certacme/certifiers"
	"github.com/certimate-go/certimate/internal/domain"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type ObtainCertificateRequest struct {
	DomainOrIPs       []string
	PrivateKeyType    certcrypto.KeyType
	PrivateKeyPEM     string
	ValidityNotBefore time.Time
	ValidityNotAfter  time.Time
	NoCommonName      bool

	// 提供商相关
	ChallengeType          string
	Provider               domain.ACMEChallengeProviderType
	ProviderAccessConfig   map[string]any
	ProviderExtendedConfig map[string]any

	// 解析相关
	DisableFollowCNAME bool
	Nameservers        []string

	// DNS-01 质询相关
	DnsPropagationWait    int
	DnsPropagationTimeout int
	DnsTTL                int

	// HTTP-01 质询相关
	HttpDelayWait int

	// ACME 相关
	PreferredChain string
	ACMEProfile    string

	// ARI 相关
	ARIReplacesAccountUrl string
	ARIReplacesCertId     string
}

type ObtainCertificateResponse struct {
	CAProvider           domain.CAProviderType
	CSR                  string
	FullChainCertificate string
	IssuerCertificate    string
	PrivateKey           string
	ACMEAccountUrl       string
	ACMECertificateUrl   string
	ARIReplaced          bool
}

func (c *ACMEClient) ObtainCertificate(ctx context.Context, request *ObtainCertificateRequest) (*ObtainCertificateResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("the request is nil")
	}

	os.Setenv("LEGO_DISABLE_CNAME_SUPPORT", strconv.FormatBool(request.DisableFollowCNAME))

	const CHALLENGE_TYPE_DNS01 = "dns-01"
	const CHALLENGE_TYPE_HTTP01 = "http-01"
	switch strings.ToLower(request.ChallengeType) {
	case CHALLENGE_TYPE_DNS01:
		{
			providerFactory, err := certifiers.ACMEDns01Registries.Get(domain.ACMEDns01ProviderType(request.Provider))
			if err != nil {
				return nil, err
			}

			provider, err := providerFactory(&certifiers.ProviderFactoryOptions{
				ProviderAccessConfig:   request.ProviderAccessConfig,
				ProviderExtendedConfig: request.ProviderExtendedConfig,
				DnsPropagationTimeout:  request.DnsPropagationTimeout,
				DnsTTL:                 request.DnsTTL,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to initialize dns-01 provider '%s': %w", request.Provider, err)
			}

			opts := &dns01.Options{}
			opts.RecursiveNameservers = request.Nameservers
			dns01.SetDefaultClient(dns01.NewClient(opts))
			c.client.Challenge.SetDNS01Provider(provider,
				dns01.CondOptions(
					request.DnsPropagationWait > 0,
					dns01.PropagationWait(time.Duration(request.DnsPropagationWait)*time.Second, true),
				),
				dns01.CondOptions(
					len(request.Nameservers) > 0 || request.DnsPropagationWait > 0,
					dns01.DisableAuthoritativeNssPropagationRequirement(),
				),
			)
		}

	case CHALLENGE_TYPE_HTTP01:
		{
			providerFactory, err := certifiers.ACMEHttp01Registries.Get(domain.ACMEHttp01ProviderType(request.Provider))
			if err != nil {
				return nil, err
			}

			provider, err := providerFactory(&certifiers.ProviderFactoryOptions{
				ProviderAccessConfig:   request.ProviderAccessConfig,
				ProviderExtendedConfig: request.ProviderExtendedConfig,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to initialize http-01 provider '%s': %w", request.Provider, err)
			}

			c.client.Challenge.SetHTTP01Provider(provider,
				http01.SetDelay(time.Duration(request.HttpDelayWait)*time.Second),
			)
		}

	default:
		return nil, fmt.Errorf("unsupported challenge type: '%s'", request.ChallengeType)
	}

	var privkey crypto.Signer
	if request.PrivateKeyPEM != "" {
		pk, err := certcrypto.ParsePEMPrivateKey([]byte(request.PrivateKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		privkey = pk
	}

	req := certificate.ObtainRequest{
		Domains:          request.DomainOrIPs,
		KeyType:          request.PrivateKeyType,
		PrivateKey:       privkey,
		Bundle:           true,
		EnableCommonName: !request.NoCommonName,
		PreferredChain:   request.PreferredChain,
		Profile:          request.ACMEProfile,
		NotBefore:        request.ValidityNotBefore,
		NotAfter:         request.ValidityNotAfter,
		ReplacesCertID:   lo.If(request.ARIReplacesAccountUrl == c.account.ACMEAccountUrl, request.ARIReplacesCertId).Else(""),
	}
	resp, err := c.client.Certificate.Obtain(ctx, req)
	if err != nil {
		ariErr := &acme.AlreadyReplacedError{}
		if !errors.As(err, &ariErr) {
			return nil, err
		}

		log.Warn("the certificate has already been replaced, try to obtain again without ARI ...")

		// reset ARI and retry if failure
		req.ReplacesCertID = ""
		resp, err = c.client.Certificate.Obtain(ctx, req)
		if err != nil {
			return nil, err
		}
	}

	// lego 自 v5 起返回的私钥 PEM 内容使用 PKCS#8 格式编码，
	// 这里转换为 PKCS#1 或 SEC1 格式编码，以满足更好的兼容性。
	privkeyPEM := strings.TrimSpace(string(resp.PrivateKey))
	if t1, err := xcert.ParsePrivateKeyFromPEM(privkeyPEM); err == nil {
		if t2, err := xcert.ConvertPrivateKeyToPEM(t1, false); err == nil {
			privkeyPEM = t2
		}
	}

	return &ObtainCertificateResponse{
		CAProvider:           domain.CAProviderType(c.account.CA),
		CSR:                  strings.TrimSpace(string(resp.CSR)),
		FullChainCertificate: strings.TrimSpace(string(resp.Certificate)),
		IssuerCertificate:    strings.TrimSpace(string(resp.IssuerCertificate)),
		PrivateKey:           privkeyPEM,
		ACMEAccountUrl:       c.account.ACMEAccountUrl,
		ACMECertificateUrl:   resp.CertURL,
		ARIReplaced:          req.ReplacesCertID != "",
	}, nil
}
