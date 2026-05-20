package upyuncdn_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/upyun-cdn"
)

var (
	fp            = tester.Args("UPYUNCDN_")
	fTestCertPath string
	fTestKeyPath  string
	fUsername     string
	fPassword     string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./upyun_cdn_test.go -args \
	--UPYUNCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UPYUNCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UPYUNCDN_USERNAME="your-username" \
	--UPYUNCDN_PASSWORD="your-password" \
	--UPYUNCDN_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			Username: fUsername,
			Password: fPassword,
			Domain:   fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
