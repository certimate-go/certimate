package acmeaccount

import (
	"context"
	"net/mail"
	"time"

	"github.com/certimate-go/certimate/internal/certacme"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/repository"
)

func (s *Service) PrepareDNSPersist(ctx context.Context, req *dtos.DNSPersistPrepareReq) (*dtos.DNSPersistPrepareResp, error) {
	if req == nil {
		return nil, domain.NewError(400, "request body is empty")
	}
	email, err := mail.ParseAddress(req.ContactEmail)
	if err != nil || email.Address != req.ContactEmail || len(req.ContactEmail) > 254 {
		return nil, domain.NewError(400, "a valid contact email is required")
	}
	domains, err := certacme.NormalizeDNSPersistDomains(req.Domains)
	if err != nil {
		return nil, domain.NewError(400, err.Error())
	}
	if err := certacme.ValidateDNSPersistNameservers(req.Nameservers); err != nil {
		return nil, domain.NewError(400, err.Error())
	}
	if len(req.ACMEProfile) > 128 {
		return nil, domain.NewError(400, "the ACME profile is too long")
	}
	algorithm := domain.CertificateKeyAlgorithmType(req.KeyAlgorithm)
	if algorithm == "" {
		algorithm = domain.CertificateKeyAlgorithmTypeRSA2048
	}
	switch algorithm {
	case domain.CertificateKeyAlgorithmTypeRSA2048, domain.CertificateKeyAlgorithmTypeRSA3072, domain.CertificateKeyAlgorithmTypeRSA4096,
		domain.CertificateKeyAlgorithmTypeRSA8192, domain.CertificateKeyAlgorithmTypeEC256, domain.CertificateKeyAlgorithmTypeEC384:
	default:
		return nil, domain.NewError(400, "unsupported key algorithm")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	accessConfig := make(map[string]any)
	if req.CAProviderAccessId != "" {
		access, err := repository.NewAccessRepository().GetById(ctx, req.CAProviderAccessId)
		if err != nil {
			return nil, err
		}
		if req.CAProvider == "" || access.Provider != req.CAProvider {
			return nil, domain.NewError(400, "the selected CA credentials do not match the CA")
		}
		accessConfig = access.Config
	}
	config, err := certacme.CreateACMEConfig(ctx, &certacme.ACMEConfigOptions{
		CAProvider: domain.CAProviderType(req.CAProvider), CAProviderAccessConfig: accessConfig,
		CAProviderExtendedConfig: req.CAProviderConfig, CertifierKeyAlgorithm: algorithm,
	})
	if err != nil {
		return nil, err
	}
	account, err := certacme.CreateACMEAccountWithSingleFlight(ctx, config, req.ContactEmail)
	if err != nil {
		return nil, err
	}
	records, err := certacme.PrepareDNSPersist(ctx, account, domains, req.ACMEProfile)
	if err != nil {
		return nil, err
	}
	return &dtos.DNSPersistPrepareResp{AccountId: account.Id, AccountUrl: account.ACMEAccountUrl, Records: records}, nil
}

func (s *Service) CheckDNSPersist(ctx context.Context, req *dtos.DNSPersistCheckReq) (*dtos.DNSPersistCheckResp, error) {
	if req == nil {
		return nil, domain.NewError(400, "request body is empty")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	records, err := certacme.CheckDNSPersistRecords(ctx, req.Records, req.Nameservers)
	if err != nil {
		return nil, domain.NewError(400, err.Error())
	}
	ready := true
	for _, record := range records {
		ready = ready && record.Status == "ready"
	}
	return &dtos.DNSPersistCheckResp{Ready: ready, Records: records}, nil
}
