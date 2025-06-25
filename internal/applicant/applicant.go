package applicant

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"golang.org/x/exp/slices"
	"golang.org/x/time/rate"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
	xslices "github.com/certimate-go/certimate/pkg/utils/slices"
)

type ApplyResult struct {
	CSR                  string
	FullChainCertificate string
	IssuerCertificate    string
	PrivateKey           string
	ACMEAccountUrl       string
	ACMECertUrl          string
	ACMECertStableUrl    string
	ARIReplaced          bool
}

type Applicant interface {
	Apply(ctx context.Context) (*ApplyResult, error)
}

type ApplicantWithWorkflowNodeConfig struct {
	Node   *domain.WorkflowNode
	Logger *slog.Logger
}

func NewWithWorkflowNode(config ApplicantWithWorkflowNodeConfig) (Applicant, error) {
	if config.Node == nil {
		return nil, fmt.Errorf("node is nil")
	}
	if config.Node.Type != domain.WorkflowNodeTypeApply {
		return nil, fmt.Errorf("node type is not '%s'", string(domain.WorkflowNodeTypeApply))
	}

	nodeCfg := config.Node.GetConfigForApply()
	options := &applicantProviderOptions{
		Domains:                 xslices.Filter(strings.Split(nodeCfg.Domains, ";"), func(s string) bool { return s != "" }),
		ContactEmail:            nodeCfg.ContactEmail,
		Provider:                domain.ACMEDns01ProviderType(nodeCfg.Provider),
		ProviderAccessConfig:    make(map[string]any),
		ProviderServiceConfig:   nodeCfg.ProviderConfig,
		CAProvider:              domain.CAProviderType(nodeCfg.CAProvider),
		CAProviderAccessConfig:  make(map[string]any),
		CAProviderServiceConfig: nodeCfg.CAProviderConfig,
		KeyAlgorithm:            nodeCfg.KeyAlgorithm,
		ACMEProfile:             nodeCfg.ACMEProfile,
		Nameservers:             xslices.Filter(strings.Split(nodeCfg.Nameservers, ";"), func(s string) bool { return s != "" }),
		DnsPropagationWait:      nodeCfg.DnsPropagationWait,
		DnsPropagationTimeout:   nodeCfg.DnsPropagationTimeout,
		DnsTTL:                  nodeCfg.DnsTTL,
		DisableFollowCNAME:      nodeCfg.DisableFollowCNAME,
	}

	accessRepo := repository.NewAccessRepository()
	if nodeCfg.ProviderAccessId != "" {
		if access, err := accessRepo.GetById(context.Background(), nodeCfg.ProviderAccessId); err != nil {
			return nil, fmt.Errorf("failed to get access #%s record: %w", nodeCfg.ProviderAccessId, err)
		} else {
			options.ProviderAccessConfig = access.Config
		}
	}
	if nodeCfg.CAProviderAccessId != "" {
		if access, err := accessRepo.GetById(context.Background(), nodeCfg.CAProviderAccessId); err != nil {
			return nil, fmt.Errorf("failed to get access #%s record: %w", nodeCfg.CAProviderAccessId, err)
		} else {
			options.CAProviderAccessId = access.Id
			options.CAProviderAccessConfig = access.Config
		}
	}

	settingsRepo := repository.NewSettingsRepository()
	if string(options.CAProvider) == "" {
		settings, _ := settingsRepo.GetByName(context.Background(), "sslProvider")

		sslProviderConfig := &acmeSSLProviderConfig{
			Config:   make(map[domain.CAProviderType]map[string]any),
			Provider: caDefault,
		}
		if settings != nil {
			if err := json.Unmarshal([]byte(settings.Content), sslProviderConfig); err != nil {
				return nil, err
			} else if sslProviderConfig.Provider == "" {
				sslProviderConfig.Provider = caDefault
			}
		}

		options.CAProvider = domain.CAProviderType(sslProviderConfig.Provider)
		options.CAProviderAccessConfig = sslProviderConfig.Config[options.CAProvider]
	}

	certRepo := repository.NewCertificateRepository()
	lastCertificate, _ := certRepo.GetByWorkflowNodeId(context.Background(), config.Node.Id)
	if lastCertificate != nil && !lastCertificate.ACMERenewed {
		newCertSan := slices.Clone(options.Domains)
		oldCertSan := strings.Split(lastCertificate.SubjectAltNames, ";")
		slices.Sort(newCertSan)
		slices.Sort(oldCertSan)

		if slices.Equal(newCertSan, oldCertSan) {
			lastCertX509, _ := certcrypto.ParsePEMCertificate([]byte(lastCertificate.Certificate))
			if lastCertX509 != nil {
				replacedARICertId, _ := certificate.MakeARICertID(lastCertX509)
				options.ARIReplaceAcct = lastCertificate.ACMEAccountUrl
				options.ARIReplaceCert = replacedARICertId
			}
		}
	}

	applicant, err := createApplicantProvider(options)
	if err != nil {
		return nil, err
	}

	return &applicantImpl{
		applicant: applicant,
		options:   options,
	}, nil
}

type applicantImpl struct {
	applicant challenge.Provider
	options   *applicantProviderOptions
}

