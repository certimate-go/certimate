//go:build tester

package huaweiibmc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweiibmc"
)

var (
	fp            = tester.InitArgs("HUAWEIIBMC_")
	fTestCertPath string
	fTestKeyPath  string
	fHost         string
	fUsername     string
	fPassword     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fHost, "HOST")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./huaweiibmc_test.go -args \
	--HUAWEIIBMC_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--HUAWEIIBMC_TESTKEYPATH="/path/to/your-test-key.pem" \
	--HUAWEIIBMC_HOST="localhost" \
	--HUAWEIIBMC_USERNAME="admin" \
	--HUAWEIIBMC_PASSWORD="password"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			Host:        fHost,
			Username:    fUsername,
			Password:    fPassword,
			AutoRestart: true,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
