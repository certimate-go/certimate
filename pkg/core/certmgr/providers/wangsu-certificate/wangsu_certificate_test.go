package wangsucertificate_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/wangsu-certificate"
)

var (
	fp               = tester.Args("WANGSUCERTIFICATE_")
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

	go test -v ./wangsu_certificate_test.go -args \
	--WANGSUCERTIFICATE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--WANGSUCERTIFICATE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--WANGSUCERTIFICATE_ACCESSKEYID="your-access-key-id" \
	--WANGSUCERTIFICATE_ACCESSKEYSECRET="your-access-key-secret"
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
