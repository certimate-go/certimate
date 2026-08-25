//go:build tester

package proxmoxbs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/proxmoxbs"
)

var (
	fp              = tester.InitArgs("PROXMOXBS_")
	fTestCertPath   string
	fTestKeyPath    string
	fServerUrl      string
	fApiToken       string
	fApiTokenSecret string
	fNodeName       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineString(&fApiTokenSecret, "APITOKENSECRET")
	fp.DefineString(&fNodeName, "NODENAME")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./proxmoxbs_test.go -args \
	--PROXMOXBS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--PROXMOXBS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--PROXMOXBS_SERVERURL="http://127.0.0.1:8007" \
	--PROXMOXBS_APITOKEN="your-api-token" \
	--PROXMOXBS_APITOKENSECRET="your-api-token-secret" \
	--PROXMOXBS_NODENAME="your-node-name"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiToken:                 fApiToken,
			ApiTokenSecret:           fApiTokenSecret,
			AllowInsecureConnections: true,
			NodeName:                 fNodeName,
			AutoRestart:              true,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
