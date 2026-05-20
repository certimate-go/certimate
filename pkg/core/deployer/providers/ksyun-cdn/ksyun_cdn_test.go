package ksyuncdn_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ksyun-cdn"
)

var (
	fp               = tester.Args("KSYUNCDN_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fDomain          string
	fCertificateId   string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fDomain, "DOMAIN")
	fp.DefineString(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./ksyun_cdn_test.go -args \
	--KSYUNCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--KSYUNCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--KSYUNCDN_ACCESSKEYID="your-access-key-id" \
	--KSYUNCDN_SECRETACCESSKEY="your-secret-access-key" \
	--KSYUNCDN_DOMAIN="example.com" \
	--KSYUNCDN_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToDomain", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			SecretAccessKey:    fSecretAccessKey,
			ResourceType:       impl.RESOURCE_TYPE_DOMAIN,
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
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			ResourceType:    impl.RESOURCE_TYPE_CERTIFICATE,
			CertificateId:   fCertificateId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
