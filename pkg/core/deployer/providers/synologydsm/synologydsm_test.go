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
	fInputCertPath       string
	fInputKeyPath        string
	fServerUrl           string
	fUsername            string
	fPassword            string
	fTotpSecret          string
	fCertificateIdOrDesc string
	fIsDefault           bool
)

func init() {
	argsPrefix := "SYNOLOGYDSM_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fServerUrl, argsPrefix+"SERVERURL", "", "")
	flag.StringVar(&fUsername, argsPrefix+"USERNAME", "", "")
	flag.StringVar(&fPassword, argsPrefix+"PASSWORD", "", "")
	flag.StringVar(&fTotpSecret, argsPrefix+"TOTPSECRET", "", "")
	flag.StringVar(&fCertificateIdOrDesc, argsPrefix+"CERTIFICATEIDORDESC", "", "")
	flag.BoolVar(&fIsDefault, argsPrefix+"ISDEFAULT", false, "")
}

/*
Shell command to run this test:

	go test -v ./synology_dsm_test.go -args \
	--SYNOLOGYDSM_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--SYNOLOGYDSM_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--SYNOLOGYDSM_SERVERURL="http://127.0.0.1:5000/" \
	--SYNOLOGYDSM_USERNAME="admin" \
	--SYNOLOGYDSM_PASSWORD="password" \
	--SYNOLOGYDSM_CERTIFICATEIDORDESC="your-certificate-id-or-desc" \
	--SYNOLOGYDSM_ISDEFAULT=true
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("SERVERURL: %v", fServerUrl),
			fmt.Sprintf("USERNAME: %v", fUsername),
			fmt.Sprintf("PASSWORD: %v", fPassword),
			fmt.Sprintf("TOTPSECRET: %v", fTotpSecret),
			fmt.Sprintf("CERTIFICATEIDORDESC: %v", fCertificateIdOrDesc),
			fmt.Sprintf("ISDEFAULT: %v", fIsDefault),
		}, "\n"))

		deployer, err := provider.NewDeployer(&provider.DeployerConfig{
			ServerUrl:                  fServerUrl,
			Username:                   fUsername,
			Password:                   fPassword,
			TotpSecret:                 fTotpSecret,
			AllowInsecureConnections:   true,
			CertificateIdOrDescription: fCertificateIdOrDesc,
			IsDefault:                  fIsDefault,
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
