package flyio_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/flyio"
)

var (
	fp            = tester.Args("FLYIO_")
	fTestCertPath string
	fTestKeyPath  string
	fApiToken     string
	fAppName      string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineString(&fAppName, "APPNAME")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./flyio_test.go -args \
	--FLYIO_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--FLYIO_TESTKEYPATH="/path/to/your-test-key.pem" \
	--FLYIO_APITOKEN="your-api-token" \
	--FLYIO_APPNAME="your-app-name" \
	--FLYIO_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ApiToken: fApiToken,
			AppName:  fAppName,
			Domain:   fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
