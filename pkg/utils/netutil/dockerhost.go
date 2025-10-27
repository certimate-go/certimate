package netutil

import (
	"context"
	"net"
	"time"
)

const (
	dockerHostName       = "host.docker.internal"
	defaultLookupTimeout = 500 * time.Millisecond
)

// IPResolver defines the interface required to resolve hostnames to IP addresses.
type IPResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

// LookupDockerHost resolves the docker host name to an IP address.
//
// It returns the resolved IP address and a flag indicating whether the lookup succeeded.
// When no address can be resolved, an empty string and false will be returned.
func LookupDockerHost(ctx context.Context, resolver IPResolver) (string, bool) {
	if resolver == nil {
		resolver = net.DefaultResolver
	}

	if ctx == nil {
		ctx = context.Background()
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, defaultLookupTimeout)
	defer cancel()

	addrs, err := resolver.LookupIPAddr(ctx, dockerHostName)
	if err != nil || len(addrs) == 0 {
		return "", false
	}

	for _, addr := range addrs {
		if ip4 := addr.IP.To4(); ip4 != nil {
			return ip4.String(), true
		}
	}

	for _, addr := range addrs {
		if addr.IP != nil {
			return addr.IP.String(), true
		}
	}

	return "", false
}
