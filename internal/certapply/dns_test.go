package certapply

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

type dnsTestTransport func(*http.Request) (*http.Response, error)

func (f dnsTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

const dnsTestSecret = "test-secret-NEVER-EXPOSE"

func dnsTestConfig(provider string) DNSProviderConfig {
	config := DNSProviderConfig{Provider: provider}
	switch provider {
	case "tencentcloud":
		config.AccessConfig = map[string]any{"secretId": "test-id", "secretKey": dnsTestSecret}
	case "aliyun":
		config.AccessConfig = map[string]any{"accessKeyId": "test-id", "accessKeySecret": dnsTestSecret, "resourceGroupId": "rg-test"}
	case "cloudflare":
		config.AccessConfig = map[string]any{"apiToken": dnsTestSecret, "apiTokenForZone": "zone-only-token"}
	}
	return config
}

func dnsTestResponse(r *http.Request, status int, value any) *http.Response {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(string(data))), ContentLength: int64(len(data)), Request: r}
}

type dnsTestCloud struct {
	t             *testing.T
	provider      string
	record        DNSRecord
	zone          string
	rr            string
	same          bool
	noData        bool
	fail          string
	missingID     bool
	boundaryOnly  bool
	fallbackToken bool
	cancelAt      string
	cancel        context.CancelFunc
	ttl           int
	requests      int
	zonePages     int
	recordPages   int
	creates       int
	deletes       []string
}

func (f *dnsTestCloud) providerClient() DNSProvider {
	f.t.Helper()
	config := dnsTestConfig(f.provider)
	config.TTL = f.ttl
	if f.fallbackToken {
		delete(config.AccessConfig, "apiTokenForZone")
	}
	p, err := newDNSProvider(config, dnsTestTransport(f.roundTrip))
	if err != nil {
		f.t.Fatal(err)
	}
	return p
}

func (f *dnsTestCloud) check(got, want any) {
	f.t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(want) {
		f.t.Fatalf("unexpected HTTP field: got %v, want %v", got, want)
	}
}

