package ucloudus3_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-us3"
)

var (
	fp            = tester.Args("UCLOUDUS3_")
	fTestCertPath string
	fTestKeyPath  string
	fPrivateKey   string
	fPublicKey    string
	fRegion       string
	fBucket       string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fPrivateKey, "PRIVATEKEY")
	fp.DefineString(&fPublicKey, "PUBLICKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fBucket, "BUCKET")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./ucloud_us3_test.go -args \
	--UCLOUDUS3_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UCLOUDUS3_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UCLOUDUS3_PRIVATEKEY="your-private-key" \
	--UCLOUDUS3_PUBLICKEY="your-public-key" \
	--UCLOUDUS3_REGION="cn-bj2" \
	--UCLOUDUS3_BUCKET="your-us3-bucket" \
	--UCLOUDUS3_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			PrivateKey: fPrivateKey,
			PublicKey:  fPublicKey,
			Region:     fRegion,
			Bucket:     fBucket,
			Domain:     fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
