//go:build tester

package email_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/email"
)

var (
	fp               = tester.InitArgs("EMAIL_")
	fTestCertPath    string
	fTestKeyPath     string
	fSmtpHost        string
	fSmtpPort        int64
	fSmtpTls         bool
	fUsername        string
	fPassword        string
	fSenderAddress   string
	fSenderName      string
	fReceiverAddress string
	fPfxPassword     string
	fJksAlias        string
	fJksKeypass      string
	fJksStorepass    string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSmtpHost, "SMTPHOST")
	fp.DefineInt64(&fSmtpPort, "SMTPPORT")
	fp.DefineBool(&fSmtpTls, "SMTPTLS")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
	fp.DefineString(&fSenderAddress, "SENDERADDRESS")
	fp.DefineString(&fSenderName, "SENDERNAME")
	fp.DefineString(&fReceiverAddress, "RECEIVERADDRESS")
	fp.DefineString(&fPfxPassword, "PFXPASSWORD")
	fp.DefineString(&fJksAlias, "JKSALIAS")
	fp.DefineString(&fJksKeypass, "JKSKEYPASS")
	fp.DefineString(&fJksStorepass, "JKSSTOREPASS")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./email_test.go -args 	--EMAIL_TESTCERTPATH="/path/to/your-test-cert.pem" 	--EMAIL_TESTKEYPATH="/path/to/your-test-key.pem" 	--EMAIL_SMTPHOST="smtp.example.com" 	--EMAIL_SMTPPORT=465 	--EMAIL_SMTPTLS=true 	--EMAIL_USERNAME="USER" 	--EMAIL_PASSWORD="PASS" 	--EMAIL_SENDERADDRESS="sender@example.com" 	--EMAIL_SENDERNAME="Certimate" 	--EMAIL_RECEIVERADDRESS="receiver@example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_PEM", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SmtpHost:                 fSmtpHost,
			SmtpPort:                 int32(fSmtpPort),
			SmtpTls:                  fSmtpTls,
			Username:                 fUsername,
			Password:                 fPassword,
			SenderAddress:            fSenderAddress,
			SenderName:               fSenderName,
			ReceiverAddress:          fReceiverAddress,
			FileFormat:               impl.FILE_FORMAT_PEM,
			AllowInsecureConnections: true,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_PFX", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SmtpHost:                 fSmtpHost,
			SmtpPort:                 int32(fSmtpPort),
			SmtpTls:                  fSmtpTls,
			Username:                 fUsername,
			Password:                 fPassword,
			SenderAddress:            fSenderAddress,
			SenderName:               fSenderName,
			ReceiverAddress:          fReceiverAddress,
			FileFormat:               impl.FILE_FORMAT_PFX,
			PfxPassword:              fPfxPassword,
			AllowInsecureConnections: true,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_JKS", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SmtpHost:                 fSmtpHost,
			SmtpPort:                 int32(fSmtpPort),
			SmtpTls:                  fSmtpTls,
			Username:                 fUsername,
			Password:                 fPassword,
			SenderAddress:            fSenderAddress,
			SenderName:               fSenderName,
			ReceiverAddress:          fReceiverAddress,
			FileFormat:               impl.FILE_FORMAT_JKS,
			JksAlias:                 fJksAlias,
			JksKeypass:               fJksKeypass,
			JksStorepass:             fJksStorepass,
			AllowInsecureConnections: true,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
