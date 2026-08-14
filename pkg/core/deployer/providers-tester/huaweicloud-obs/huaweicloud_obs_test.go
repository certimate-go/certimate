//go:build tester

package huaweicloudobs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-obs"
)

var (
	fp               = tester.InitArgs("HUAWEICLOUDOBS_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fBucket          string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fBucket, "BUCKET")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./huaweicloud_obs_test.go -args \
	--HUAWEICLOUDOBS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--HUAWEICLOUDOBS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--HUAWEICLOUDOBS_ACCESSKEYID="your-access-key-id" \
	--HUAWEICLOUDOBS_SECRETACCESSKEY="your-secret-access-key" \
	--HUAWEICLOUDOBS_REGION="cn-north-4" \
	--HUAWEICLOUDOBS_BUCKET="your-bucket" \
	--HUAWEICLOUDOBS_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			Region:          fRegion,
			Bucket:          fBucket,
			Domain:          fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
