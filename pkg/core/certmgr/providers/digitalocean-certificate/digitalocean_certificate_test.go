package digitaloceancertificate_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/digitalocean-certificate"
	it "github.com/certimate-go/certimate/pkg/core/certmgr/testing"
)

var (
	fp            = it.Args("DIGITALOCEANCERTIFICATE_")
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
		require.NoError(t, err)

		it.TestUpload(t, provider, it.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
