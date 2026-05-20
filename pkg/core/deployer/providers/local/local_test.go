package local_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/local"
)

var (
	fp              = tester.Args("LOCAL_")
	fTestCertPath   string
	fTestKeyPath    string
	fFilePathForCrt string
	fFilePathForKey string
	fPfxPassword    string
	fJksAlias       string
	fJksKeypass     string
	fJksStorepass   string
	fShellEnv       string
	fPreCommand     string
	fPostCommand    string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fFilePathForCrt, "OUTPUTCERTPATH")
	fp.DefineString(&fFilePathForKey, "OUTPUTKEYPATH")
	fp.DefineString(&fPfxPassword, "PFXPASSWORD")
	fp.DefineString(&fJksAlias, "JKSALIAS")
	fp.DefineString(&fJksKeypass, "JKSKEYPASS")
	fp.DefineString(&fJksStorepass, "JKSSTOREPASS")
	fp.DefineString(&fShellEnv, "SHELLENV")
	fp.DefineString(&fPreCommand, "PRECOMMAND")
	fp.DefineString(&fPostCommand, "POSTCOMMAND")
}

/*
Shell command to run this test:

	go test -v ./local_test.go -args \
	--LOCAL_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--LOCAL_TESTKEYPATH="/path/to/your-test-key.pem" \
	--LOCAL_OUTPUTCERTPATH="/path/to/your-output-cert" \
	--LOCAL_OUTPUTKEYPATH="/path/to/your-output-key" \
	--LOCAL_PFXPASSWORD="your-pfx-password" \
	--LOCAL_JKSALIAS="your-jks-alias" \
	--LOCAL_JKSKEYPASS="your-jks-keypass" \
	--LOCAL_JKSSTOREPASS="your-jks-storepass" \
	--LOCAL_SHELLENV="sh" \
	--LOCAL_PRECOMMAND="echo 'hello world'" \
	--LOCAL_POSTCOMMAND="echo 'bye-bye world'"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_PEM", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			FileFormat:     impl.FILE_FORMAT_PEM,
			FilePathForCrt: fFilePathForCrt + ".pem",
			FilePathForKey: fFilePathForKey + ".pem",
			ShellEnv:       fShellEnv,
			PreCommand:     fPreCommand,
			PostCommand:    fPostCommand,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_PFX", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			FileFormat:     impl.FILE_FORMAT_PFX,
			FilePathForCrt: fFilePathForCrt + ".pfx",
			PfxPassword:    fPfxPassword,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_JKS", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			FileFormat:     impl.FILE_FORMAT_JKS,
			FilePathForCrt: fFilePathForCrt + ".jks",
			JksAlias:       fJksAlias,
			JksKeypass:     fJksKeypass,
			JksStorepass:   fJksStorepass,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
