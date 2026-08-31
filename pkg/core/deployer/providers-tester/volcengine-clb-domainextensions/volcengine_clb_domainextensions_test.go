//go:build tester

package volcengineclbdomainextensions_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-clb-domainextensions"
)

var (
	fp               = tester.InitArgs("VOLCENGINECLBDOMAINEXTENSIONS_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fProjectName     string
	fRegion          string
	fListenerId      string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fProjectName, "PROJECTNAME")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fListenerId, "LISTENERID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./volcengine_clb_domainextensions_test.go -args \
	--VOLCENGINECLBDOMAINEXTENSIONS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--VOLCENGINECLBDOMAINEXTENSIONS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--VOLCENGINECLBDOMAINEXTENSIONS_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINECLBDOMAINEXTENSIONS_SECRETACCESSKEY="your-secret-access-key" \
	--VOLCENGINECLBDOMAINEXTENSIONS_PROJECTNAME="your-project-name" \
	--VOLCENGINECLBDOMAINEXTENSIONS_REGION="cn-beijing" \
	--VOLCENGINECLBDOMAINEXTENSIONS_LISTENERID="your-listener-id" \
	--VOLCENGINECLBDOMAINEXTENSIONS_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	provider, err := impl.NewDeployer(&impl.DeployerConfig{
		AccessKeyId:     fAccessKeyId,
		SecretAccessKey: fSecretAccessKey,
		ProjectName:     fProjectName,
		Region:          fRegion,
		ListenerId:      fListenerId,
		Domain:          fDomain,
	})
	require.NoError(t, err)

	tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
}
