package zenlayerga_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/zenlayer-ga"
)

var (
	fp                 = tester.Args("ZENLAYERGA_")
	fTestCertPath      string
	fTestKeyPath       string
	fAccessKeyId       string
	fAccessKeyPassword string
	fAcceleratorId     string
	fCertificateId     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeyPassword, "ACCESSKEYPASSWORD")
	fp.DefineString(&fAcceleratorId, "ACCELERATORID")
	fp.DefineString(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./zenlayer_ga_test.go -args \
	--ZENLAYERGA_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ZENLAYERGA_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ZENLAYERGA_ACCESSKEYID="your-access-key-id" \
	--ZENLAYERGA_ACCESSKEYPASSWORD="your-access-key-password" \
	--ZENLAYERGA_ACCELERATORID="your-ga-accelerator-id" \
	--ZENLAYERGA_CERTIFICATEID="your-ga-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToAccelerator", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:       fAccessKeyId,
			AccessKeyPassword: fAccessKeyPassword,
			ResourceType:      impl.RESOURCE_TYPE_ACCELERATOR,
			AcceleratorId:     fAcceleratorId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_ToCertificate", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:       fAccessKeyId,
			AccessKeyPassword: fAccessKeyPassword,
			ResourceType:      impl.RESOURCE_TYPE_CERTIFICATE,
			CertificateId:     fCertificateId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
