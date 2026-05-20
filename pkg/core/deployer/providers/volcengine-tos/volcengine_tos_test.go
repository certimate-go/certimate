package volcenginetos_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-tos"
)

var (
	fp               = tester.Args("VOLCENGINETOS_")
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

	go test -v ./volcengine_tos_test.go -args \
	--VOLCENGINETOS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--VOLCENGINETOS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--VOLCENGINETOS_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINETOS_ACCESSKEYSECRET="your-access-key-secret" \
	--VOLCENGINETOS_REGION="cn-beijing" \
	--VOLCENGINETOS_BUCKET="your-tos-bucket" \
	--VOLCENGINETOS_DOMAIN="example.com"
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
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
