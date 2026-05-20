package cpanel_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/cpanel"
)

var (
	fp            = tester.Args("CPANEL_")
	fTestCertPath string
	fTestKeyPath  string
	fServerUrl    string
	fUsername     string
	fApiToken     string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./cpanel_test.go -args \
	--CPANEL_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--CPANEL_TESTKEYPATH="/path/to/your-test-key.pem" \
	--CPANEL_SERVERURL="http://127.0.0.1:2082" \
	--CPANEL_USERNAME="your-username" \
	--CPANEL_APITOKEN="your-api-token" \
	--CPANEL_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToWebsite", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			Username:                 fUsername,
			ApiToken:                 fApiToken,
			AllowInsecureConnections: true,
			DeployTarget:             impl.DEPLOY_TARGET_WEBSITE,
			Domain:                   fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
