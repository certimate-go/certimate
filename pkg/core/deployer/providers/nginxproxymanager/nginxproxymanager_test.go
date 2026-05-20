package nginxproxymanager_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/nginxproxymanager"
)

var (
	fp            = tester.Args("NGINXPROXYMANAGER_")
	fTestCertPath string
	fTestKeyPath  string
	fServerUrl    string
	fUsername     string
	fPassword     string
	fHostType     string
	fHostId       int64
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fHostType, "HOSTTYPE")
	fp.DefineInt64(&fHostId, "HOSTID")
}

/*
Shell command to run this test:

	go test -v ./nginxproxymanager_test.go -args \
	--NGINXPROXYMANAGER_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--NGINXPROXYMANAGER_TESTKEYPATH="/path/to/your-test-key.pem" \
	--NGINXPROXYMANAGER_SERVERURL="http://127.0.0.1:20410" \
	--NGINXPROXYMANAGER_USERNAME="your-username" \
	--NGINXPROXYMANAGER_PASSWORD="your-password" \
	--NGINXPROXYMANAGER_HOSTTYPE="proxy" \
	--NGINXPROXYMANAGER_HOSTID="your-host-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToHost", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			AuthMethod:               impl.AUTH_METHOD_PASSWORD,
			Username:                 fUsername,
			Password:                 fPassword,
			AllowInsecureConnections: true,
			ResourceType:             impl.RESOURCE_TYPE_HOST,
			HostType:                 fHostType,
			HostMatchPattern:         impl.HOST_MATCH_PATTERN_SPECIFIED,
			HostId:                   fHostId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
