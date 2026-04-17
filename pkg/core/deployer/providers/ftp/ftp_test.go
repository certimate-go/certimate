package ftp_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/ftp"
)

var (
	fInputCertPath  string
	fInputKeyPath   string
	fFtpHost        string
	fFtpPort        int64
	fFtpUsername    string
	fFtpPassword    string
	fOutputCertPath string
	fOutputKeyPath  string
)

func init() {
	argsPrefix := "FTP_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fFtpHost, argsPrefix+"FTPHOST", "", "")
	flag.Int64Var(&fFtpPort, argsPrefix+"FTPPORT", 0, "")
	flag.StringVar(&fFtpUsername, argsPrefix+"FTPUSERNAME", "", "")
	flag.StringVar(&fFtpPassword, argsPrefix+"FTPPASSWORD", "", "")
	flag.StringVar(&fOutputCertPath, argsPrefix+"OUTPUTCERTPATH", "", "")
	flag.StringVar(&fOutputKeyPath, argsPrefix+"OUTPUTKEYPATH", "", "")
}

/*
Shell command to run this test:

	go test -v ./ftp_test.go -args \
	--FTP_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--FTP_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--FTP_FTPHOST="localhost" \
	--FTP_FTPPORT=21 \
	--FTP_FTPUSERNAME="USER" \
	--FTP_FTPPASSWORD="PASS" \
	--FTP_OUTPUTCERTPATH="/path/to/your-output-cert.pem" \
	--FTP_OUTPUTKEYPATH="/path/to/your-output-key.pem"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("FTPHOST: %v", fFtpHost),
			fmt.Sprintf("FTPPORT: %v", fFtpPort),
			fmt.Sprintf("FTPUSERNAME: %v", fFtpUsername),
			fmt.Sprintf("FTPPASSWORD: %v", fFtpPassword),
			fmt.Sprintf("OUTPUTCERTPATH: %v", fOutputCertPath),
			fmt.Sprintf("OUTPUTKEYPATH: %v", fOutputKeyPath),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			FtpHost:        fFtpHost,
			FtpPort:        int32(fFtpPort),
			FtpUsername:    fFtpUsername,
			FtpPassword:    fFtpPassword,
			OutputFormat:   provider.OUTPUT_FORMAT_PEM,
			OutputCertPath: fOutputCertPath,
			OutputKeyPath:  fOutputKeyPath,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
		}

		fInputCertData, _ := os.ReadFile(fInputCertPath)
		fInputKeyData, _ := os.ReadFile(fInputKeyPath)
		res, err := provider.Deploy(context.Background(), string(fInputCertData), string(fInputKeyData))
		if err != nil {
			t.Errorf("err: %+v", err)
		}

		t.Logf("ok: %v", res)
	})
}
