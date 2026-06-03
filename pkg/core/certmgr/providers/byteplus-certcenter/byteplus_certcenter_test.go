package bytepluscertcenter_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/byteplus-certcenter"
)

var (
	fp            = tester.Args("BYTEPLUSCERTCENTER_")
	fTestCertPath string
	fTestKeyPath  string
	fAccessKey    string
	fSecretKey    string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKey, "ACCESSKEY")
	fp.DefineString(&fSecretKey, "SECRETKEY")
}

/*
Shell command to run this test:

	go test -v ./byteplus_certcenter_test.go -args \
	--BYTEPLUSCERTCENTER_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BYTEPLUSCERTCENTER_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BYTEPLUSCERTCENTER_ACCESSKEY="your-access-key" \
	--BYTEPLUSCERTCENTER_SECRETKEY="your-secret-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKey: fAccessKey,
			SecretKey: fSecretKey,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
