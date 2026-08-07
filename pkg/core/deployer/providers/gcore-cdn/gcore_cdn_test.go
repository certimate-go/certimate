package gcorecdn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/gcore-cdn"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp            = it.Args("GCORECDN_")
	fTestCertPath string
	fTestKeyPath  string
	fApiToken     string
	fResourceId   int64
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineInt64(&fResourceId, "RESOURCEID")
}

/*
Shell command to run this test:

	go test -v ./gcore_cdn_test.go -args \
	--GCORECDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--GCORECDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--GCORECDN_APITOKEN="your-api-token" \
	--GCORECDN_RESOURCEID="your-cdn-resource-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ApiToken:   fApiToken,
			ResourceId: fResourceId,
		})
		require.NoError(t, err)

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
