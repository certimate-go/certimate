package vercel_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/vercel"
)

var (
	fInputCertPath  string
	fInputKeyPath   string
	fApiAccessToken string
	fTeamId         string
)

func init() {
	argsPrefix := "VERCEL_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fApiAccessToken, argsPrefix+"APIACCESSTOKEN", "", "")
	flag.StringVar(&fTeamId, argsPrefix+"TEAMID", "", "")
}

/*
Shell command to run this test:

	go test -v ./vercel_test.go -args \
	--VERCEL_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--VERCEL_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--VERCEL_APIACCESSTOKEN="your-api-access-token" \
	--VERCEL_TEAMID="your-team-id"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("APIACCESSTOKEN: %v", fApiAccessToken),
			fmt.Sprintf("TEAMID: %v", fTeamId),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			ApiAccessToken: fApiAccessToken,
			ResourceType:   provider.RESOURCE_TYPE_WEBSITE,
			TeamId:         fTeamId,
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
