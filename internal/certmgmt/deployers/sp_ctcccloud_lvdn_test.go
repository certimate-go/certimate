package deployers

import (
	"testing"

	"github.com/certimate-go/certimate/internal/domain"
)

func TestCTCCCloudLVDNProviderRegistered(t *testing.T) {
	const provider = domain.DeploymentProviderType("ctcccloud-lvdn")

	if _, err := Registries.Get(provider); err != nil {
		t.Fatalf("expected provider %q to be registered: %v", provider, err)
	}
}
