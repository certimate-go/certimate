package dogecloud_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/dogecloud"
)

var (
	fp            = tester.Args("DOGECLOUD_")
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

	go test -v ./dogecloud_test.go -args \
	--DOGECLOUD_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--DOGECLOUD_TESTKEYPATH="/path/to/your-test-key.pem" \
	--DOGECLOUD_ACCESSKEY="your-access-key" \
	--DOGECLOUD_SECRETKEY="your-secret-key"
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
