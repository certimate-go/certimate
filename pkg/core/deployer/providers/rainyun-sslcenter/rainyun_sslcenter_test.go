package rainyunsslcenter_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/rainyun-sslcenter"
)

var (
	fp            = tester.Args("RAINYUNSSLCENTER_")
	fTestCertPath string
	fTestKeyPath  string
	fApiKey       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiKey, "APIKEY")
}

/*
Shell command to run this test:

	go test -v ./rainyun_sslcenter_test.go -args \
	--RAINYUNRCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--RAINYUNRCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--RAINYUNRCDN_APIKEY="your-api-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ApiKey: fApiKey,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
