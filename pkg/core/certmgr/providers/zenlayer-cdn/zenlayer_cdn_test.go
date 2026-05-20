package zenlayercdn_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/zenlayer-cdn"
)

var (
	fp                 = tester.Args("ZENLAYERCDN_")
	fTestCertPath      string
	fTestKeyPath       string
	fAccessKeyId       string
	fAccessKeyPassword string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeyPassword, "ACCESSKEYPASSWORD")
}

/*
Shell command to run this test:

	go test -v ./zenlayer_cdn_test.go -args \
	--ZENLAYERCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ZENLAYERCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ZENLAYERCDN_ACCESSKEYID="your-access-key-id" \
	--ZENLAYERCDN_ACCESSKEYPASSWORD="your-secret-access-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKeyId:       fAccessKeyId,
			AccessKeyPassword: fAccessKeyPassword,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
