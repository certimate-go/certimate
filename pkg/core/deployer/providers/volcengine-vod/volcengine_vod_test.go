package volcenginevod_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-vod"
)

var (
	fInputCertPath   string
	fInputKeyPath    string
	fAccessKeyId     string
	fAccessKeySecret string
	fDomain          string
	fSpaceName       string
	fDomainType      string
)

func init() {
	argsPrefix := "VOLCENGINEVOD_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fAccessKeyId, argsPrefix+"ACCESSKEYID", "", "")
	flag.StringVar(&fAccessKeySecret, argsPrefix+"ACCESSKEYSECRET", "", "")
	flag.StringVar(&fDomain, argsPrefix+"DOMAIN", "", "")
	flag.StringVar(&fSpaceName, argsPrefix+"SPACENAME", "", "")
	flag.StringVar(&fDomainType, argsPrefix+"DOMAINTYPE", "", "")
}

/*
Shell command to run this test:

	go test -v ./volcengine_vod_test.go -args \
	--VOLCENGINEVOD_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--VOLCENGINEVOD_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--VOLCENGINEVOD_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINEVOD_ACCESSKEYSECRET="your-access-key-secret" \
	--VOLCENGINEVOD_DOMAIN="example.com" \
	--VOLCENGINEVOD_DOMAINTYPE="play" \
	--VOLCENGINEVOD_SPACENAME="vod-space-name"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("ACCESSKEYID: %v", fAccessKeyId),
			fmt.Sprintf("ACCESSKEYSECRET: %v", fAccessKeySecret),
			fmt.Sprintf("DOMAIN: %v", fDomain),
			fmt.Sprintf("DOMAINTYPE: %v", fDomainType),
			fmt.Sprintf("SPACENAME: %v", fSpaceName),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			AccessKeySecret:    fAccessKeySecret,
			DomainMatchPattern: provider.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
			SpaceName:          fSpaceName,
			DomainType:         fDomainType,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		fInputCertData, _ := os.ReadFile(fInputCertPath)
		fInputKeyData, _ := os.ReadFile(fInputKeyPath)
		res, err := provider.Deploy(context.Background(), string(fInputCertData), string(fInputKeyData))
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		t.Logf("ok: %v", res)
	})
}
