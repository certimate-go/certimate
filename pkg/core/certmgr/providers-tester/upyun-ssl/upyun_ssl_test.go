//go:build tester

package upyunssl_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/upyun-ssl"
)

var (
	fp            = tester.InitArgs("UPYUNSSL_")
	fTestCertPath string
	fTestKeyPath  string
	fUsername     string
	fPassword     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./upyun_ssl_test.go -args \
	--UPYUNSSL_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UPYUNSSL_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UPYUNSSL_USERNAME="your-username" \
	--UPYUNSSL_PASSWORD="your-password"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			Username: fUsername,
			Password: fPassword,
		})
		require.NoError(t, err)

		tester.Upload(t, provider, tester.UploadInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
