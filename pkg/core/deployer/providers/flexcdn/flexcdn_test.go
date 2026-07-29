package flexcdn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/flexcdn"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp             = it.Args("FLEXCDN_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fAccessKeyId   string
	fAccessKey     string
	fCertificateId int64
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKey, "ACCESSKEY")
	fp.DefineInt64(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./flexcdn_test.go -args \
	--FLEXCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--FLEXCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--FLEXCDN_SERVERURL="http://127.0.0.1:7788" \
	--FLEXCDN_ACCESSKEYID="your-access-key-id" \
	--FLEXCDN_ACCESSKEY="your-access-key" \
	--FLEXCDN_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToCertificate", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiRole:                  "user",
			AccessKeyId:              fAccessKeyId,
			AccessKey:                fAccessKey,
			AllowInsecureConnections: true,
			DeployTarget:             impl.DEPLOY_TARGET_CERTIFICATE,
			CertificateId:            fCertificateId,
		})
		require.NoError(t, err)

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
