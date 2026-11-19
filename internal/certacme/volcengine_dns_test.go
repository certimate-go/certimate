package certacme

import (
	"strings"
	"testing"

	"github.com/go-acme/lego/v5/lego"
)

func TestVolcengineDNSRequiresPlugin(t *testing.T) {
	t.Setenv("LEGO_DISABLE_CNAME_SUPPORT", "false")
	client := &ACMEClient{client: &lego.Client{}}
	_, err := client.ObtainCertificate(t.Context(), &ObtainCertificateRequest{
		ChallengeType: "dns-01",
		Provider:      "volcengine-dns",
	})
	if err == nil || !strings.Contains(err.Error(), "install its plugin") || !strings.Contains(err.Error(), "volcengine-dns") {
		t.Fatalf("missing plugin should give an installation hint, got %v", err)
	}
}
