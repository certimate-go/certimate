package certacme

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/acme/api"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dnspersist01"
	"github.com/go-acme/lego/v5/lego"
	"golang.org/x/net/idna"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
)

type dnsPersistValue struct {
	issuer     string
	accountURI string
	wildcard   bool
	expires    time.Time
}

type dnsPersistResolver struct {
	core       *api.Core
	dns        *dnspersist01.Client
	accountURI string
	timeout    time.Duration
}

type dnsPersistTransport struct {
	ctx       context.Context
	transport http.RoundTripper
}

type dnsPersistResponseBody struct {
	io.ReadCloser
	stop   func() bool
	cancel context.CancelFunc
}

func (b *dnsPersistResponseBody) Close() error {
	b.stop()
	b.cancel()
	return b.ReadCloser.Close()
}

func (t *dnsPersistTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancel(req.Context())
	stop := context.AfterFunc(t.ctx, cancel)
	if err := t.ctx.Err(); err != nil {
		stop()
		cancel()
		return nil, err
	}
	response, err := t.transport.RoundTrip(req.Clone(ctx))
	if err != nil {
		stop()
		cancel()
		return nil, err
	}
	response.Body = &dnsPersistResponseBody{ReadCloser: response.Body, stop: stop, cancel: cancel}
	return response, nil
}

func NormalizeDNSPersistDomains(domains []string) ([]string, error) {
	if len(domains) == 0 || len(domains) > 100 {
		return nil, errors.New("DNS-PERSIST-01 requires between 1 and 100 domain names")
	}
	result := make([]string, 0, len(domains))
	for _, value := range domains {
		value = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
		wildcard := strings.HasPrefix(value, "*.")
		name, err := idna.Lookup.ToASCII(strings.TrimPrefix(value, "*."))
		if err != nil || !validDNSPersistName(name) {
			return nil, fmt.Errorf("invalid DNS-PERSIST-01 domain %q", value)
		}
		if wildcard {
			name = "*." + name
		}
		if !slices.Contains(result, name) {
			result = append(result, name)
		}
	}
	return result, nil
}

func validDNSPersistName(name string) bool {
	if len(name) > 253 || !strings.Contains(name, ".") || net.ParseIP(name) != nil {
		return false
	}
	for _, label := range strings.Split(name, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, ch := range label {
			if !(ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '-') {
				return false
			}
		}
	}
	return true
}

func ValidateDNSPersistNameservers(nameservers []string) error {
	if len(nameservers) > 8 {
		return errors.New("at most 8 DNS resolvers are allowed")
	}
	for _, nameserver := range nameservers {
		host := nameserver
		if parsedHost, port, err := net.SplitHostPort(nameserver); err == nil {
			host = parsedHost
			n, err := strconv.Atoi(port)
			if err != nil || n < 1 || n > 65535 {
				return fmt.Errorf("invalid DNS resolver port %q", port)
			}
		}
		if net.ParseIP(host) == nil && !validDNSPersistName(strings.ToLower(host)) {
			return fmt.Errorf("invalid DNS resolver %q", nameserver)
		}
	}
	return nil
}

func newDNSPersistCore(ctx context.Context, config *lego.Config, account *ACMEAccount) (*api.Core, error) {
	client := *config.HTTPClient
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	client.Transport = &dnsPersistTransport{ctx: ctx, transport: transport}
	return api.New(&client, config.UserAgent, config.CADirURL, account.ACMEAccountUrl, account.GetPrivateKey())
}

func PrepareDNSPersist(ctx context.Context, account *ACMEAccount, domains []string, profile string) ([]domain.DNSPersistRecord, error) {
	domains, err := NormalizeDNSPersistDomains(domains)
	if err != nil {
		return nil, err
	}
	config := lego.NewConfig(account)
	config.UserAgent = app.AppUserAgent
	config.CADirURL = account.ACMEDirectoryUrl
	core, err := newDNSPersistCore(ctx, config, account)
	if err != nil {
		return nil, err
	}
	order, err := core.Orders.New(ctx, domains, &api.OrderOptions{Profile: profile})
	if err != nil {
		return nil, err
	}
	if len(order.Authorizations) == 0 {
		return nil, errors.New("the CA returned no authorizations for DNS-PERSIST-01 preparation")
	}
	var records []domain.DNSPersistRecord
	for _, authURL := range order.Authorizations {
		auth, err := core.Authorizations.Get(ctx, authURL)
		if err != nil {
			return nil, err
		}
		ch, err := challenge.FindChallenge(challenge.DNSPersist01, auth)
		if err != nil && auth.Status == acme.StatusValid {
			issuer := ""
			switch account.ACMEDirectoryUrl {
			case caDirUrls[domain.CAProviderTypeGoogleTrustServices.String()]:
				issuer = "pki.goog"
			case caDirUrls[domain.CAProviderTypeLiteSSL.String()]:
				issuer = "trustasia.com"
			}
			if issuer != "" {
				ch = acme.Challenge{Type: domain.ACMEChallengeDNSPersist01, AccountURI: account.ACMEAccountUrl, IssuerDomainNames: []string{issuer}}
				err = nil
			}
		}
		if err != nil {
			return nil, fmt.Errorf("the CA did not offer DNS-PERSIST-01 for %s; cached authorizations may need to expire before preparing this domain: %w", auth.Identifier.Value, err)
		}
		record, err := dnsPersistRecord(auth, ch, account.ACMEAccountUrl)
		if err != nil {
			return nil, err
		}
		merged := false
		for i, existing := range records {
			previous, _ := parseDNSPersistValue(existing.Value)
			current, _ := parseDNSPersistValue(record.Value)
			if existing.Name == record.Name && previous.issuer == current.issuer && previous.accountURI == current.accountURI {
				if current.wildcard {
					records[i] = record
				}
				merged = true
				break
			}
		}
		if !merged {
			records = append(records, record)
		}
	}
	return records, nil
}

