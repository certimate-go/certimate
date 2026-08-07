package baotapanel_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/baotapanel"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp            = it.Args("BAOTAPANEL_")
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

	go test -v ./baotapanel_test.go -args \
	--BAOTAPANEL_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BAOTAPANEL_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BAOTAPANEL_SERVERURL="http://127.0.0.1:8888" \
	--BAOTAPANEL_APIKEY="your-api-key" \
	--BAOTAPANEL_SITETYPE="php" \
	--BAOTAPANEL_SITENAME="your-site-name"
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

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
