//go:build tester

package tencentcloudecdn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-ecdn"
)

var (
	fp            = tester.InitArgs("TENCENTCLOUDECDN_")
	fTestCertPath string
	fTestKeyPath  string
	fSecretId     string
	fSecretKey    string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./tencentcloud_ecdn_test.go -args \
	--TENCENTCLOUDECDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDECDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDECDN_SECRETID="your-secret-id" \
	--TENCENTCLOUDECDN_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDECDN_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:           fSecretId,
			SecretKey:          fSecretKey,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
