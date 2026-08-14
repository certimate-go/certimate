//go:build tester

package baishancdn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/baishan-cdn"
)

var (
	fp            = tester.InitArgs("BAISHANCDN_")
	fTestCertPath string
	fTestKeyPath  string
	fApiToken     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiToken, "APITOKEN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./baishan_cdn_test.go -args \
	--BAISHANCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BAISHANCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BAISHANCDN_APITOKEN="your-api-token"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			ApiToken: fApiToken,
		})
		require.NoError(t, err)

		tester.Upload(t, provider, tester.UploadInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