func dnsPersistRecord(auth acme.Authorization, ch acme.Challenge, accountURI string) (domain.DNSPersistRecord, error) {
	if auth.Identifier.Type != "dns" || len(ch.IssuerDomainNames) == 0 {
		return domain.DNSPersistRecord{}, errors.New("unsupported DNS-PERSIST-01 challenge format; this experimental implementation requires draft-01 issuer-domain-names")
	}
	if ch.AccountURI != "" {
		accountURI = ch.AccountURI
	}
	issuers := slices.Clone(ch.IssuerDomainNames)
	for i := range issuers {
		issuers[i] = strings.ToLower(issuers[i])
	}
	slices.Sort(issuers)
	info, err := dnspersist01.GetChallengeInfo(auth, issuers[0], accountURI, time.Time{})
	if err != nil {
		return domain.DNSPersistRecord{}, err
	}
	if _, err = parseDNSPersistValue(info.Value); err != nil {
		return domain.DNSPersistRecord{}, err
	}
	return domain.DNSPersistRecord{Domain: challenge.GetTargetedDomain(auth), Name: info.FQDN, Type: "TXT", Value: info.Value}, nil
}

func parseDNSPersistValue(value string) (dnsPersistValue, error) {
	parts := strings.Split(value, ";")
	parsed := dnsPersistValue{issuer: strings.ToLower(strings.Trim(parts[0], " \t"))}
	if !validDNSPersistName(parsed.issuer) {
		return parsed, errors.New("invalid persistent TXT issuer")
	}
	seen := make(map[string]bool)
	for _, part := range parts[1:] {
		tag, data, found := strings.Cut(strings.Trim(part, " \t"), "=")
		tag = strings.ToLower(strings.Trim(tag, " \t"))
		data = strings.Trim(data, " \t")
		if !found || tag == "" || seen[tag] {
			return parsed, errors.New("malformed or duplicate persistent TXT parameter")
		}
		seen[tag] = true
		for _, ch := range tag {
			if !(ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9') {
				return parsed, errors.New("invalid persistent TXT parameter name")
			}
		}
		for _, ch := range data {
			if ch < 33 || ch > 126 {
				return parsed, errors.New("invalid persistent TXT parameter value")
			}
		}
		switch tag {
		case "accounturi":
			uri, err := url.Parse(data)
			if err != nil || !uri.IsAbs() || data == "" {
				return parsed, errors.New("invalid persistent TXT account URI")
			}
			parsed.accountURI = data
		case "policy":
			parsed.wildcard = strings.EqualFold(data, "wildcard")
		case "persistuntil":
			seconds, err := strconv.ParseInt(data, 10, 64)
			if err != nil || seconds < 0 {
				return parsed, errors.New("invalid persistent TXT expiration")
			}
			parsed.expires = time.Unix(seconds, 0)
		}
	}
	if parsed.accountURI == "" {
		return parsed, errors.New("persistent TXT account URI is missing")
	}
	return parsed, nil
}

func matchesDNSPersistValue(value string, expected dnsPersistValue, now time.Time, issuers, accountURIs []string) bool {
	actual, err := parseDNSPersistValue(value)
	issuerMatches := actual.issuer == expected.issuer || slices.ContainsFunc(issuers, func(issuer string) bool { return strings.EqualFold(issuer, actual.issuer) })
	accountMatches := actual.accountURI == expected.accountURI || slices.Contains(accountURIs, actual.accountURI)
	return err == nil && issuerMatches && accountMatches &&
		(!expected.wildcard || actual.wildcard) && (actual.expires.IsZero() || actual.expires.After(now))
}

