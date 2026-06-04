package cloudflaressl_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/cloudflare-ssl"
)

var (
	fp             = tester.Args("CLOUDFLARESSL_")
	fTestCertPath  string
	fTestKeyPath   string
	fApiToken      string
	fZoneId        string
	fCertificateId string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineString(&fZoneId, "ZONEID")
	fp.DefineString(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./cloudflare_ssl_test.go -args \
	--CLOUDFLARESSL_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--CLOUDFLARESSL_TESTKEYPATH="/path/to/your-test-key.pem" \
	--CLOUDFLARESSL_APITOKEN="your-api-token" \
	--CLOUDFLARESSL_ZONEID="your-zone-id" \
	--CLOUDFLARESSL_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ApiToken:      fApiToken,
			ZoneId:        fZoneId,
			CertificateId: fCertificateId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
