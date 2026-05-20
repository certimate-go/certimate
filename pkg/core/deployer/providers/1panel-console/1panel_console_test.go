package onepanelconsole_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/1panel-console"
)

var (
	fp            = tester.Args("1PANELCONSOLE_")
	fTestCertPath string
	fTestKeyPath  string
	fServerUrl    string
	fApiVersion   string
	fApiKey       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiVersion, "APIVERSION", "v2")
	fp.DefineString(&fApiKey, "APIKEY")
}

/*
Shell command to run this test:

	go test -v ./1panel_console_test.go -args \
	--1PANELCONSOLE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--1PANELCONSOLE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--1PANELCONSOLE_SERVERURL="http://127.0.0.1:20410" \
	--1PANELCONSOLE_APIVERSION="v2" \
	--1PANELCONSOLE_APIKEY="your-api-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiVersion:               fApiVersion,
			ApiKey:                   fApiKey,
			AllowInsecureConnections: true,
			AutoRestart:              true,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
