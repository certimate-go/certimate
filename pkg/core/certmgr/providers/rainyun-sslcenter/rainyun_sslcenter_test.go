package rainyunsslcenter_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/rainyun-sslcenter"
)

var (
	fp            = tester.Args("RAINYUNSSLCENTER_")
	fTestCertPath string
	fTestKeyPath  string
	fApiKey       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiKey, "APIKEY")
}

/*
Shell command to run this test:

	go test -v ./rainyun_sslcenter_test.go -args \
	--RAINYUNSSLCENTER_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--RAINYUNSSLCENTER_TESTKEYPATH="/path/to/your-test-key.pem" \
	--RAINYUNSSLCENTER_APIKEY="your-api-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			ApiKey: fApiKey,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
