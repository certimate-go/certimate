package zenlayerga_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/zenlayer-ga"
)

var (
	fp                 = tester.Args("ZENLAYERGA_")
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

	go test -v ./zenlayer_ga_test.go -args \
	--ZENLAYERGA_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ZENLAYERGA_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ZENLAYERGA_ACCESSKEYID="your-access-key-id" \
	--ZENLAYERGA_ACCESSKEYPASSWORD="your-secret-access-key"
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
