//go:build tester

package baotapanelgo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/baotapanelgo"
)

var (
	fp            = tester.InitArgs("BAOTAPANELGO_")
	fTestCertPath string
	fTestKeyPath  string
	fServerUrl    string
	fApiKey       string
	fSiteType     string
	fSiteName     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiKey, "APIKEY")
	fp.DefineString(&fSiteType, "SITETYPE")
	fp.DefineString(&fSiteName, "SITENAME")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./baotapanelgo_test.go -args \
	--BAOTAPANELGO_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BAOTAPANELGO_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BAOTAPANELGO_SERVERURL="http://127.0.0.1:8888" \
	--BAOTAPANELGO_APIKEY="your-api-key" \
	--BAOTAPANELGO_SITETYPE="your-site-type" \
	--BAOTAPANELGO_SITENAME="your-site-name"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiKey:                   fApiKey,
			AllowInsecureConnections: true,
			SiteType:                 fSiteType,
			SiteNames:                []string{fSiteName},
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
