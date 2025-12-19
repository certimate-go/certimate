package ucloudupathx_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-upathx"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fPrivateKey    string
	fPublicKey     string
	fRegion        string
	fAcceleratorId string
	fListenerPort  int
)

func init() {
	argsPrefix := "UCLOUDUPATHX_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fPrivateKey, argsPrefix+"PRIVATEKEY", "", "")
	flag.StringVar(&fPublicKey, argsPrefix+"PUBLICKEY", "", "")
	flag.StringVar(&fRegion, argsPrefix+"REGION", "", "")
	flag.StringVar(&fAcceleratorId, argsPrefix+"ACCELERATORID", "", "")
	flag.IntVar(&fListenerPort, argsPrefix+"LISTENERPORT", 443, "")
}

/*
Shell command to run this test:

	go test -v ./ucloud_upathx_test.go -args \
	--UCLOUDUPATHX_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--UCLOUDUPATHX_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--UCLOUDUPATHX_PRIVATEKEY="your-private-key" \
	--UCLOUDUPATHX_PUBLICKEY="your-public-key" \
	--UCLOUDUPATHX_ACCELERATORID="your-uga-id" \
	--UCLOUDUPATHX_ACCELERATORPORT="443"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("PRIVATEKEY: %v", fPrivateKey),
			fmt.Sprintf("PUBLICKEY: %v", fPublicKey),
			fmt.Sprintf("ACCELERATORID: %v", fAcceleratorId),
			fmt.Sprintf("LISTENERPORT: %v", fListenerPort),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			PrivateKey:    fPrivateKey,
			PublicKey:     fPublicKey,
			AcceleratorId: fAcceleratorId,
			ListenerPort:  int32(fListenerPort),
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
