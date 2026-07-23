package f5bigip_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/f5bigip"
)

var (
	fp             = tester.Args("F5BIGIP_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fUsername      string
	fPassword      string
	fPartition     string
	fClientSSLName string
	fDeployTarget  string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
	fp.DefineString(&fPartition, "PARTITION")
	fp.DefineString(&fClientSSLName, "CLIENTSSLNAME")
	fp.DefineString(&fDeployTarget, "DEPLOYTARGET")
}

func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			Username:                 fUsername,
			Password:                 fPassword,
			AllowInsecureConnections: true,
			DeployTarget:             fDeployTarget,
			Partition:                fPartition,
			ClientSSLProfileName:     fClientSSLName,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
