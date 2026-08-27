//go:build tester

package flyio_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/flyio"
)

var (
	fp            = tester.InitArgs("FLYIO_")
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

	go test -tags=tester -v ./flyio_test.go -args \
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
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
