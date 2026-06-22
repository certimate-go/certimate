package digitaloceancertificate_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/digitalocean-certificate"
)

var (
	fp            = tester.Args("DIGITALOCEANCERTIFICATE_")
	fTestCertPath string
	fTestKeyPath  string
	fAccessToken  string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessToken, "ACCESSTOKEN")
}

/*
Shell command to run this test:

	go test -v ./digitalocean_certificate_test.go -args \
	--DIGITALOCEANCERTIFICATE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--DIGITALOCEANCERTIFICATE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--DIGITALOCEANCERTIFICATE_ACCESSTOKEN="your-access-token"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessToken: fAccessToken,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
