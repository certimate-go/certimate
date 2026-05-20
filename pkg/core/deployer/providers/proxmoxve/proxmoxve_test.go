package proxmoxve_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/proxmoxve"
)

var (
	fp              = tester.Args("PROXMOXVE_")
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

	go test -v ./proxmoxve_test.go -args \
	--PROXMOXVE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--PROXMOXVE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--PROXMOXVE_SERVERURL="http://127.0.0.1:8006" \
	--PROXMOXVE_APITOKEN="your-api-token" \
	--PROXMOXVE_APITOKENSECRET="your-api-token-secret" \
	--PROXMOXVE_NODENAME="your-cluster-node-name"
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
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
