//go:build tester

package huaweicloudcdn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-cdn"
)

var (
	fp               = tester.InitArgs("HUAWEICLOUDCDN_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./huaweicloud_cdn_test.go -args \
	--HUAWEICLOUDCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--HUAWEICLOUDCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--HUAWEICLOUDCDN_ACCESSKEYID="your-access-key-id" \
	--HUAWEICLOUDCDN_SECRETACCESSKEY="your-secret-access-key" \
	--HUAWEICLOUDCDN_REGION="cn-north-1" \
	--HUAWEICLOUDCDN_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			SecretAccessKey:    fSecretAccessKey,
			Region:             fRegion,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
