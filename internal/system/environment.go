package system

import (
	"context"

	"github.com/certimate-go/certimate/pkg/utils/netutil"
)

type DockerHostInfo struct {
	Reachable bool   `json:"reachable"`
	Address   string `json:"address,omitempty"`
}

type Environment struct {
	DockerHost DockerHostInfo `json:"dockerHost"`
}

type EnvironmentService struct {
	resolver netutil.IPResolver
}

func NewEnvironmentService(resolver netutil.IPResolver) *EnvironmentService {
	return &EnvironmentService{resolver: resolver}
}

func (s *EnvironmentService) GetEnvironment(ctx context.Context) (*Environment, error) {
	addr, ok := netutil.LookupDockerHost(ctx, s.resolver)

	return &Environment{
		DockerHost: DockerHostInfo{
			Reachable: ok,
			Address:   addr,
		},
	}, nil
}
