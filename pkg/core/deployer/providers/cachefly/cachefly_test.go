package cachefly_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/cachefly"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp            = it.Args("CACHEFLY_")
	fTestCertPath string
	fTestKeyPath  string
	fApiToken     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiToken, "APITOKEN")
}

/*
Shell command to run this test:

	go test -v ./cachefly_test.go -args \
	--CACHEFLY_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--CACHEFLY_TESTKEYPATH="/path/to/your-test-key.pem" \
	--CACHEFLY_APITOKEN="your-api-token"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ApiToken: fApiToken,
		})
		require.NoError(t, err)

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
