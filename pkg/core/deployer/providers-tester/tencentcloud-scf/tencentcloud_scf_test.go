//go:build tester

package tencentcloudscf_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-scf"
)

var (
	fp            = tester.InitArgs("TENCENTCLOUDSCF_")
	fTestCertPath string
	fTestKeyPath  string
	fSecretId     string
	fSecretKey    string
	fRegion       string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./tencentcloud_scf_test.go -args \
	--TENCENTCLOUDSCF_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDSCF_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDSCF_SECRETID="your-secret-id" \
	--TENCENTCLOUDSCF_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDSCF_REGION="ap-guangzhou" \
	--TENCENTCLOUDSCF_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:           fSecretId,
			SecretKey:          fSecretKey,
			Region:             fRegion,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
