package ucloudussl_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ucloud-ussl"
)

var (
	fp            = tester.Args("UCLOUDUSSL_")
	fTestCertPath string
	fTestKeyPath  string
	fPrivateKey   string
	fPublicKey    string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fPrivateKey, "PRIVATEKEY")
	fp.DefineString(&fPublicKey, "PUBLICKEY")
}

/*
Shell command to run this test:

	go test -v ./ucloud_ussl_test.go -args \
	--UCLOUDUSSL_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UCLOUDUSSL_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UCLOUDUSSL_PRIVATEKEY="your-private-key" \
	--UCLOUDUSSL_PUBLICKEY="your-public-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			PrivateKey: fPrivateKey,
			PublicKey:  fPublicKey,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
