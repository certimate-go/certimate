//go:build tester

package volcenginevod_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-vod"
)

var (
	fp               = tester.InitArgs("VOLCENGINEVOD_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fSpaceName       string
	fDomainType      string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fSpaceName, "SPACENAME")
	fp.DefineString(&fDomainType, "DOMAINTYPE")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./volcengine_vod_test.go -args \
	--VOLCENGINEVOD_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--VOLCENGINEVOD_TESTKEYPATH="/path/to/your-test-key.pem" \
	--VOLCENGINEVOD_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINEVOD_SECRETACCESSKEY="your-secret-access-key" \
	--VOLCENGINEVOD_SPACENAME="vod-space-name" \
	--VOLCENGINEVOD_DOMAINTYPE="play" \
	--VOLCENGINEVOD_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			SecretAccessKey:    fSecretAccessKey,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			SpaceName:          fSpaceName,
			DomainType:         fDomainType,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
