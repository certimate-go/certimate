package synologydsm_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/synologydsm"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fHostname      string
	fPort          int
	fScheme        string
	fUsername      string
	fPassword      string
	fTotpSecret    string
	fInsecure      bool
	fCertName      string
	fCertID        string
	fIsDefault     bool
)

func init() {
	argsPrefix := "SYNOLOGYDSM_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fHostname, argsPrefix+"HOSTNAME", "", "")
	flag.IntVar(&fPort, argsPrefix+"PORT", 5001, "")
	flag.StringVar(&fScheme, argsPrefix+"SCHEME", "https", "")
	flag.StringVar(&fUsername, argsPrefix+"USERNAME", "", "")
	flag.StringVar(&fPassword, argsPrefix+"PASSWORD", "", "")
	flag.StringVar(&fTotpSecret, argsPrefix+"TOTPSECRET", "", "")
	flag.BoolVar(&fInsecure, argsPrefix+"INSECURE", true, "")
	flag.StringVar(&fCertName, argsPrefix+"CERTNAME", "", "")
	flag.StringVar(&fCertID, argsPrefix+"CERTID", "", "")
	flag.BoolVar(&fIsDefault, argsPrefix+"ISDEFAULT", false, "")
}

/*
Shell command to run this test:

	go test -v ./synology_dsm_test.go -args \
	--SYNOLOGYDSM_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--SYNOLOGYDSM_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--SYNOLOGYDSM_HOSTNAME="192.168.1.100" \
	--SYNOLOGYDSM_PORT="5001" \
	--SYNOLOGYDSM_SCHEME="https" \
	--SYNOLOGYDSM_USERNAME="admin" \
	--SYNOLOGYDSM_PASSWORD="your-password" \
	--SYNOLOGYDSM_INSECURE="true" \
	--SYNOLOGYDSM_CERTNAME="test-cert"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("HOSTNAME: %v", fHostname),
			fmt.Sprintf("PORT: %v", fPort),
			fmt.Sprintf("SCHEME: %v", fScheme),
			fmt.Sprintf("USERNAME: %v", fUsername),
			fmt.Sprintf("PASSWORD: %v", fPassword),
			fmt.Sprintf("TOTPSECRET: %v", fTotpSecret),
			fmt.Sprintf("INSECURE: %v", fInsecure),
			fmt.Sprintf("CERTNAME: %v", fCertName),
			fmt.Sprintf("CERTID: %v", fCertID),
			fmt.Sprintf("ISDEFAULT: %v", fIsDefault),
		}, "\n"))

		deployer, err := provider.NewDeployer(&provider.DeployerConfig{
			Hostname:                 fHostname,
			Port:                     int32(fPort),
			Scheme:                   fScheme,
			Username:                 fUsername,
			Password:                 fPassword,
			TotpSecret:               fTotpSecret,
			AllowInsecureConnections: fInsecure,
			CertificateID:            fCertID,
			CertificateName:          fCertName,
			IsDefault:                fIsDefault,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		fInputCertData, _ := os.ReadFile(fInputCertPath)
		fInputKeyData, _ := os.ReadFile(fInputKeyPath)
		res, err := deployer.Deploy(context.Background(), string(fInputCertData), string(fInputKeyData))
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		t.Logf("ok: %v", res)
	})
}
