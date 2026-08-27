//go:build tester

package huaweicloudaad_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-aad"
)

var (
	fp               = tester.InitArgs("HUAWEICLOUDAAD_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fInstanceId      string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fInstanceId, "INSTANCEID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./huaweicloud_aad_test.go -args \
	--HUAWEICLOUDAAD_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--HUAWEICLOUDAAD_TESTKEYPATH="/path/to/your-test-key.pem" \
	--HUAWEICLOUDAAD_ACCESSKEYID="your-access-key-id" \
	--HUAWEICLOUDAAD_SECRETACCESSKEY="your-secret-access-key" \
	--HUAWEICLOUDAAD_INSTANCEID="your-aad-instance-id" \
	--HUAWEICLOUDAAD_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			SecretAccessKey:    fSecretAccessKey,
			InstanceId:         fInstanceId,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
