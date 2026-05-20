package ucloudulb_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ucloud-ulb"
)

var (
	fp            = tester.Args("UCLOUDULB_")
	fTestCertPath string
	fTestKeyPath  string
	fPrivateKey   string
	fPublicKey    string
	fRegion       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fPrivateKey, "PRIVATEKEY")
	fp.DefineString(&fPublicKey, "PUBLICKEY")
	fp.DefineString(&fRegion, "REGION")
}

/*
Shell command to run this test:

	go test -v ./ucloud_ulb_test.go -args \
	--UCLOUDULB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UCLOUDULB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UCLOUDULB_PRIVATEKEY="your-private-key" \
	--UCLOUDULB_PUBLICKEY="your-public-key" \
	--UCLOUDULB_REGION="cn-bj2"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			PrivateKey: fPrivateKey,
			PublicKey:  fPublicKey,
			Region:     fRegion,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
