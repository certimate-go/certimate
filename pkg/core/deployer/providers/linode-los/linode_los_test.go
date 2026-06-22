package linodelos_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/linode-los"
)

var (
	fp            = tester.Args("LINODELOS_")
	fTestCertPath string
	fTestKeyPath  string
	fApiToken     string
	fRegionId     string
	fBucket       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiToken, "ACCESSTOKEN")
	fp.DefineString(&fRegionId, "REGIONID")
	fp.DefineString(&fBucket, "BUCKET")
}

/*
Shell command to run this test:

	go test -v ./linode_los_test.go -args \
	--LINODELOS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--LINODELOS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--LINODELOS_ACCESSTOKEN="your-api-token" \
	--LINODELOS_REGIONID="your-bucket-region" \
	--LINODELOS_BUCKET="your-bucket-name"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessToken: fApiToken,
			RegionId:    fRegionId,
			Bucket:      fBucket,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
