//go:build tester

package kong_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/kong"
)

var (
	fp             = tester.InitArgs("KONG_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fApiToken      string
	fCertificateId string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineString(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./kong_test.go -args \
	--KONG_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--KONG_TESTKEYPATH="/path/to/your-test-key.pem" \
	--KONG_SERVERURL="http://127.0.0.1:9080" \
	--KONG_APITOKEN="your-admin-token" \
	--KONG_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiToken:                 fApiToken,
			AllowInsecureConnections: true,
			DeployTarget:             impl.DEPLOY_TARGET_CERTIFICATE,
			CertificateId:            fCertificateId,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
