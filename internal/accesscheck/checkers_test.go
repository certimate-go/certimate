package accesscheck

import (
	"testing"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

func TestCoreCheckersRegistered(t *testing.T) {
	providerTypes := []domain.AccessProviderType{
		domain.AccessProviderTypeSSH,
		domain.AccessProviderTypeCloudflare,
	}

	for _, providerType := range providerTypes {
		t.Run(string(providerType), func(t *testing.T) {
			entry, ok := DefaultRegistry().Get(string(providerType))
			if !ok {
				t.Fatalf("expected %q to be registered in the default registry", providerType)
			}
			if entry.Source != SourceCore {
				t.Fatalf("source = %q, want %q", entry.Source, SourceCore)
			}
			if entry.OwnerProviderType != "" {
				t.Fatalf("owner provider type = %q, want empty for a core checker", entry.OwnerProviderType)
			}
			if got := entry.Checker.ProviderType(); got != string(providerType) {
				t.Fatalf("provider type = %q, want %q", got, providerType)
			}
			scope := entry.Checker.Scope()
			if len(scope) != 2 || scope[0] != core.AccessCheckItemConnect || scope[1] != core.AccessCheckItemAuth {
				t.Fatalf("scope = %v, want [%s %s]", scope, core.AccessCheckItemConnect, core.AccessCheckItemAuth)
			}
		})
	}
}
