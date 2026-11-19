package pluginhost

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"testing"
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
