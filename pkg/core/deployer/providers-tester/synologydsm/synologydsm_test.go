//go:build tester

package synologydsm_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/synologydsm"
)

var (
	fp                   = tester.InitArgs("SYNOLOGYDSM_")
	fTestCertPath        string
	fTestKeyPath         string
	fServerUrl           string
	fUsername            string
	fPassword            string
	fTotpSecret          string
	fCertificateIdOrDesc string
	fIsDefault           bool
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
	fp.DefineString(&fTotpSecret, "TOTPSECRET")
	fp.DefineString(&fCertificateIdOrDesc, "CERTIFICATEIDORDESC")
	fp.DefineBool(&fIsDefault, "ISDEFAULT")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./synology_dsm_test.go -args \
	--SYNOLOGYDSM_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--SYNOLOGYDSM_TESTKEYPATH="/path/to/your-test-key.pem" \
	--SYNOLOGYDSM_SERVERURL="http://127.0.0.1:5000/" \
	--SYNOLOGYDSM_USERNAME="admin" \
	--SYNOLOGYDSM_PASSWORD="password" \
	--SYNOLOGYDSM_CERTIFICATEIDORDESC="your-certificate-id-or-desc" \
	--SYNOLOGYDSM_ISDEFAULT=true
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                  fServerUrl,
			Username:                   fUsername,
			Password:                   fPassword,
			TotpSecret:                 fTotpSecret,
			AllowInsecureConnections:   true,
			CertificateIdOrDescription: fCertificateIdOrDesc,
			IsDefault:                  fIsDefault,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
