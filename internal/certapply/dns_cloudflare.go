package certapply

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/certimate-go/certimate/internal/domain"
)

type dnsCloudflare struct {
	client    *http.Client
	token     string
	zoneToken string
	ttl       int
}

type dnsCFObject struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Content *string `json:"content"`
}

type dnsCFEnvelope struct {
	Success    bool            `json:"success"`
	Result     json.RawMessage `json:"result"`
	ResultInfo *struct {
		Page       int `json:"page"`
		TotalPages int `json:"total_pages"`
		TotalCount int `json:"total_count"`
	} `json:"result_info"`
}

func newDNSCloudflare(config DNSProviderConfig, transport http.RoundTripper) (DNSProvider, error) {
	var access domain.AccessConfigForCloudflare
	if err := dnsDecodeAccess(config.AccessConfig, &access); err != nil {
		return nil, err
	}
	if !dnsCredential(access.ApiToken, false) || !dnsCredential(access.ApiTokenForZone, true) {
		return nil, errors.New("invalid Cloudflare ApiToken or ApiTokenForZone")
	}
	if access.ApiTokenForZone == "" {
		access.ApiTokenForZone = access.ApiToken
	}
	return &dnsCloudflare{
		client: &http.Client{Transport: transport, Timeout: dnsHTTPTimeout},
		token:  access.ApiToken, zoneToken: access.ApiTokenForZone, ttl: config.TTL,
	}, nil
}

func (p *dnsCloudflare) request(ctx context.Context, method, path, token string, input any) (*dnsCFEnvelope, error) {
	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return nil, errors.New("DNS Cloudflare cannot encode request")
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://api.cloudflare.com/client/v4"+path, body)
	if err != nil {
		return nil, errors.New("DNS Cloudflare invalid request")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, dnsAPIError(ctx, "Cloudflare HTTP request", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("DNS Cloudflare HTTP request failed")
	}
	var result dnsCFEnvelope
	data, err := io.ReadAll(io.LimitReader(resp.Body, dnsMaxResponse+1))
	if err != nil || len(data) > dnsMaxResponse || json.Unmarshal(data, &result) != nil {
		return nil, dnsAPIError(ctx, "Cloudflare response decode", err)
	}
	if !result.Success {
		return nil, errors.New("DNS Cloudflare API request failed")
	}
	return &result, nil
}

// Visit every page, including short pages when result_info says more remain.
func (p *dnsCloudflare) list(ctx context.Context, path, token string, query url.Values, visit func(dnsCFObject)) error {
	seen := 0
	for page := 1; ; page++ {
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", "50")
		resp, err := p.request(ctx, http.MethodGet, path+"?"+query.Encode(), token, nil)
		if err != nil {
			return err
		}
		var items []dnsCFObject
		if json.Unmarshal(resp.Result, &items) != nil || items == nil || resp.ResultInfo == nil {
			return errors.New("DNS Cloudflare invalid paginated result")
		}
		for _, item := range items {
			visit(item)
		}
		seen += len(items)
		info := resp.ResultInfo
		if info.Page != 0 && info.Page != page {
			return errors.New("DNS Cloudflare unexpected result page")
		}
		if page >= info.TotalPages && seen >= info.TotalCount {
			return nil
		}
		if len(items) == 0 {
			return errors.New("DNS Cloudflare pagination is incomplete")
		}
	}
}

func (p *dnsCloudflare) zone(ctx context.Context, name string) (dnsCFObject, error) {
	for candidate := name; candidate != ""; {
		var found dnsCFObject
		err := p.list(ctx, "/zones", p.zoneToken, url.Values{"name": {candidate}}, func(item dnsCFObject) {
			zone := dnsName(item.Name)
			if zone == candidate && dnsWithinZone(name, zone) && item.ID != "" {
				found = item
				found.Name = zone
			}
		})
		if err != nil {
			return dnsCFObject{}, err
		}
		if found.ID != "" {
			return found, nil
		}
		_, candidate, _ = strings.Cut(candidate, ".")
	}
	return dnsCFObject{}, errors.New("DNS Cloudflare hosted zone not found")
}

func (p *dnsCloudflare) Present(ctx context.Context, record DNSRecord) (func(context.Context) error, error) {
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
	path := "/zones/" + url.PathEscape(zone.ID) + "/dns_records"
	exists := false
	err = p.list(ctx, path, p.token, url.Values{"type": {"TXT"}, "name": {record.Name}}, func(item dnsCFObject) {
		if dnsName(item.Name) == record.Name && item.Type == "TXT" && item.Content != nil && *item.Content == record.Value {
			exists = true
		}
	})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, nil
	}
	resp, err := p.request(ctx, http.MethodPost, path, p.token, struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Content string `json:"content"`
		TTL     int    `json:"ttl"`
	}{Type: "TXT", Name: record.Name, Content: record.Value, TTL: p.ttl})
	if err != nil {
		return nil, err
	}
	var created dnsCFObject
	if json.Unmarshal(resp.Result, &created) != nil || created.ID == "" {
		return nil, errors.New("DNS Cloudflare create returned no record ID")
	}
	return dnsCleanup(func(ctx context.Context) error {
		resp, err := p.request(ctx, http.MethodDelete, path+"/"+url.PathEscape(created.ID), p.token, nil)
		if err != nil {
			return err
		}
		var deleted dnsCFObject
		if json.Unmarshal(resp.Result, &deleted) != nil || deleted.ID != created.ID {
			return errors.New("DNS Cloudflare delete returned no matching record ID")
		}
		return nil
	}), nil
}
