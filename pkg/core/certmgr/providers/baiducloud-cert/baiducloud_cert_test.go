package baiducloudcert_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/baiducloud-cert"
)

var (
	fp               = tester.Args("BAIDUCLOUDCERT_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
}

/*
Shell command to run this test:

	go test -v ./baiducloud_cert_test.go -args \
	--BAIDUCLOUDCERT_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BAIDUCLOUDCERT_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BAIDUCLOUDCERT_ACCESSKEYID="your-access-key-id" \
	--BAIDUCLOUDCERT_SECRETACCESSKEY="your-access-key-secret"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
