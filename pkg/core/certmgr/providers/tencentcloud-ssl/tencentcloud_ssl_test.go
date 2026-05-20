package tencentcloudssl_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
)

var (
	fp            = tester.Args("TENCENTCLOUDSSL_")
	fTestCertPath string
	fTestKeyPath  string
	fSecretId     string
	fSecretKey    string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
}

/*
Shell command to run this test:

	go test -v ./tencentcloud_ssl_test.go -args \
	--TENCENTCLOUDSSL_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDSSL_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDSSL_SECRETID="your-secret-id" \
	--TENCENTCLOUDSSL_SECRETKEY="your-secret-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			SecretId:  fSecretId,
			SecretKey: fSecretKey,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
