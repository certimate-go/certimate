package nginxproxymanager_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/nginxproxymanager"
	it "github.com/certimate-go/certimate/pkg/core/certmgr/testing"
)

var (
	fp            = it.Args("NGINXPROXYMANAGER_")
	fTestCertPath string
	fTestKeyPath  string
	fServerUrl    string
	fUsername     string
	fPassword     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
}

/*
Shell command to run this test:

	go test -v ./nginxproxymanager_test.go -args \
	--NGINXPROXYMANAGER_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--NGINXPROXYMANAGER_TESTKEYPATH="/path/to/your-test-key.pem" \
	--NGINXPROXYMANAGER_SERVERURL="http://127.0.0.1:81" \
	--NGINXPROXYMANAGER_USERNAME="your-username" \
	--NGINXPROXYMANAGER_PASSWORD="your-password"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			ServerUrl:                fServerUrl,
			AuthMethod:               impl.AUTH_METHOD_PASSWORD,
			Username:                 fUsername,
			Password:                 fPassword,
			AllowInsecureConnections: true,
		})
		require.NoError(t, err)

		it.TestUpload(t, provider, it.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
