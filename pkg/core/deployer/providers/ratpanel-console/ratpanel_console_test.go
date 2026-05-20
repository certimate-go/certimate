package ratpanelconsole_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ratpanel-console"
)

var (
	fp             = tester.Args("RATPANELCONSOLE_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fAccessTokenId int64
	fAccessToken   string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineInt64(&fAccessTokenId, "ACCESSTOKENID")
	fp.DefineString(&fAccessToken, "ACCESSTOKEN")
}

/*
Shell command to run this test:

	go test -v ./ratpanel_console_test.go -args \
	--RATPANELCONSOLE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--RATPANELCONSOLE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--RATPANELCONSOLE_SERVERURL="http://127.0.0.1:8888" \
	--RATPANELCONSOLE_ACCESSTOKENID="your-access-token-id" \
	--RATPANELCONSOLE_ACCESSTOKEN="your-access-token"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			AccessTokenId:            fAccessTokenId,
			AccessToken:              fAccessToken,
			AllowInsecureConnections: true,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
