package dokploy_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/dokploy"
)

var (
	fp            = tester.Args("DOKPLOY_")
	fTestCertPath string
	fTestKeyPath  string
	fServerUrl    string
	fApiKey       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiKey, "APIKEY")
}

/*
Shell command to run this test:

	go test -v ./dokploy_test.go -args \
	--DOKPLOY_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--DOKPLOY_TESTKEYPATH="/path/to/your-test-key.pem" \
	--DOKPLOY_SERVERURL="http://127.0.0.1:3000" \
	--DOKPLOY_APIKEY="your-api-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			ServerUrl:                fServerUrl,
			ApiKey:                   fApiKey,
			AllowInsecureConnections: true,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
