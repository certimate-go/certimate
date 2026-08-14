//go:build tester

package onepanelssl_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/1panel"
)

var (
	fp            = tester.InitArgs("1PANEL_")
	fTestCertPath string
	fTestKeyPath  string
	fServerUrl    string
	fApiVersion   string
	fApiKey       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiVersion, "APIVERSION", "v1")
	fp.DefineString(&fApiKey, "APIKEY")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./1panel_test.go -args \
	--1PANEL_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--1PANEL_TESTKEYPATH="/path/to/your-test-key.pem" \
	--1PANEL_SERVERURL="http://127.0.0.1:20410" \
	--1PANEL_APIVERSION="v1" \
	--1PANEL_APIKEY="your-api-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			ServerUrl:  fServerUrl,
			ApiVersion: fApiVersion,
			ApiKey:     fApiKey,
		})
		require.NoError(t, err)

		tester.Upload(t, provider, tester.UploadInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
