package byteplustos_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/byteplus-tos"
)

var (
	fp               = tester.Args("BYTEPLUSTOS_")
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

	go test -v ./byteplus_tos_test.go -args \
	--BYTEPLUSTOS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BYTEPLUSTOS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BYTEPLUSTOS_ACCESSKEYID="your-access-key-id" \
	--BYTEPLUSTOS_SECRETACCESSKEY="your-secret-access-key" \
	--BYTEPLUSTOS_REGION="cn-beijing" \
	--BYTEPLUSTOS_BUCKET="your-tos-bucket" \
	--BYTEPLUSTOS_DOMAIN="example.com"
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
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
