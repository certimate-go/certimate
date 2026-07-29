package volcenginedcdn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-dcdn"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp               = it.Args("VOLCENGINEDCDN_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./volcengine_dcdn_test.go -args \
	--VOLCENGINEDCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--VOLCENGINEDCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--VOLCENGINEDCDN_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINEDCDN_SECRETACCESSKEY="your-secret-access-key" \
	--VOLCENGINEDCDN_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			SecretAccessKey:    fSecretAccessKey,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
