//go:build tester

package tencentcloudeo_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-eo"
)

var (
	fp            = tester.InitArgs("TENCENTCLOUDEO_")
	fTestCertPath string
	fTestKeyPath  string
	fSecretId     string
	fSecretKey    string
	fZoneId       string
	fDomains      string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fZoneId, "ZONEID")
	fp.DefineString(&fDomains, "DOMAINS")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./tencentcloud_eo_test.go -args \
	--TENCENTCLOUDEO_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDEO_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDEO_SECRETID="your-secret-id" \
	--TENCENTCLOUDEO_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDEO_ZONEID="your-zone-id" \
	--TENCENTCLOUDEO_DOMAINS="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:           fSecretId,
			SecretKey:          fSecretKey,
			ZoneId:             fZoneId,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domains:            strings.Split(fDomains, ";"),
			EnableMultipleSSL:  true,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
