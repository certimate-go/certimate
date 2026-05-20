package ssh_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ssh"
)

var (
	fp              = tester.Args("SSH_")
	fTestCertPath   string
	fTestKeyPath    string
	fSshHost        string
	fSshPort        int64
	fSshUsername    string
	fSshPassword    string
	fOutputCertPath string
	fOutputKeyPath  string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSshHost, "SSHHOST")
	fp.DefineInt64(&fSshPort, "SSHPORT")
	fp.DefineString(&fSshUsername, "SSHUSERNAME")
	fp.DefineString(&fSshPassword, "SSHPASSWORD")
	fp.DefineString(&fOutputCertPath, "OUTPUTCERTPATH")
	fp.DefineString(&fOutputKeyPath, "OUTPUTKEYPATH")
}

/*
Shell command to run this test:

	go test -v ./ssh_test.go -args \
	--SSH_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--SSH_TESTKEYPATH="/path/to/your-test-key.pem" \
	--SSH_SSHHOST="localhost" \
	--SSH_SSHPORT=22 \
	--SSH_SSHUSERNAME="root" \
	--SSH_SSHPASSWORD="password" \
	--SSH_OUTPUTCERTPATH="/path/to/your-output-cert.pem" \
	--SSH_OUTPUTKEYPATH="/path/to/your-output-key.pem"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_PEM", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerConfig: impl.ServerConfig{
				SshHost:     fSshHost,
				SshPort:     int32(fSshPort),
				SshUsername: fSshUsername,
				SshPassword: fSshPassword,
			},
			OutputFormat:   impl.OUTPUT_FORMAT_PEM,
			OutputCertPath: fOutputCertPath + ".pem",
			OutputKeyPath:  fOutputKeyPath + ".pem",
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
