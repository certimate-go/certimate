//go:build tester

package netlify_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/netlify"
)

var (
	fp            = tester.InitArgs("NETLIFY_")
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

	go test -tags=tester -v ./netlify_test.go -args \
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
			DeployTarget: impl.DEPLOY_TARGET_WEBSITE,
			SiteId:       fSiteId,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
