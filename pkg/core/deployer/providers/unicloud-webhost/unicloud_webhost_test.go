package unicloudwebhost_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/unicloud-webhost"
)

var (
	fp             = tester.Args("UNICLOUDWEBHOST_")
	fTestCertPath  string
	fTestKeyPath   string
	fUsername      string
	fPassword      string
	fSpaceProvider string
	fSpaceId       string
	fDomain        string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
	fp.DefineString(&fSpaceProvider, "SPACEPROVIDER")
	fp.DefineString(&fSpaceId, "SPACEID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./unicloud_webhost_test.go -args \
	--UNICLOUDWEBHOST_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UNICLOUDWEBHOST_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UNICLOUDWEBHOST_USERNAME="your-username" \
	--UNICLOUDWEBHOST_PASSWORD="your-password" \
	--UNICLOUDWEBHOST_SPACEPROVIDER="aliyun/tencent" \
	--UNICLOUDWEBHOST_SPACEID="your-space-id" \
	--UNICLOUDWEBHOST_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			Username:      fUsername,
			Password:      fPassword,
			SpaceProvider: fSpaceProvider,
			SpaceId:       fSpaceId,
			Domain:        fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
