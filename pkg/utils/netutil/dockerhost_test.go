package netutil

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

type mockResolver struct {
	addrs []net.IPAddr
	err   error
}

func (m mockResolver) LookupIPAddr(context.Context, string) ([]net.IPAddr, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.addrs, nil
}

func TestLookupDockerHost_ReturnsIPv4WhenAvailable(t *testing.T) {
	resolver := mockResolver{addrs: []net.IPAddr{{IP: net.ParseIP("172.17.0.1")}}}

	addr, ok := LookupDockerHost(context.Background(), resolver)

	if !ok {
		t.Fatalf("expected lookup to succeed")
	}
	if addr != "172.17.0.1" {
		t.Fatalf("expected address '172.17.0.1', got %q", addr)
	}
}

func TestLookupDockerHost_FallsBackToIPv6(t *testing.T) {
	resolver := mockResolver{addrs: []net.IPAddr{{IP: net.ParseIP("::1")}}}

	addr, ok := LookupDockerHost(context.Background(), resolver)

	if !ok {
		t.Fatalf("expected lookup to succeed")
	}
	if addr != "::1" {
		t.Fatalf("expected address '::1', got %q", addr)
	}
}

func TestLookupDockerHost_ReturnsFalseOnError(t *testing.T) {
	resolver := mockResolver{err: errors.New("lookup failed")}

	if addr, ok := LookupDockerHost(context.Background(), resolver); ok || addr != "" {
		t.Fatalf("expected lookup to fail, got ok=%v addr=%q", ok, addr)
	}
}

func TestLookupDockerHost_TimeoutContext(t *testing.T) {
	resolver := mockResolver{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	cancel()

	if addr, ok := LookupDockerHost(ctx, resolver); ok || addr != "" {
		t.Fatalf("expected lookup to fail due to timeout, got ok=%v addr=%q", ok, addr)
	}
}
