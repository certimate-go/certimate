package upyunfile_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/upyun-file"
)

var (
	fp            = tester.Args("UPYUNFILE_")
	fTestCertPath string
	fTestKeyPath  string
	fUsername     string
	fPassword     string
	fBucket       string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
	fp.DefineString(&fBucket, "BUCKET")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./upyun_file_test.go -args \
	--UPYUNFILE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UPYUNFILE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UPYUNFILE_USERNAME="your-username" \
	--UPYUNFILE_PASSWORD="your-password" \
	--UPYUNFILE_BUCKET="your-bucket" \
	--UPYUNFILE_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			Username: fUsername,
			Password: fPassword,
			Bucket:   fBucket,
			Domain:   fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
