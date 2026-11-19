package pluginhost

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/certimate-go/certimate/pkg/plugin"
)

func TestMirrorDialOnlyUsesPinnedPublicAddresses(t *testing.T) {
	public := netip.MustParseAddr("8.8.8.8")
	for _, scenario := range []struct {
		name string
		ips  []netip.Addr
	}{
		{"mixed-private", []netip.Addr{public, netip.MustParseAddr("127.0.0.1")}},
		{"ipv4-mapped-private", []netip.Addr{netip.MustParseAddr("::ffff:192.168.1.1")}},
		{"translated-loopback", []netip.Addr{netip.MustParseAddr("64:ff9b::7f00:1")}},
		{"shared-network", []netip.Addr{netip.MustParseAddr("100.100.100.200")}},
		{"only-public", []netip.Addr{public}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			dialed := ""
			lookup := func(_ context.Context, _, hostname string) ([]netip.Addr, error) {
				if hostname != "cdn.example" {
					t.Fatalf("looked up %q", hostname)
				}
				return scenario.ips, nil
			}
			stopped := errors.New("test dial")
			dial := func(_ context.Context, _, address string) (net.Conn, error) {
				dialed = address
				return nil, stopped
			}
			_, err := dialPublicMirror(t.Context(), "tcp", "cdn.example:443", lookup, dial)
			if scenario.name == "only-public" {
				if !errors.Is(err, stopped) || dialed != "8.8.8.8:443" {
					t.Fatalf("dial should receive only the resolved IP, not hostname: %q, %v", dialed, err)
				}
			} else if err == nil || dialed != "" {
				t.Fatalf("private address reached the dialer: %q, %v", dialed, err)
			}
		})
	}
	transport := newOfficialMirrorTransport()
	if transport.Proxy != nil || transport.DialTLSContext != nil || transport.DialContext == nil {
		t.Fatal("official transport can bypass validated connections")
	}
}

func TestOfficialTransportRejectsCrossOriginRedirect(t *testing.T) {
	transport := sourceRoundTripper(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "cdn.jsdelivr.net" {
			t.Fatalf("followed a cross-origin redirect to %s", req.URL.Host)
		}
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": {"https://example.com/elsewhere"}}, Body: http.NoBody}, nil
	})
	service := NewMarketService(MarketConfig{CatalogTransport: transport})
	status, err := service.RefreshSource(t.Context(), "official")
	if err != nil || status.Error == "" {
		t.Fatalf("redirect did not fail the catalog refresh: %+v, %v", status, err)
	}
}

func TestOfficialIconLogsMirrorFailures(t *testing.T) {
	const provider = "example"
	const mirror = "https://mirror.example"
	const digest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const redirectTarget = "https://raw.githubusercontent.com/certimate-go/plugins/main/icon.png"
	indexURL := mirror + plugin.OfficialMarketIndexPath
	iconURL := strings.TrimSuffix(indexURL, "/index.json") + "/market/icons/" + digest + ".png"
	for _, scenario := range []struct {
		name   string
		status int
		reason string
	}{
		{"missing", http.StatusNotFound, "icon unavailable"},
		{"blocked redirect", http.StatusMovedPermanently, redirectTarget},
		{"digest mismatch", http.StatusOK, "icon checksum mismatch"},
		{"transport error", 0, "mirror connection failed"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			var output bytes.Buffer
			service := NewMarketService(MarketConfig{
				IndexURL: indexURL,
				Logger:   slog.New(slog.NewJSONHandler(&output, nil)),
				CatalogTransport: sourceRoundTripper(func(req *http.Request) (*http.Response, error) {
					if req.URL.String() != iconURL {
						t.Fatalf("unexpected upstream request: %s", req.URL)
					}
					if scenario.status == 0 {
						return nil, errors.New("mirror connection failed")
					}
					return &http.Response{
						StatusCode: scenario.status,
						Header:     http.Header{"Location": {redirectTarget}},
						Body:       http.NoBody,
					}, nil
				}),
			})
			catalog := service.official
			catalog.cachedAt = time.Now()
			catalog.syncInfo.OfficialMirror = indexURL
			catalog.syncInfo.OfficialState = &plugin.OfficialIndexState{Epoch: 1, Generation: 1}
			catalog.syncInfo.OfficialExpiresAt = catalog.cachedAt.Add(time.Hour)
			catalog.cache = []*plugin.MarketManifest{{
				Manifest: plugin.Manifest{ProviderType: provider},
				Official: &plugin.OfficialMarketRelease{Icon: &plugin.OfficialMarketIcon{SHA256: digest, Format: "png"}},
			}}

			if _, _, err := service.OfficialIcon(t.Context(), provider, digest); err == nil {
				t.Fatal("icon lookup unexpectedly succeeded")
			}
			var record struct {
				Provider       string `json:"provider"`
				MirrorURL      string `json:"mirror_url"`
				UpstreamStatus *int   `json:"upstream_status"`
				Error          string `json:"error"`
			}
			if err := json.Unmarshal(output.Bytes(), &record); err != nil {
				t.Fatalf("expected an icon failure log: %v", err)
			}
			if record.Provider != provider || record.MirrorURL != iconURL || !strings.Contains(record.Error, scenario.reason) {
				t.Fatalf("unexpected failure log: %s", &output)
			}
			if scenario.status == 0 {
				if record.UpstreamStatus != nil {
					t.Fatalf("unexpected upstream status without a response: %s", &output)
				}
			} else if record.UpstreamStatus == nil || *record.UpstreamStatus != scenario.status {
				t.Fatalf("wrong upstream status: %s", &output)
			}
		})
	}
}
