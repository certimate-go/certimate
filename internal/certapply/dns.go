package certapply

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DNSRecord is a fully qualified DNS name and an unmodified TXT value.
type DNSRecord struct {
	Name  string
	Value string
}

// A nil cleanup preserves existing records; a non-nil cleanup owns only the newly created ID.
type DNSProvider interface {
	Present(context.Context, DNSRecord) (func(context.Context) error, error)
}

type DNSProviderConfig struct {
	Provider     string
	AccessConfig map[string]any
	TTL          int
}

const (
	dnsHTTPTimeout = 30 * time.Second
	dnsMaxResponse = 1 << 20
	dnsPageSize    = 100
)

func NewDNSProvider(config DNSProviderConfig) (DNSProvider, error) {
	return newDNSProvider(config, http.DefaultTransport)
}

func newDNSProvider(config DNSProviderConfig, transport http.RoundTripper) (DNSProvider, error) {
	if config.TTL < 0 {
		return nil, errors.New("DNS TTL must not be negative")
	}
	if config.TTL == 0 {
		config.TTL = 600
	}
	transport = &dnsBoundedTransport{base: transport, timeout: dnsHTTPTimeout}
	switch config.Provider {
	case "tencentcloud":
		return newDNSTencent(config, transport)
	case "aliyun":
		return newDNSAliyun(config, transport)
	case "cloudflare":
		return newDNSCloudflare(config, transport)
	default:
		return nil, errors.New("unsupported DNS provider")
	}
}

func dnsDecodeAccess(config map[string]any, target any) error {
	data, err := json.Marshal(config)
	if err == nil {
		err = json.Unmarshal(data, target)
	}
	if err != nil {
		return errors.New("invalid DNS access configuration")
	}
	return nil
}

func dnsCredential(value string, optional bool) bool {
	if value == "" {
		return optional
	}
	for _, c := range value {
		if c < 33 || c > 126 {
			return false
		}
	}
	return true
}

func dnsName(name string) string {
	return strings.ToLower(strings.TrimSuffix(name, "."))
}

func dnsValidateRecord(record DNSRecord) (DNSRecord, error) {
	record.Name = dnsName(record.Name)
	if len(record.Name) == 0 || len(record.Name) > 253 {
		return record, errors.New("invalid DNS record name")
	}
	for i, label := range strings.Split(record.Name, ".") {
		if label == "*" && i == 0 {
			continue
		}
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return record, errors.New("invalid DNS record name")
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
				return record, errors.New("invalid DNS record name (use ASCII/punycode)")
			}
		}
	}
	return record, nil
}

func dnsWithinZone(name, zone string) bool {
	return zone != "" && (name == zone || strings.HasSuffix(name, "."+zone))
}

func dnsRelativeName(name, zone string) string {
	if name == zone {
		return "@"
	}
	return strings.TrimSuffix(name, "."+zone)
}

func dnsString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// SDK errors may include credentials; preserve only cancellation and deadline identities.
func dnsAPIError(ctx context.Context, operation string, err error) error {
	if ctx.Err() != nil {
		return fmt.Errorf("DNS %s: %w", operation, ctx.Err())
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("DNS %s: %w", operation, context.Canceled)
	}
	var timeout net.Error
	if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &timeout) && timeout.Timeout() {
		return fmt.Errorf("DNS %s: %w", operation, context.DeadlineExceeded)
	}
	return fmt.Errorf("DNS %s failed", operation)
}

func dnsCleanup(remove func(context.Context) error) func(context.Context) error {
	var mu sync.Mutex
	done := false
	return func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		if done {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := remove(ctx); err != nil {
			return err
		}
		done = true
		return nil
	}
}

// Bound SDK body reads and reject redirects before credentials can be forwarded.
type dnsBoundedTransport struct {
	base    http.RoundTripper
	timeout time.Duration
}

func (t *dnsBoundedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(request.Context(), t.timeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	response, err := t.base.RoundTrip(request.Clone(ctx))
	if err != nil {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
		return nil, dnsAPIError(ctx, "HTTP request", err)
	}
	if response == nil || response.Body == nil {
		return nil, errors.New("DNS HTTP response is empty")
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 && response.StatusCode < 400 {
		return nil, errors.New("DNS HTTP redirects are not allowed")
	}
	if response.ContentLength > dnsMaxResponse {
		return nil, errors.New("DNS HTTP response exceeds size limit")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, dnsMaxResponse+1))
	if err != nil {
		return nil, dnsAPIError(ctx, "HTTP response read", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(body) > dnsMaxResponse {
		return nil, errors.New("DNS HTTP response exceeds size limit")
	}
	response.Body = io.NopCloser(bytes.NewReader(body))
	response.ContentLength = int64(len(body))
	return response, nil
}