func (f *dnsTestCloud) roundTrip(r *http.Request) (*http.Response, error) {
	f.requests++
	f.check(r.URL.Scheme, "https")
	var action string
	fields := map[string]any{}
	query := r.URL.Query()
	switch f.provider {
	case "tencentcloud":
		f.check(r.URL.Host, "dnspod.tencentcloudapi.com")
		f.check(r.Method, http.MethodPost)
		if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
			f.t.Fatal(err)
		}
		action = r.Header.Get("X-TC-Action")
	case "aliyun":
		f.check(r.URL.Host, "alidns.aliyuncs.com")
		f.check(r.Method, http.MethodPost)
		// dara preserves lower-case header keys rather than canonicalizing.
		for key, values := range r.Header {
			if strings.EqualFold(key, "x-acs-action") {
				action = values[0]
			}
		}
		if r.Body == nil {
			r.Body = http.NoBody
		}
		if err := r.ParseForm(); err != nil {
			f.t.Fatal(err)
		}
		for key, values := range r.Form {
			fields[key] = values[0]
		}
	case "cloudflare":
		f.check(r.URL.Host, "api.cloudflare.com")
		if r.URL.Path == "/client/v4/zones" {
			action = "zones"
			f.check(r.Method, http.MethodGet)
			token := "zone-only-token"
			if f.fallbackToken {
				token = dnsTestSecret
			}
			f.check(r.Header.Get("Authorization"), "Bearer "+token)
		} else {
			f.check(r.Header.Get("Authorization"), "Bearer "+dnsTestSecret)
			switch r.Method {
			case http.MethodGet:
				action = "records"
			case http.MethodPost:
				action = "create"
				if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
					f.t.Fatal(err)
				}
			case http.MethodDelete:
				action = "delete"
			default:
				f.t.Fatalf("unexpected method %s", r.Method)
			}
		}
	}
	stage := map[string]string{
		"DescribeDomainList": "zones", "DescribeDomains": "zones", "zones": "zones",
		"DescribeRecordList": "records", "DescribeDomainRecords": "records", "records": "records",
		"CreateRecord": "create", "AddDomainRecord": "create", "create": "create",
		"DeleteRecord": "delete", "DeleteDomainRecord": "delete", "delete": "delete",
	}[action]
	if stage == "" {
		f.t.Fatalf("unexpected API operation %q", action)
	}
	if f.cancelAt == stage {
		f.cancel()
		<-r.Context().Done()
		return nil, r.Context().Err()
	}
	if stage == f.fail {
		return f.apiError(r, "Forbidden."+dnsTestSecret), nil
	}
	if stage == "records" && f.noData {
		return f.apiError(r, "ResourceNotFound.NoDataOfRecord"), nil
	}
	switch stage {
	case "zones":
		f.zonePages++
		return f.zones(r, fields, query), nil
	case "records":
		f.recordPages++
		return f.records(r, fields, query), nil
	case "create":
		f.creates++
		ttl := f.ttl
		if ttl == 0 {
			ttl = 600
		}
		switch f.provider {
		case "tencentcloud":
			f.check(fields["Domain"], f.zone)
			f.check(fields["SubDomain"], f.rr)
			f.check(fields["RecordType"], "TXT")
			f.check(fields["RecordLine"], "默认")
			f.check(fields["Value"], f.record.Value)
			f.check(fields["TTL"], ttl)
		case "aliyun":
			f.check(fields["DomainName"], f.zone)
			f.check(fields["RR"], f.rr)
			f.check(fields["Type"], "TXT")
			f.check(fields["Value"], f.record.Value)
			f.check(fields["TTL"], ttl)
		case "cloudflare":
			f.check(r.URL.Path, "/client/v4/zones/zone-new/dns_records")
			f.check(fields["name"], dnsName(f.record.Name))
			f.check(fields["type"], "TXT")
			f.check(fields["content"], f.record.Value)
			f.check(fields["ttl"], ttl)
		}
		return f.idResponse(r, f.missingID), nil
	case "delete":
		id := ""
		switch f.provider {
		case "tencentcloud":
			f.check(fields["Domain"], f.zone)
			id = fmt.Sprint(fields["RecordId"])
		case "aliyun":
			id = fmt.Sprint(fields["RecordId"])
		case "cloudflare":
			f.check(r.URL.Path, "/client/v4/zones/zone-new/dns_records/900")
			id = "900"
		}
		f.check(id, "900")
		f.deletes = append(f.deletes, id)
		return f.idResponse(r, false), nil
	}
	panic("unreachable")
}

func (f *dnsTestCloud) apiError(r *http.Request, code string) *http.Response {
	switch f.provider {
	case "tencentcloud":
		return dnsTestResponse(r, 200, map[string]any{"Response": map[string]any{"Error": map[string]any{"Code": code, "Message": dnsTestSecret}}})
	case "aliyun":
		return dnsTestResponse(r, 403, map[string]any{"Code": code, "Message": dnsTestSecret, "RequestId": "test"})
	default:
		// HTTP 200 is not sufficient: Cloudflare success must also be true.
		return dnsTestResponse(r, 200, map[string]any{"success": false, "errors": []any{map[string]any{"code": 10000, "message": dnsTestSecret}}})
	}
}

func (f *dnsTestCloud) idResponse(r *http.Request, missing bool) *http.Response {
	result := map[string]any{}
	if !missing {
		switch f.provider {
		case "tencentcloud":
			result["RecordId"] = 900
		case "aliyun":
			result["RecordId"] = "900"
		case "cloudflare":
			result["id"] = "900"
		}
	}
	if f.provider == "tencentcloud" {
		return dnsTestResponse(r, 200, map[string]any{"Response": result})
	}
	if f.provider == "cloudflare" {
		return dnsTestResponse(r, 200, map[string]any{"success": true, "result": result})
	}
	return dnsTestResponse(r, 200, result)
}

func dnsTestCFPage(r *http.Request, items []any, page, total int) *http.Response {
	return dnsTestResponse(r, 200, map[string]any{"success": true, "result": items, "result_info": map[string]any{"page": page, "total_pages": 2, "total_count": total}})
}

