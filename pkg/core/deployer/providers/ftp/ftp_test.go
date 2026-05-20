package ftp_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ftp"
)

var (
	fp              = tester.Args("FTP_")
	fTestCertPath   string
	fTestKeyPath    string
	fFtpHost        string
	fFtpPort        int64
	fFtpUsername    string
	fFtpPassword    string
	fOutputCertPath string
	fOutputKeyPath  string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fFtpHost, "FTPHOST")
	fp.DefineInt64(&fFtpPort, "FTPPORT")
	fp.DefineString(&fFtpUsername, "FTPUSERNAME")
	fp.DefineString(&fFtpPassword, "FTPPASSWORD")
	fp.DefineString(&fOutputCertPath, "OUTPUTCERTPATH")
	fp.DefineString(&fOutputKeyPath, "OUTPUTKEYPATH")
}

/*
Shell command to run this test:

	go test -v ./ftp_test.go -args \
	--FTP_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--FTP_TESTKEYPATH="/path/to/your-test-key.pem" \
	--FTP_FTPHOST="localhost" \
	--FTP_FTPPORT=21 \
	--FTP_FTPUSERNAME="USER" \
	--FTP_FTPPASSWORD="PASS" \
	--FTP_OUTPUTCERTPATH="/path/to/your-output-cert.pem" \
	--FTP_OUTPUTKEYPATH="/path/to/your-output-key.pem"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_PEM", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			FtpHost:        fFtpHost,
			FtpPort:        int32(fFtpPort),
			FtpUsername:    fFtpUsername,
			FtpPassword:    fFtpPassword,
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
