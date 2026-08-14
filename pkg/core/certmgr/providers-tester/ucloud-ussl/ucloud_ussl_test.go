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

	go test -tags=tester -v ./ucloud_ussl_test.go -args \
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
		require.NoError(t, err)

		tester.Upload(t, provider, tester.UploadInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
