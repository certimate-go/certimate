package awscloudfront_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aws-cloudfront"
)

var (
	fp               = tester.Args("AWSCLOUDFRONT_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fDistribuitionId string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fDistribuitionId, "DISTRIBUTIONID")
}

/*
Shell command to run this test:

	go test -v ./aws_cloudfront_test.go -args \
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
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			Region:          fRegion,
			DistributionId:  fDistribuitionId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
