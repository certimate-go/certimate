package ratpanel_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ratpanel"
)

var (
	fp             = tester.Args("RATPANEL_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fAccessTokenId int64
	fAccessToken   string
	fSiteName      string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineInt64(&fAccessTokenId, "ACCESSTOKENID")
	fp.DefineString(&fAccessToken, "ACCESSTOKEN")
	fp.DefineString(&fSiteName, "SITENAME")
}

/*
Shell command to run this test:

	go test -v ./ratpanel_test.go -args \
	--RATPANEL_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--RATPANEL_TESTKEYPATH="/path/to/your-test-key.pem" \
	--RATPANEL_SERVERURL="http://127.0.0.1:8888" \
	--RATPANEL_ACCESSTOKENID="your-access-token-id" \
	--RATPANEL_ACCESSTOKEN="your-access-token" \
	--RATPANEL_SITENAME="your-site-name"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			AccessTokenId:            fAccessTokenId,
			AccessToken:              fAccessToken,
			AllowInsecureConnections: true,
			ResourceType:             impl.RESOURCE_TYPE_WEBSITE,
			SiteNames:                []string{fSiteName},
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
