package ctcccloudfaas_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ctcccloud-faas"
)

var (
	fp               = tester.Args("CTCCCLOUDFAAS_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegionId        string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegionId, "REGIONID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./ctcccloud_faas_test.go -args \
	--CTCCCLOUDFAAS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--CTCCCLOUDFAAS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--CTCCCLOUDFAAS_ACCESSKEYID="your-access-key-id" \
	--CTCCCLOUDFAAS_SECRETACCESSKEY="your-secret-access-key" \
	--CTCCCLOUDFAAS_REGIONID="your-region-id" \
	--CTCCCLOUDFAAS_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			RegionId:        fRegionId,
			Domain:          fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
