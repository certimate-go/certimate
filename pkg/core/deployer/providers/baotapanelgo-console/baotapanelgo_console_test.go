package baotapanelgoconsole_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/baotapanelgo-console"
)

var (
	fp            = tester.Args("BAOTAPANELGOCONSOLE_")
	fTestCertPath string
	fTestKeyPath  string
	fServerUrl    string
	fApiKey       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiKey, "APIKEY")
}

/*
Shell command to run this test:

	go test -v ./baotapanelgo_console_test.go -args \
	--BAOTAPANELGOCONSOLE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BAOTAPANELGOCONSOLE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BAOTAPANELGOCONSOLE_SERVERURL="http://127.0.0.1:8888" \
	--BAOTAPANELGOCONSOLE_APIKEY="your-api-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiKey:                   fApiKey,
			AllowInsecureConnections: true,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
