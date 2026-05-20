package zenlayercdn_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/zenlayer-cdn"
)

var (
	fp                 = tester.Args("ZENLAYERCDN_")
	fTestCertPath      string
	fTestKeyPath       string
	fAccessKeyId       string
	fAccessKeyPassword string
	fDomain            string
	fCertificateId     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeyPassword, "ACCESSKEYPASSWORD")
	fp.DefineString(&fDomain, "DOMAIN")
	fp.DefineString(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./zenlayer_cdn_test.go -args \
	--ZENLAYERCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ZENLAYERCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ZENLAYERCDN_ACCESSKEYID="your-access-key-id" \
	--ZENLAYERCDN_ACCESSKEYPASSWORD="your-access-key-secret" \
	--ZENLAYERCDN_DOMAIN="example.com" \
	--ZENLAYERCDN_CERTIFICATEID="your-cdn-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToDomain", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			AccessKeyPassword:  fAccessKeyPassword,
			DeployTarget:       impl.DEPLOY_TARGET_DOMAIN,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
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
			DeployTarget:      impl.DEPLOY_TARGET_CERTIFICATE,
			CertificateId:     fCertificateId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
