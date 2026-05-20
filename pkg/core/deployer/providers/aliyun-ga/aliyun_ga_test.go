package aliyunga_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-ga"
)

var (
	fp               = tester.Args("ALIYUNGA_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fAcceleratorId   string
	fListenerId      string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fAcceleratorId, "ACCELERATORID")
	fp.DefineString(&fListenerId, "LISTENERID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./aliyun_ga_test.go -args \
	--ALIYUNGA_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ALIYUNGA_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ALIYUNGA_ACCESSKEYID="your-access-key-id" \
	--ALIYUNGA_ACCESSKEYSECRET="your-access-key-secret" \
	--ALIYUNGA_ACCELERATORID="your-ga-accelerator-id" \
	--ALIYUNGA_LISTENERID="your-ga-listener-id" \
	--ALIYUNGA_DOMAIN="your-ga-sni-domain"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToAccelerator", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			ResourceType:    impl.RESOURCE_TYPE_ACCELERATOR,
			AcceleratorId:   fAcceleratorId,
			Domain:          fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_ToListener", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			ResourceType:    impl.RESOURCE_TYPE_LISTENER,
			ListenerId:      fListenerId,
			Domain:          fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
