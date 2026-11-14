//go:build tester

package ucloudussl_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ucloud-ussl"
)

var (
	fp            = tester.InitArgs("UCLOUDUSSL_")
	fTestCertPath string
	fTestKeyPath  string
	fPublicKey    string
	fPrivateKey   string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fPublicKey, "PUBLICKEY")
	fp.DefineString(&fPrivateKey, "PRIVATEKEY")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./ucloud_ussl_test.go -args \
	--UCLOUDUSSL_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UCLOUDUSSL_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UCLOUDUSSL_PUBLICKEY="your-public-key" \
	--UCLOUDUSSL_PRIVATEKEY="your-private-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			PublicKey:  fPublicKey,
			PrivateKey: fPrivateKey,
		})
		require.NoError(t, err)

		tester.Upload(t, provider, tester.UploadInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
