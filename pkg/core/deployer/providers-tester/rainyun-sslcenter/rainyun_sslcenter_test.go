//go:build tester

package rainyunsslcenter_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/rainyun-sslcenter"
)

var (
	fp            = tester.InitArgs("RAINYUNSSLCENTER_")
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

	go test -tags=tester -v ./rainyun_sslcenter_test.go -args \
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
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