func (f *dnsTestCloud) zones(r *http.Request, fields map[string]any, query url.Values) *http.Response {
	names := []string{"example.com", "ample.com", f.zone}
	if f.boundaryOnly {
		names = []string{"ample.com", "invalidexample.com", "other.com"}
	}
	page := 1
	switch f.provider {
	case "tencentcloud":
		f.check(fields["Limit"], dnsPageSize)
		if fmt.Sprint(fields["Offset"]) == "2" {
			page = 2
		} else {
			f.check(fields["Offset"], 0)
		}
	case "aliyun":
		f.check(fields["PageSize"], dnsPageSize)
		f.check(fields["ResourceGroupId"], "rg-test")
		page, _ = strconv.Atoi(fmt.Sprint(fields["PageNumber"]))
	case "cloudflare":
		page, _ = strconv.Atoi(query.Get("page"))
		candidate := query.Get("name")
		if f.boundaryOnly || candidate != f.zone {
			// A wrong suffix must not be accepted just because the API returned it.
			return dnsTestResponse(r, 200, map[string]any{"success": true, "result": []any{map[string]any{"name": "ample.com", "id": "wrong-zone"}}, "result_info": map[string]any{"page": page, "total_pages": 1, "total_count": 1}})
		}
		if page == 1 {
			return dnsTestCFPage(r, []any{map[string]any{"name": "example.com", "id": "zone-parent"}}, 1, 2)
		}
		f.check(page, 2)
		return dnsTestCFPage(r, []any{map[string]any{"name": f.zone, "id": "zone-new"}}, 2, 2)
	}
	if page == 1 {
		names = names[:2]
	} else {
		f.check(page, 2)
		names = names[2:]
	}
	items := []any{}
	for _, name := range names {
		if f.provider == "tencentcloud" {
			items = append(items, map[string]any{"Name": name, "DomainId": 100})
		} else {
			items = append(items, map[string]any{"DomainName": name})
		}
	}
	if f.provider == "tencentcloud" {
		return dnsTestResponse(r, 200, map[string]any{"Response": map[string]any{"DomainList": items, "DomainCountInfo": map[string]any{"DomainTotal": 3}}})
	}
	return dnsTestResponse(r, 200, map[string]any{"Domains": map[string]any{"Domain": items}, "TotalCount": 3, "PageNumber": page, "PageSize": 2})
}

func (f *dnsTestCloud) records(r *http.Request, fields map[string]any, query url.Values) *http.Response {
	page := 1
	switch f.provider {
	case "tencentcloud":
		f.check(fields["Domain"], f.zone)
		f.check(fields["Subdomain"], f.rr)
		if _, wrong := fields["SubDomain"]; wrong {
			f.t.Fatal("query must use Subdomain, not SubDomain")
		}
		f.check(fields["RecordType"], "TXT")
		if fmt.Sprint(fields["Offset"]) == "3" {
			page = 2
		} else {
			f.check(fields["Offset"], 0)
		}
	case "aliyun":
		f.check(fields["DomainName"], f.zone)
		page, _ = strconv.Atoi(fmt.Sprint(fields["PageNumber"]))
	case "cloudflare":
		f.check(r.URL.Path, "/client/v4/zones/zone-new/dns_records")
		f.check(query.Get("name"), dnsName(f.record.Name))
		f.check(query.Get("type"), "TXT")
		page, _ = strconv.Atoi(query.Get("page"))
	}
	type entry struct{ name, kind, value string }
	names := []entry{{f.rr, "TXT", "other-value"}, {"another", "TXT", f.record.Value}, {f.rr, "A", f.record.Value}}
	if page == 2 {
		names = []entry{{"another", "TXT", "unrelated"}}
		if f.same {
			names = []entry{{f.rr, "TXT", f.record.Value}}
		}
	} else {
		f.check(page, 1)
	}
	items := []any{}
	for i, item := range names {
		switch f.provider {
		case "tencentcloud":
			items = append(items, map[string]any{"RecordId": i + 1, "Name": item.name, "Type": item.kind, "Value": item.value})
		case "aliyun":
			items = append(items, map[string]any{"RecordId": strconv.Itoa(i + 1), "RR": item.name, "Type": item.kind, "Value": item.value})
		case "cloudflare":
			name := item.name + "." + f.zone
			if item.name == "@" {
				name = f.zone
			}
			items = append(items, map[string]any{"id": strconv.Itoa(i + 1), "name": name, "type": item.kind, "content": item.value})
		}
	}
	if f.provider == "tencentcloud" {
		return dnsTestResponse(r, 200, map[string]any{"Response": map[string]any{"RecordList": items, "RecordCountInfo": map[string]any{"TotalCount": 4, "ListCount": len(items)}}})
	}
	if f.provider == "aliyun" {
		return dnsTestResponse(r, 200, map[string]any{"DomainRecords": map[string]any{"Record": items}, "TotalCount": 4, "PageNumber": page})
	}
	return dnsTestCFPage(r, items, page, 4)
}

