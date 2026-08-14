//go:build tester

package aliyunoss_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-oss"
)

var (
	fp               = tester.InitArgs("ALIYUNOSS_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fRegion          string
	fBucket          string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fBucket, "BUCKET")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./aliyun_oss_test.go -args \
	--ALIYUNOSS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ALIYUNOSS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ALIYUNOSS_ACCESSKEYID="your-access-key-id" \
	--ALIYUNOSS_ACCESSKEYSECRET="your-access-key-secret" \
	--ALIYUNOSS_REGION="cn-hangzhou" \
	--ALIYUNOSS_BUCKET="your-oss-bucket" \
	--ALIYUNOSS_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			Region:          fRegion,
			Bucket:          fBucket,
			Domain:          fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
