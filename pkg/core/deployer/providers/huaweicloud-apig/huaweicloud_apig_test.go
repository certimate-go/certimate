package huaweicloudapig_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-apig"
)

var (
	fp               = tester.Args("HUAWEICLOUDAPIG_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fCertificateId   string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./huaweicloud_apig_test.go -args \
	--HUAWEICLOUDAPIG_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--HUAWEICLOUDAPIG_TESTKEYPATH="/path/to/your-test-key.pem" \
	--HUAWEICLOUDAPIG_ACCESSKEYID="your-access-key-id" \
	--HUAWEICLOUDAPIG_SECRETACCESSKEY="your-secret-access-key" \
	--HUAWEICLOUDAPIG_REGION="cn-north-1" \
	--HUAWEICLOUDAPIG_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToCertificate", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			Region:          fRegion,
			DeployTarget:    impl.DEPLOY_TARGET_CERTIFICATE,
			CertificateId:   fCertificateId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