func dnsTestFixture(t *testing.T, provider string) *dnsTestCloud {
	return &dnsTestCloud{t: t, provider: provider, zone: "sub.example.com", rr: "_dnsauth.www", record: DNSRecord{Name: "_DNSAUTH.WWW.Sub.Example.COM.", Value: " raw=+/% \"quote\"\n原始TXT "}}
}

func TestDNSProvidersRawRecords(t *testing.T) {
	for _, provider := range []string{"tencentcloud", "aliyun", "cloudflare"} {
		for _, same := range []bool{false, true} {
			for _, apex := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/same=%t/apex=%t", provider, same, apex), func(t *testing.T) {
					f := dnsTestFixture(t, provider)
					f.same = same
					if apex {
						f.record.Name, f.rr, f.ttl = f.zone, "@", 1800
					}
					cleanup, err := f.providerClient().Present(context.Background(), f.record)
					if err != nil {
						t.Fatal(err)
					}
					if f.zonePages < 2 || f.recordPages != 2 {
						t.Fatalf("pagination skipped: zones=%d records=%d", f.zonePages, f.recordPages)
					}
					if same {
						if cleanup != nil || f.creates != 0 || len(f.deletes) != 0 {
							t.Fatal("pre-existing exact record must have no cleanup or mutation")
						}
						return
					}
					if cleanup == nil || f.creates != 1 {
						t.Fatal("new record must return cleanup")
					}
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					if !errors.Is(cleanup(ctx), context.Canceled) || len(f.deletes) != 0 {
						t.Fatal("canceled cleanup must not delete")
					}
					for range 2 {
						if err := cleanup(context.Background()); err != nil {
							t.Fatal(err)
						}
					}
					if len(f.deletes) != 1 || f.deletes[0] != "900" {
						t.Fatalf("must delete only newly created ID once: %v", f.deletes)
					}
				})
			}
		}
	}
}

func TestDNSProvidersFailures(t *testing.T) {
	for _, provider := range []string{"tencentcloud", "aliyun", "cloudflare"} {
		for _, stage := range []string{"zones", "records", "create", "delete", "missing-id", "no-zone"} {
			t.Run(provider+"/"+stage, func(t *testing.T) {
				f := dnsTestFixture(t, provider)
				f.fail, f.missingID, f.boundaryOnly = stage, stage == "missing-id", stage == "no-zone"
				cleanup, err := f.providerClient().Present(context.Background(), f.record)
				if stage == "delete" {
					if err != nil || cleanup == nil {
						t.Fatalf("present: %v", err)
					}
					err = cleanup(context.Background())
					if err == nil {
						t.Fatal("expected delete permission error")
					}
					f.fail = ""
					if retryErr := cleanup(context.Background()); retryErr != nil {
						t.Fatal(retryErr)
					}
				} else if cleanup != nil {
					t.Fatal("failed present returned cleanup")
				}
				if err == nil || strings.Contains(err.Error(), dnsTestSecret) {
					t.Fatalf("expected sanitized error, got %v", err)
				}
				if (stage == "zones" || stage == "records" || stage == "no-zone") && f.creates != 0 {
					t.Fatal("failed discovery must not create records")
				}
			})
		}
	}
}

