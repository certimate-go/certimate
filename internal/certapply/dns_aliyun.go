package certapply

import (
	"context"
	"errors"
	"net/http"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"
	alidns "github.com/go-acme/alidns-20150109/v5/client"

	"github.com/certimate-go/certimate/internal/domain"
)

type dnsAliyun struct {
	client        *alidns.Client
	resourceGroup string
	ttl           int64
}

type dnsAliyunHTTP struct{ client *http.Client }

func (h *dnsAliyunHTTP) Call(request *http.Request, _ *http.Transport) (*http.Response, error) {
	return h.client.Do(request)
}

func newDNSAliyun(config DNSProviderConfig, transport http.RoundTripper) (DNSProvider, error) {
	var access domain.AccessConfigForAliyun
	if err := dnsDecodeAccess(config.AccessConfig, &access); err != nil {
		return nil, err
	}
	if !dnsCredential(access.AccessKeyId, false) || !dnsCredential(access.AccessKeySecret, false) || !dnsCredential(access.ResourceGroupId, true) {
		return nil, errors.New("invalid Aliyun AccessKeyId, AccessKeySecret or ResourceGroupId")
	}
	client, err := alidns.NewClient(&openapi.Config{
		AccessKeyId:     dara.String(access.AccessKeyId),
		AccessKeySecret: dara.String(access.AccessKeySecret),
		Endpoint:        dara.String("alidns.aliyuncs.com"),
		Protocol:        dara.String("HTTPS"),
		ReadTimeout:     dara.Int(int(dnsHTTPTimeout.Milliseconds())),
		ConnectTimeout:  dara.Int(int(dnsHTTPTimeout.Milliseconds())),
		HttpClient:      &dnsAliyunHTTP{client: &http.Client{Transport: transport, Timeout: dnsHTTPTimeout}},
	})
	if err != nil {
		return nil, errors.New("cannot initialize Aliyun DNS client")
	}
	return &dnsAliyun{client: client, resourceGroup: access.ResourceGroupId, ttl: int64(config.TTL)}, nil
}

func dnsAliyunRuntime() *dara.RuntimeOptions {
	// A lost create response may already have created a record without returning its cleanup ID.
	return &dara.RuntimeOptions{Autoretry: dara.Bool(false)}
}

func (p *dnsAliyun) zone(ctx context.Context, name string) (string, error) {
	best := ""
	seen := int64(0)
	for page := int64(1); ; page++ {
		req := &alidns.DescribeDomainsRequest{PageNumber: dara.Int64(page), PageSize: dara.Int64(dnsPageSize)}
		if p.resourceGroup != "" {
			req.ResourceGroupId = dara.String(p.resourceGroup)
		}
		resp, err := alidns.DescribeDomainsWithContext(ctx, p.client, req, dnsAliyunRuntime())
		if err != nil {
			return "", dnsAPIError(ctx, "Aliyun DescribeDomains", err)
		}
		if resp == nil || resp.Body == nil || resp.Body.Domains == nil {
			return "", errors.New("DNS Aliyun DescribeDomains returned no result")
		}
		result := resp.Body
		for _, item := range result.Domains.Domain {
			if item == nil {
				continue
			}
			zone := dnsName(dnsString(item.DomainName))
			if dnsWithinZone(name, zone) && len(zone) > len(best) {
				best = zone
			}
		}
		n := len(result.Domains.Domain)
		seen += int64(n)
		if result.TotalCount != nil && seen >= *result.TotalCount {
			break
		}
		if n == 0 {
			if result.TotalCount != nil {
				return "", errors.New("DNS Aliyun domain pagination is incomplete")
			}
			break
		}
	}
	if best == "" {
		return "", dnsHostedZoneNotFound("Aliyun", name)
	}
	return best, nil
}

func (p *dnsAliyun) checkDomain(ctx context.Context, name string) error {
	_, err := p.zone(ctx, dnsName(name))
	return err
}

func (p *dnsAliyun) Present(ctx context.Context, record DNSRecord) (func(context.Context) error, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	record, err := dnsValidateRecord(record)
	if err != nil {
		return nil, err
	}
	zone, err := p.zone(ctx, record.Name)
	if err != nil {
		return nil, err
	}
	rr := dnsRelativeName(record.Name, zone)
	seen := int64(0)
	for page := int64(1); ; page++ {
		// API keyword matching can be fuzzy, so check record identity locally.
		req := &alidns.DescribeDomainRecordsRequest{
			DomainName: dara.String(zone), PageNumber: dara.Int64(page), PageSize: dara.Int64(dnsPageSize),
		}
		resp, err := alidns.DescribeDomainRecordsWithContext(ctx, p.client, req, dnsAliyunRuntime())
		if err != nil {
			return nil, dnsAPIError(ctx, "Aliyun DescribeDomainRecords", err)
		}
		if resp == nil || resp.Body == nil || resp.Body.DomainRecords == nil {
			return nil, errors.New("DNS Aliyun DescribeDomainRecords returned no result")
		}
		result := resp.Body
		for _, item := range result.DomainRecords.Record {
			if item != nil && dnsName(dnsString(item.RR)) == rr && dnsString(item.Type) == "TXT" && item.Value != nil && *item.Value == record.Value {
				return nil, nil
			}
		}
		n := len(result.DomainRecords.Record)
		seen += int64(n)
		if result.TotalCount != nil && seen >= *result.TotalCount {
			break
		}
		if n == 0 {
			if result.TotalCount != nil {
				return nil, errors.New("DNS Aliyun record pagination is incomplete")
			}
			break
		}
	}
	resp, err := alidns.AddDomainRecordWithContext(ctx, p.client, &alidns.AddDomainRecordRequest{
		DomainName: dara.String(zone), RR: dara.String(rr), Type: dara.String("TXT"), Value: dara.String(record.Value), TTL: dara.Int64(p.ttl),
	}, dnsAliyunRuntime())
	if err != nil {
		return nil, dnsAPIError(ctx, "Aliyun AddDomainRecord", err)
	}
	if resp == nil || resp.Body == nil || dnsString(resp.Body.RecordId) == "" {
		return nil, errors.New("DNS Aliyun AddDomainRecord returned no record ID")
	}
	id := *resp.Body.RecordId
	return dnsCleanup(func(ctx context.Context) error {
		resp, err := alidns.DeleteDomainRecordWithContext(ctx, p.client, &alidns.DeleteDomainRecordRequest{RecordId: dara.String(id)}, dnsAliyunRuntime())
		if err != nil {
			return dnsAPIError(ctx, "Aliyun DeleteDomainRecord", err)
		}
		if resp == nil || resp.Body == nil || dnsString(resp.Body.RecordId) != id {
			return errors.New("DNS Aliyun DeleteDomainRecord returned no matching record ID")
		}
		return nil
	}), nil
}
