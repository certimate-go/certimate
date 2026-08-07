package vercel_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/vercel"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp              = it.Args("VERCEL_")
	fTestCertPath   string
	fTestKeyPath    string
	fApiAccessToken string
	fTeamId         string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiAccessToken, "APIACCESSTOKEN")
	fp.DefineString(&fTeamId, "TEAMID")
}

/*
Shell command to run this test:

	go test -v ./vercel_test.go -args \
	--VERCEL_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--VERCEL_TESTKEYPATH="/path/to/your-test-key.pem" \
	--VERCEL_APIACCESSTOKEN="your-api-access-token" \
	--VERCEL_TEAMID="your-team-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ApiAccessToken: fApiAccessToken,
			TeamId:         fTeamId,
		})
		require.NoError(t, err)

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