func TestDNSTencentNoData(t *testing.T) {
	f := dnsTestFixture(t, "tencentcloud")
	f.noData = true
	cleanup, err := f.providerClient().Present(context.Background(), f.record)
	if err != nil || cleanup == nil || f.creates != 1 {
		t.Fatalf("NoDataOfRecord must be an empty list: %v", err)
	}
	if err := cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestDNSCloudflareZoneTokenFallback(t *testing.T) {
	f := dnsTestFixture(t, "cloudflare")
	f.fallbackToken, f.same = true, true
	cleanup, err := f.providerClient().Present(context.Background(), f.record)
	if err != nil || cleanup != nil {
		t.Fatalf("optional zone token fallback: %v", err)
	}
}

func TestDNSEmptyTXTValues(t *testing.T) {
	for _, provider := range []string{"tencentcloud", "aliyun", "cloudflare"} {
		for _, same := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/same=%t", provider, same), func(t *testing.T) {
				f := dnsTestFixture(t, provider)
				f.record.Value, f.same = "", same
				cleanup, err := f.providerClient().Present(context.Background(), f.record)
				if err != nil || (cleanup == nil) != same {
					t.Fatalf("empty TXT value must retain normal ownership semantics: %v", err)
				}
				if cleanup != nil {
					if err := cleanup(context.Background()); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}

func TestDNSProvidersDeadline(t *testing.T) {
	for _, provider := range []string{"tencentcloud", "aliyun", "cloudflare"} {
		t.Run(provider, func(t *testing.T) {
			calls := 0
			p, err := newDNSProvider(dnsTestConfig(provider), dnsTestTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				<-r.Context().Done()
				return nil, r.Context().Err()
			}))
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			cleanup, err := p.Present(ctx, DNSRecord{Name: "_dnsauth.example.com", Value: "raw"})
			if cleanup != nil || !errors.Is(err, context.DeadlineExceeded) || calls != 1 {
				t.Fatalf("deadline must propagate without retries: calls=%d error=%v", calls, err)
			}
		})
	}
}

func TestDNSProvidersContext(t *testing.T) {
	for _, provider := range []string{"tencentcloud", "aliyun", "cloudflare"} {
		for _, stage := range []string{"before", "zones", "records", "create", "delete"} {
			t.Run(provider+"/"+stage, func(t *testing.T) {
				f := dnsTestFixture(t, provider)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				f.cancelAt, f.cancel = stage, cancel
				if stage == "before" {
					cancel()
				}
				cleanup, err := f.providerClient().Present(ctx, f.record)
				if stage == "delete" {
					if err != nil || cleanup == nil {
						t.Fatalf("present: %v", err)
					}
					err = cleanup(ctx)
				}
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("expected context.Canceled, got %v", err)
				}
				if stage == "before" && f.requests != 0 {
					t.Fatal("canceled context issued HTTP request")
				}
			})
		}
	}
}

