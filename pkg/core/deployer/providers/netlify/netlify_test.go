package netlify_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/netlify"
)

var (
	fp            = tester.Args("NETLIFY_")
	fTestCertPath string
	fTestKeyPath  string
	fApiToken     string
	fSiteId       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineString(&fSiteId, "SITEID")
}

/*
Shell command to run this test:

	go test -v ./netlify_test.go -args \
	--NETLIFY_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--NETLIFY_TESTKEYPATH="/path/to/your-test-key.pem" \
	--NETLIFY_APITOKEN="your-api-token" \
	--NETLIFY_SITEID="your-site-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ApiToken:     fApiToken,
			ResourceType: impl.RESOURCE_TYPE_WEBSITE,
			SiteId:       fSiteId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
