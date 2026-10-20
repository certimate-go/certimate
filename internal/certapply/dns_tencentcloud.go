package certapply

import (
	"context"
	"errors"
	"net/http"

	dnspod "github.com/go-acme/tencentclouddnspod/v20210323"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/certimate-go/certimate/internal/domain"
)

type dnsTencent struct {
	client *dnspod.Client
	ttl    uint64
}

func newDNSTencent(config DNSProviderConfig, transport http.RoundTripper) (DNSProvider, error) {
	var access domain.AccessConfigForTencentCloud
	if err := dnsDecodeAccess(config.AccessConfig, &access); err != nil {
		return nil, err
	}
	if !dnsCredential(access.SecretId, false) || !dnsCredential(access.SecretKey, false) {
		return nil, errors.New("invalid Tencent Cloud SecretId or SecretKey")
	}
	p := profile.NewClientProfile()
	p.HttpProfile.ReqTimeout = int(dnsHTTPTimeout.Seconds())
	p.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	client, err := dnspod.NewClient(common.NewCredential(access.SecretId, access.SecretKey), "", p)
	if err != nil {
		return nil, errors.New("cannot initialize Tencent Cloud DNS client")
	}
	client.WithHttpTransport(transport)
	return &dnsTencent{client: client, ttl: uint64(config.TTL)}, nil
}

func (p *dnsTencent) zone(ctx context.Context, name string) (string, error) {
	best := ""
	for offset := int64(0); ; {
		req := dnspod.NewDescribeDomainListRequest()
		req.Type = common.StringPtr("ALL")
		req.Offset = common.Int64Ptr(offset)
		req.Limit = common.Int64Ptr(dnsPageSize)
		resp, err := dnspod.DescribeDomainListWithContext(ctx, p.client, req)
		if err != nil {
			return "", dnsAPIError(ctx, "Tencent DescribeDomainList", err)
		}
		if resp == nil || resp.Response == nil {
			return "", errors.New("DNS Tencent DescribeDomainList returned no result")
		}
		result := resp.Response
		for _, item := range result.DomainList {
			if item == nil {
				continue
			}
			zone := dnsName(dnsString(item.Name))
			if dnsWithinZone(name, zone) && len(zone) > len(best) {
				best = zone
			}
		}
		offset += int64(len(result.DomainList))
		if result.DomainCountInfo != nil && result.DomainCountInfo.DomainTotal != nil {
			if uint64(offset) >= *result.DomainCountInfo.DomainTotal {
				break
			}
			if len(result.DomainList) == 0 {
				return "", errors.New("DNS Tencent domain pagination is incomplete")
			}
		} else if len(result.DomainList) == 0 {
			break
		}
	}
	if best == "" {
		return "", errors.New("DNS Tencent hosted zone not found")
	}
	return best, nil
}

func (p *dnsTencent) Present(ctx context.Context, record DNSRecord) (func(context.Context) error, error) {
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
	for offset := uint64(0); ; {
		req := dnspod.NewDescribeRecordListRequest()
		req.Domain = common.StringPtr(zone)
		req.Subdomain = common.StringPtr(rr)
		req.RecordType = common.StringPtr("TXT")
		req.Offset = common.Uint64Ptr(offset)
		req.Limit = common.Uint64Ptr(dnsPageSize)
		resp, err := dnspod.DescribeRecordListWithContext(ctx, p.client, req)
		if err != nil {
			var sdkErr *tcerr.TencentCloudSDKError
			if errors.As(err, &sdkErr) && sdkErr.GetCode() == "ResourceNotFound.NoDataOfRecord" {
				break
			}
			return nil, dnsAPIError(ctx, "Tencent DescribeRecordList", err)
		}
		if resp == nil || resp.Response == nil {
			return nil, errors.New("DNS Tencent DescribeRecordList returned no result")
		}
		result := resp.Response
		for _, item := range result.RecordList {
			if item != nil && dnsName(dnsString(item.Name)) == rr && dnsString(item.Type) == "TXT" && item.Value != nil && *item.Value == record.Value {
				return nil, nil
			}
		}
		offset += uint64(len(result.RecordList))
		// TotalCount may include records outside the Subdomain filter.
		if len(result.RecordList) == 0 || result.RecordCountInfo != nil && result.RecordCountInfo.TotalCount != nil && offset >= *result.RecordCountInfo.TotalCount {
			break
		}
	}
	req := dnspod.NewCreateRecordRequest()
	req.Domain = common.StringPtr(zone)
	req.SubDomain = common.StringPtr(rr)
	req.RecordType = common.StringPtr("TXT")
	req.RecordLine = common.StringPtr("默认")
	req.Value = common.StringPtr(record.Value)
	req.TTL = common.Uint64Ptr(p.ttl)
	resp, err := dnspod.CreateRecordWithContext(ctx, p.client, req)
	if err != nil {
		return nil, dnsAPIError(ctx, "Tencent CreateRecord", err)
	}
	if resp == nil || resp.Response == nil || resp.Response.RecordId == nil || *resp.Response.RecordId == 0 {
		return nil, errors.New("DNS Tencent CreateRecord returned no record ID")
	}
	id := *resp.Response.RecordId
	return dnsCleanup(func(ctx context.Context) error {
		req := dnspod.NewDeleteRecordRequest()
		req.Domain = common.StringPtr(zone)
		req.RecordId = common.Uint64Ptr(id)
		resp, err := dnspod.DeleteRecordWithContext(ctx, p.client, req)
		if err != nil {
			return dnsAPIError(ctx, "Tencent DeleteRecord", err)
		}
		if resp == nil || resp.Response == nil {
			return errors.New("DNS Tencent DeleteRecord returned no result")
		}
		return nil
	}), nil
}
