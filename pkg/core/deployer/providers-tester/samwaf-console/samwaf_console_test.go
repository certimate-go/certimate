//go:build tester

package samwafconsole_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/samwaf-console"
)

var (
	fp            = tester.InitArgs("SAMWAFCONSOLE_")
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

	go test -tags=tester -v ./samwaf_console_test.go -args \
	--SAMWAFCONSOLE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--SAMWAFCONSOLE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--SAMWAFCONSOLE_SERVERURL="http://127.0.0.1:26666" \
	--SAMWAFCONSOLE_APIKEY="your-api-key" \
	--SAMWAFCONSOLE_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiKey:                   fApiKey,
			AllowInsecureConnections: true,
			AutoRestart:              true,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
