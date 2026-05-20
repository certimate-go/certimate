package aliyunesasaas_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-esa-saas"
)

var (
	fp               = tester.Args("ALIYUNESASAAS_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fRegion          string
	fSiteId          int64
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineInt64(&fSiteId, "SITEID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./aliyun_esasaas_test.go -args \
	--ALIYUNESASAAS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ALIYUNESASAAS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ALIYUNESASAAS_ACCESSKEYID="your-access-key-id" \
	--ALIYUNESASAAS_ACCESSKEYSECRET="your-access-key-secret" \
	--ALIYUNESASAAS_REGION="cn-hangzhou" \
	--ALIYUNESASAAS_SITEID="your-esa-site-id"\
	--ALIYUNESASAAS_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			AccessKeySecret:    fAccessKeySecret,
			Region:             fRegion,
			SiteId:             fSiteId,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