func TestDNSConfigValidation(t *testing.T) {
	transport := dnsTestTransport(func(*http.Request) (*http.Response, error) {
		t.Fatal("constructor must not perform HTTP or credential discovery")
		return nil, errors.New("unexpected request")
	})
	for _, provider := range []string{"tencentcloud", "aliyun", "cloudflare"} {
		for _, invalid := range []string{"nil", "empty", "wrong-type", "whitespace", "control", "optional", "negative-ttl"} {
			t.Run(provider+"/"+invalid, func(t *testing.T) {
				config := dnsTestConfig(provider)
				key := map[string]string{"tencentcloud": "secretKey", "aliyun": "accessKeySecret", "cloudflare": "apiToken"}[provider]
				switch invalid {
				case "nil":
					config.AccessConfig = nil
				case "empty":
					config.AccessConfig[key] = ""
				case "wrong-type":
					config.AccessConfig[key] = 123
				case "whitespace":
					config.AccessConfig[key] = "  "
				case "control":
					config.AccessConfig[key] = "secret\r\nAuthorization: bad"
				case "optional":
					switch provider {
					case "cloudflare":
						config.AccessConfig["apiTokenForZone"] = " "
					case "aliyun":
						config.AccessConfig["resourceGroupId"] = []string{"invalid"}
					default:
						config.AccessConfig["secretId"] = false
					}
				case "negative-ttl":
					config.TTL = -1
				}
				if p, err := newDNSProvider(config, transport); err == nil || p != nil || strings.Contains(err.Error(), dnsTestSecret) {
					t.Fatalf("invalid config accepted or leaked: %v", err)
				}
			})
		}
		if _, err := newDNSProvider(dnsTestConfig(provider), transport); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := newDNSProvider(DNSProviderConfig{Provider: "unsupported"}, transport); err == nil {
		t.Fatal("unsupported provider accepted")
	}
}

func TestDNSInvalidRecordNames(t *testing.T) {
	for _, name := range []string{"", "a..example.com", ".example.com", "a.example.com..", "a/../example.com", " example.com", "-bad.example.com", strings.Repeat("a", 64) + ".com", "中文.example.com"} {
		if _, err := dnsValidateRecord(DNSRecord{Name: name}); err == nil {
			t.Fatalf("accepted invalid name %q", name)
		}
	}
	if dnsWithinZone("badexample.com", "example.com") || !dnsWithinZone("a.example.com", "example.com") {
		t.Fatal("zone matching must use label boundaries")
	}
}

func TestDNSBoundedHTTP(t *testing.T) {
	for _, provider := range []string{"tencentcloud", "aliyun", "cloudflare"} {
		for _, mode := range []string{"oversize", "content-length", "redirect", "malformed", "http-error", "transport-error"} {
			t.Run(provider+"/"+mode, func(t *testing.T) {
				calls := 0
				transport := dnsTestTransport(func(r *http.Request) (*http.Response, error) {
					calls++
					if mode == "transport-error" {
						return nil, errors.New(dnsTestSecret)
					}
					body := dnsTestSecret
					status := 200
					if mode == "oversize" || mode == "content-length" {
						body = strings.Repeat("x", dnsMaxResponse+1)
					}
					if mode == "redirect" {
						status = 302
					}
					if mode == "http-error" {
						status = 403
					}
					length := int64(-1)
					if mode == "content-length" {
						length = int64(len(body))
					}
					return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), ContentLength: length, Header: http.Header{"Location": {"https://never-follow.invalid/"}}}, nil
				})
				p, err := newDNSProvider(dnsTestConfig(provider), transport)
				if err != nil {
					t.Fatal(err)
				}
				cleanup, err := p.Present(context.Background(), DNSRecord{Name: "_dnsauth.example.com", Value: "raw"})
				if err == nil || cleanup != nil || strings.Contains(err.Error(), dnsTestSecret) || calls != 1 {
					t.Fatalf("unsafe HTTP handling: calls=%d error=%v", calls, err)
				}
			})
		}
	}
}

type dnsTestContextBody struct{ ctx context.Context }

func (b *dnsTestContextBody) Read([]byte) (int, error) { <-b.ctx.Done(); return 0, b.ctx.Err() }
func (b *dnsTestContextBody) Close() error             { return nil }

func TestDNSHTTPTimeout(t *testing.T) {
	for _, bodyTimeout := range []bool{false, true} {
		t.Run(fmt.Sprint(bodyTimeout), func(t *testing.T) {
			transport := &dnsBoundedTransport{timeout: 5 * time.Millisecond, base: dnsTestTransport(func(r *http.Request) (*http.Response, error) {
				if bodyTimeout {
					return &http.Response{StatusCode: 200, Body: &dnsTestContextBody{ctx: r.Context()}}, nil
				}
				<-r.Context().Done()
				return nil, r.Context().Err()
			})}
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://never-connect.invalid", nil)
			_, err := transport.RoundTrip(req)
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("HTTP headers/body must time out: %v", err)
			}
		})
	}
}
