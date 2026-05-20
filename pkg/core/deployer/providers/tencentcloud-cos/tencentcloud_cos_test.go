package tencentcloudcos_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-cos"
)

var (
	fp            = tester.Args("TENCENTCLOUDCOS_")
	fTestCertPath string
	fTestKeyPath  string
	fSecretId     string
	fSecretKey    string
	fRegion       string
	fBucket       string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fBucket, "BUCKET")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./tencentcloud_cos_test.go -args \
	--TENCENTCLOUDCOS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDCOS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDCOS_SECRETID="your-secret-id" \
	--TENCENTCLOUDCOS_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDCOS_REGION="ap-guangzhou" \
	--TENCENTCLOUDCOS_BUCKET="your-cos-bucket" \
	--TENCENTCLOUDCOS_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:  fSecretId,
			SecretKey: fSecretKey,
			Region:    fRegion,
			Bucket:    fBucket,
			Domain:    fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
