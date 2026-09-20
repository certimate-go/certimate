//go:build tester

package autossl_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certsource/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certsource/providers/autossl"
)

var (
	fp         = tester.InitArgs("AUTOSSL_")
	fDomain    string
	fCertId    string
	fAccessKey string
	fSecretKey string
)

func init() {
	fp.DefineString(&fDomain, "DOMAIN")
	fp.DefineString(&fCertId, "CERTID")
	fp.DefineString(&fAccessKey, "ACCESSKEY")
	fp.DefineString(&fSecretKey, "SECRETKEY")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./autossl_test.go -args \
		--AUTOSSL_DOMAIN="example.com" \
		--AUTOSSL_ACCESSKEY="your-access-key" \
		--AUTOSSL_SECRETKEY="your-secret-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Fetch", func(t *testing.T) {
		provider, err := impl.NewCertsource(&impl.CertsourceConfig{
			AccessKey: fAccessKey,
			SecretKey: fSecretKey,
		})
		require.NoError(t, err)

		tester.Fetch(t, provider, tester.FetchInput{Domain: fDomain, CertId: fCertId})
	})
}