var _ Applicant = (*applicantImpl)(nil)

func (d *applicantImpl) Apply(ctx context.Context) (*ApplyResult, error) {
	limiter := getLimiter(fmt.Sprintf("apply_%s", d.options.ContactEmail))
	if err := limiter.Wait(ctx); err != nil {
		return nil, err
	}

	return applyUseLego(d.applicant, d.options)
}

const (
	limitBurst         = 300
	limitRate  float64 = float64(1) / float64(36)
)

var limiters sync.Map

func getLimiter(key string) *rate.Limiter {
	limiter, _ := limiters.LoadOrStore(key, rate.NewLimiter(rate.Limit(limitRate), 300))
	return limiter.(*rate.Limiter)
}

func applyUseLego(legoProvider challenge.Provider, options *applicantProviderOptions) (*ApplyResult, error) {
	user, err := newAcmeUser(string(options.CAProvider), options.CAProviderAccessId, options.ContactEmail)
	if err != nil {
		return nil, err
	}

	// Some unified lego environment variables are configured here.
	// link: https://github.com/go-acme/lego/issues/1867
	os.Setenv("LEGO_DISABLE_CNAME_SUPPORT", strconv.FormatBool(options.DisableFollowCNAME))

	// Create an ACME client config
	config := lego.NewConfig(user)
	config.Certificate.KeyType = parseLegoKeyAlgorithm(domain.CertificateKeyAlgorithmType(options.KeyAlgorithm))
	switch user.getCAProvider() {
	case caSSLCom:
		if strings.HasPrefix(options.KeyAlgorithm, "RSA") {
			config.CADirURL = caDirUrls[caSSLCom+"RSA"]
		} else if strings.HasPrefix(options.KeyAlgorithm, "EC") {
			config.CADirURL = caDirUrls[caSSLCom+"ECC"]
		} else {
			config.CADirURL = caDirUrls[caSSLCom]
		}

	case caCustom:
		caDirURL := xmaps.GetString(options.CAProviderAccessConfig, "endpoint")
		if caDirURL != "" {
			config.CADirURL = caDirURL
		} else {
			return nil, fmt.Errorf("invalid ca provider endpoint")
		}

	default:
		config.CADirURL = caDirUrls[user.CA]
	}

	// Create an ACME client
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	// Set the DNS01 challenge provider
	client.Challenge.SetDNS01Provider(legoProvider,
		dns01.CondOption(
			len(options.Nameservers) > 0,
			dns01.AddRecursiveNameservers(dns01.ParseNameservers(options.Nameservers)),
		),
		dns01.CondOption(
			options.DnsPropagationWait > 0,
			dns01.PropagationWait(time.Duration(options.DnsPropagationWait)*time.Second, true),
		),
		dns01.CondOption(
			len(options.Nameservers) > 0 || options.DnsPropagationWait > 0,
			dns01.DisableAuthoritativeNssPropagationRequirement(),
		),
	)

	// New users need to register first
	if !user.hasRegistration() {
		reg, err := registerAcmeUserWithSingleFlight(client, user, options.CAProviderAccessConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to register acme user: %w", err)
		}
		user.Registration = reg
	}

	// Obtain a certificate
	certRequest := certificate.ObtainRequest{
		Domains: options.Domains,
		Bundle:  true,
		Profile: options.ACMEProfile,
	}
	if options.ARIReplaceAcct == user.Registration.URI {
		certRequest.ReplacesCertID = options.ARIReplaceCert
	}

	certResource, err := client.Certificate.Obtain(certRequest)
	if err != nil {
		return nil, err
	}

	return &ApplyResult{
		CSR:                  strings.TrimSpace(string(certResource.CSR)),
		FullChainCertificate: strings.TrimSpace(string(certResource.Certificate)),
		IssuerCertificate:    strings.TrimSpace(string(certResource.IssuerCertificate)),
		PrivateKey:           strings.TrimSpace(string(certResource.PrivateKey)),
		ACMEAccountUrl:       user.Registration.URI,
		ACMECertUrl:          certResource.CertURL,
		ACMECertStableUrl:    certResource.CertStableURL,
		ARIReplaced:          certRequest.ReplacesCertID != "",
	}, nil
}

func parseLegoKeyAlgorithm(algo domain.CertificateKeyAlgorithmType) certcrypto.KeyType {
	alogMap := map[domain.CertificateKeyAlgorithmType]certcrypto.KeyType{
		domain.CertificateKeyAlgorithmTypeRSA2048: certcrypto.RSA2048,
		domain.CertificateKeyAlgorithmTypeRSA3072: certcrypto.RSA3072,
		domain.CertificateKeyAlgorithmTypeRSA4096: certcrypto.RSA4096,
		domain.CertificateKeyAlgorithmTypeRSA8192: certcrypto.RSA8192,
		domain.CertificateKeyAlgorithmTypeEC256:   certcrypto.EC256,
		domain.CertificateKeyAlgorithmTypeEC384:   certcrypto.EC384,
		domain.CertificateKeyAlgorithmTypeEC512:   certcrypto.KeyType("P512"),
	}

	if keyType, ok := alogMap[algo]; ok {
		return keyType
	}

	return certcrypto.RSA2048
}
