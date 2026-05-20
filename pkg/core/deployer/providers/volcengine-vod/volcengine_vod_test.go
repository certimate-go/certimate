package volcenginevod_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-vod"
)

var (
	fp               = tester.Args("VOLCENGINEVOD_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fSpaceName       string
	fDomainType      string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fSpaceName, "SPACENAME")
	fp.DefineString(&fDomainType, "DOMAINTYPE")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./volcengine_vod_test.go -args \
	--VOLCENGINEVOD_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--VOLCENGINEVOD_TESTKEYPATH="/path/to/your-test-key.pem" \
	--VOLCENGINEVOD_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINEVOD_ACCESSKEYSECRET="your-access-key-secret" \
	--VOLCENGINEVOD_SPACENAME="vod-space-name" \
	--VOLCENGINEVOD_DOMAINTYPE="play" \
	--VOLCENGINEVOD_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			AccessKeySecret:    fAccessKeySecret,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			SpaceName:          fSpaceName,
			DomainType:         fDomainType,
			Domain:             fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
