package volcenginelive_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/volcengine-live"
)

var (
	fp               = tester.Args("VOLCENGINELIVE_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
}

/*
Shell command to run this test:

	go test -v ./volcengine_live_test.go -args \
	--VOLCENGINELIVE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--VOLCENGINELIVE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--VOLCENGINELIVE_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINELIVE_ACCESSKEYSECRET="your-access-key-secret"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
