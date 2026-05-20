package baotawaf_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/baotawaf"
)

var (
	fp            = tester.Args("BAOTAWAF_")
	fTestCertPath string
	fTestKeyPath  string
	fServerUrl    string
	fApiKey       string
	fSiteName     string
	fSitePort     int64
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiKey, "APIKEY")
	fp.DefineString(&fSiteName, "SITENAME")
	fp.DefineInt64(&fSitePort, "SITEPORT")
}

/*
Shell command to run this test:

	go test -v ./baotawaf_test.go -args \
	--BAOTAWAF_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BAOTAWAF_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BAOTAWAF_SERVERURL="http://127.0.0.1:8888" \
	--BAOTAWAF_APIKEY="your-api-key" \
	--BAOTAWAF_SITENAME="your-site-name" \
	--BAOTAWAF_SITEPORT=443
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiKey:                   fApiKey,
			AllowInsecureConnections: true,
			SiteNames:                []string{fSiteName},
			SitePort:                 int32(fSitePort),
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
