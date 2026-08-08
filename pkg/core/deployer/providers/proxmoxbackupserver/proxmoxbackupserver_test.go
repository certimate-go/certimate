package proxmoxbackupserver_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/proxmoxbackupserver"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp              = it.Args("PROXMOXBACKUPSERVER_")
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

	go test -v ./proxmoxbackupserver_test.go -args \
	--PROXMOXBACKUPSERVER_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--PROXMOXBACKUPSERVER_TESTKEYPATH="/path/to/your-test-key.pem" \
	--PROXMOXBACKUPSERVER_SERVERURL="http://127.0.0.1:8007" \
	--PROXMOXBACKUPSERVER_APITOKEN="your-api-token" \
	--PROXMOXBACKUPSERVER_APITOKENSECRET="your-api-token-secret" \
	--PROXMOXBACKUPSERVER_NODENAME="your-node-name"
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

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
