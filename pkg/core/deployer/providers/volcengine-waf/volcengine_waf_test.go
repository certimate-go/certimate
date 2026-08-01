package volcenginewaf_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-waf"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp               = it.Args("VOLCENGINEWAF_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fAccessMode      string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fAccessMode, "ACCESSMODE")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./volcengine_waf_test.go -args \
	--VOLCENGINEWAF_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--VOLCENGINEWAF_TESTKEYPATH="/path/to/your-test-key.pem" \
	--VOLCENGINEWAF_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINEWAF_SECRETACCESSKEY="your-secret-access-key" \
	--VOLCENGINEWAF_REGION="cn-beijing" \
	--VOLCENGINEWAF_ACCESSMODE="cname" \
	--VOLCENGINEWAF_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			Region:          fRegion,
			AccessMode:      fAccessMode,
			Domain:          fDomain,
		})
		require.NoError(t, err)

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
