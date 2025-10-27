package system

import (
	"context"
	"net"
	"testing"
)

type stubResolver struct {
	addrs []net.IPAddr
	err   error
}

func (s stubResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return s.addrs, s.err
}

func TestEnvironmentService_GetEnvironment(t *testing.T) {
	resolver := stubResolver{addrs: []net.IPAddr{{IP: net.ParseIP("172.17.0.1")}}}
	svc := NewEnvironmentService(resolver)

	env, err := svc.GetEnvironment(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env.DockerHost.Address != "172.17.0.1" {
		t.Fatalf("expected docker host address '172.17.0.1', got %q", env.DockerHost.Address)
	}
	if !env.DockerHost.Reachable {
		t.Fatalf("expected docker host to be reachable")
	}
}

func TestEnvironmentService_GetEnvironment_Unreachable(t *testing.T) {
	resolver := stubResolver{}
	svc := NewEnvironmentService(resolver)

	env, err := svc.GetEnvironment(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env.DockerHost.Address != "" {
		t.Fatalf("expected empty address, got %q", env.DockerHost.Address)
	}
	if env.DockerHost.Reachable {
		t.Fatalf("expected docker host to be unreachable")
	}
}
