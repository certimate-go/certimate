package bunnycdn_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/bunny-cdn"
)

var (
	fp            = tester.Args("BUNNYCDN_")
	fTestCertPath string
	fTestKeyPath  string
	fApiKey       string
	fPullZoneId   string
	fHostName     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiKey, "APIKEY")
	fp.DefineString(&fPullZoneId, "PULLZONEID")
	fp.DefineString(&fHostName, "HOSTNAME")
}

/*
Shell command to run this test:

	go test -v ./bunny_cdn_test.go -args \
	--BUNNYCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BUNNYCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BUNNYCDN_APITOKEN="your-api-token" \
	--BUNNYCDN_PULLZONEID="your-pull-zone-id" \
	--BUNNYCDN_HOSTNAME="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ApiKey:     fApiKey,
			PullZoneId: fPullZoneId,
			Hostname:   fHostName,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
