package tencentcloudeomakers_test

import (
	"strings"
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-eomakers"
)

var (
	fp               = tester.Args("TENCENTCLOUDEOMAKERS_")
	fTestCertPath    string
	fTestKeyPath     string
	fSecretId        string
	fSecretKey       string
	fMakersApiToken  string
	fMakersProjectId string
	fDomains         string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fMakersApiToken, "MAKERSAPITOKEN")
	fp.DefineString(&fMakersProjectId, "MAKERSPROJECTID")
	fp.DefineString(&fDomains, "DOMAINS")
}

/*
Shell command to run this test:

	go test -v ./tencentcloud_eomakers_test.go -args \
	--TENCENTCLOUDEOMAKERS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDEOMAKERS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDEOMAKERS_SECRETID="your-secret-id" \
	--TENCENTCLOUDEOMAKERS_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDEOMAKERS_MAKERSAPITOKEN="your-makers-api-token" \
	--TENCENTCLOUDEOMAKERS_MAKERSPROJECTID="your-makers-project-id" \
	--TENCENTCLOUDEOMAKERS_DOMAINS="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:           fSecretId,
			SecretKey:          fSecretKey,
			MakersApiToken:     fMakersApiToken,
			MakersProjectId:    fMakersProjectId,
			DomainMatchPattern: impl.DomainMatchPatternExact,
			Domains:            strings.Split(fDomains, ";"),
			EnableMultipleSSL:  true,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