func CheckDNSPersistRecords(ctx context.Context, records []domain.DNSPersistRecord, nameservers []string) ([]domain.DNSPersistRecordStatus, error) {
	if len(records) == 0 || len(records) > 100 {
		return nil, errors.New("between 1 and 100 persistent TXT records are required")
	}
	if err := ValidateDNSPersistNameservers(nameservers); err != nil {
		return nil, err
	}
	for _, record := range records {
		domains, err := NormalizeDNSPersistDomains([]string{record.Domain})
		if err != nil || record.Type != "TXT" || len(record.Value) > 4096 ||
			!strings.EqualFold(record.Name, "_validation-persist."+strings.TrimPrefix(domains[0], "*.")+".") {
			return nil, errors.New("invalid persistent TXT record")
		}
		if _, err := parseDNSPersistValue(record.Value); err != nil {
			return nil, err
		}
	}
	client := dnspersist01.NewClient(&dnspersist01.Options{RecursiveNameservers: nameservers, Timeout: 5 * time.Second})
	result := make([]domain.DNSPersistRecordStatus, 0, len(records))
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result = append(result, checkDNSPersistRecord(ctx, client, record, nil, nil))
	}
	return result, nil
}

func checkDNSPersistRecord(ctx context.Context, client *dnspersist01.Client, record domain.DNSPersistRecord, issuers, accountURIs []string) domain.DNSPersistRecordStatus {
	status := domain.DNSPersistRecordStatus{DNSPersistRecord: record, Status: "missing"}
	expected, err := parseDNSPersistValue(record.Value)
	if err != nil {
		status.Status, status.Message = "mismatch", err.Error()
		return status
	}
	result, err := client.LookupTXT(ctx, record.Name)
	if err != nil {
		status.Status, status.Message = "lookup_failed", err.Error()
		return status
	}
	if len(result.Records) > 0 {
		status.Status = "mismatch"
	}
	for _, txt := range result.Records {
		if matchesDNSPersistValue(txt.Value, expected, time.Now(), issuers, accountURIs) {
			status.Status = "ready"
			return status
		}
	}
	return status
}

func (c *ACMEClient) newDNSPersistCertifier(ctx context.Context, request *ObtainCertificateRequest) (*certificate.Certifier, error) {
	if _, err := NormalizeDNSPersistDomains(request.DomainOrIPs); err != nil {
		return nil, err
	}
	if err := ValidateDNSPersistNameservers(request.Nameservers); err != nil {
		return nil, err
	}
	core, err := newDNSPersistCore(ctx, c.config, c.account)
	if err != nil {
		return nil, err
	}
	timeout := c.config.Certificate.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	resolver := &dnsPersistResolver{
		core: core, accountURI: c.account.ACMEAccountUrl, timeout: timeout,
		dns: dnspersist01.NewClient(&dnspersist01.Options{RecursiveNameservers: request.Nameservers, Timeout: 5 * time.Second}),
	}
	return certificate.NewCertifier(core, resolver, certificate.CertifierOptions{
		Timeout: c.config.Certificate.Timeout, OverallRequestLimit: c.config.Certificate.OverallRequestLimit,
	}), nil
}

func (r *dnsPersistResolver) Solve(ctx context.Context, authorizations []acme.Authorization) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	for _, auth := range authorizations {
		if auth.Status == acme.StatusValid {
			continue
		}
		ch, err := challenge.FindChallenge(challenge.DNSPersist01, auth)
		if err != nil {
			return fmt.Errorf("the CA does not offer DNS-PERSIST-01 for %s: %w", auth.Identifier.Value, err)
		}
		record, err := dnsPersistRecord(auth, ch, r.accountURI)
		if err != nil {
			return err
		}
		status := checkDNSPersistRecord(ctx, r.dns, record, ch.IssuerDomainNames, []string{r.accountURI})
		if status.Status != "ready" {
			return fmt.Errorf("DNS-PERSIST-01 TXT record is %s for %s; publish %s IN TXT %q and retry. %s", status.Status, record.Domain, record.Name, record.Value, status.Message)
		}
		if err := r.validate(ctx, ch); err != nil {
			return fmt.Errorf("DNS-PERSIST-01 validation failed for %s: %w", record.Domain, err)
		}
	}
	return nil
}

func (r *dnsPersistResolver) validate(ctx context.Context, ch acme.Challenge) error {
	response, err := r.core.Challenges.New(ctx, ch.URL)
	for err == nil {
		switch response.Status {
		case acme.StatusValid:
			return nil
		case acme.StatusPending, acme.StatusProcessing:
		default:
			if response.Error != nil {
				return response.Error
			}
			return fmt.Errorf("challenge ended with status %q", response.Status)
		}
		wait := response.RetryAfter
		if wait <= 0 {
			wait = time.Second
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		response, err = r.core.Challenges.Get(ctx, ch.URL)
	}
	return err
}
