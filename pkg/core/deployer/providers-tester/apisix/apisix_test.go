//go:build tester

package apisix_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/apisix"
)

var (
	fp             = tester.InitArgs("APISIX_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fApiKey        string
	fCertificateId string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiKey, "APIKEY")
	fp.DefineString(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./apisix_test.go -args \
	--APISIX_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--APISIX_TESTKEYPATH="/path/to/your-test-key.pem" \
	--APISIX_SERVERURL="http://127.0.0.1:9080" \
	--APISIX_APIKEY="your-api-key" \
	--APISIX_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiKey:                   fApiKey,
			AllowInsecureConnections: true,
			DeployTarget:             impl.DEPLOY_TARGET_CERTIFICATE,
			CertificateId:            fCertificateId,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
