package netlify_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/netlify"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fApiToken      string
	fSiteId        string
)

func init() {
	argsPrefix := "NETLIFY_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fApiToken, argsPrefix+"APITOKEN", "", "")
	flag.StringVar(&fSiteId, argsPrefix+"SITEID", "", "")
}

/*
Shell command to run this test:

	go test -v ./netlify_test.go -args \
	--NETLIFY_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--NETLIFY_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--NETLIFY_APITOKEN="your-api-token" \
	--NETLIFY_SITEID="your-site-id"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("APITOKEN: %v", fApiToken),
			fmt.Sprintf("SITEID: %v", fSiteId),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			ApiToken:     fApiToken,
			ResourceType: provider.RESOURCE_TYPE_WEBSITE,
			SiteId:       fSiteId,
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
