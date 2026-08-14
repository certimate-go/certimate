//go:build tester

package wangsucertificate_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/wangsu-certificate"
)

var (
	fp               = tester.InitArgs("WANGSUCERTIFICATE_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fCertificateId   string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./wangsu_certificate_test.go -args \
	--WANGSUCERTIFICATE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--WANGSUCERTIFICATE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--WANGSUCERTIFICATE_ACCESSKEYID="your-access-key-id" \
	--WANGSUCERTIFICATE_ACCESSKEYSECRET="your-access-key-secret" \
	--WANGSUCERTIFICATE_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			CertificateId:   fCertificateId,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
