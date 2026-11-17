//go:build tester

package awscloudfront_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aws-cloudfront"
)

var (
	fp               = tester.InitArgs("AWSCLOUDFRONT_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fDistributionId  string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fDistributionId, "DISTRIBUTIONID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./aws_cloudfront_test.go -args \
	--AWSCLOUDFRONT_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--AWSCLOUDFRONT_TESTKEYPATH="/path/to/your-test-key.pem" \
	--AWSCLOUDFRONT_ACCESSKEYID="your-access-key-id" \
	--AWSCLOUDFRONT_SECRETACCESSKEY="your-secret-access-id" \
	--AWSCLOUDFRONT_REGION="us-east-1" \
	--AWSCLOUDFRONT_DISTRIBUTIONID="your-distribution-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:       fAccessKeyId,
			SecretAccessKey:   fSecretAccessKey,
			Region:            fRegion,
			DistributionId:    fDistributionId,
			CertificateSource: impl.CERTIFICATE_SOURCE_ACM,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
