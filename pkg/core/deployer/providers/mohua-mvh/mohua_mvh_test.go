package mohuamvh_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/mohua-mvh"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fUsername      string
	fApiPassword   string
	fHostID        string
	fDomainID      string
)

func init() {
	argsPrefix := "MOHUAMVH_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fUsername, argsPrefix+"USERNAME", "", "")
	flag.StringVar(&fApiPassword, argsPrefix+"APIPASSWORD", "", "")
	flag.StringVar(&fHostID, argsPrefix+"HOSTID", "", "")
	flag.StringVar(&fDomainID, argsPrefix+"DOMAINID", "", "")
}

/*
Shell command to run this test:

	go test -v ./mohuamvh_test.go -args \
	--MOHUAMVH_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--MOHUAMVH_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--MOHUAMVH_USERNAME="your-username" \
	--MOHUAMVH_APIPASSWORD="your-api-password" \
	--MOHUAMVH_HOSTID="your-virtual-host-id" \
	--MOHUAMVH_DOMAINID="your-domain-id"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("USERNAME: %v", fUsername),
			fmt.Sprintf("APIPASSWORD: %v", fApiPassword),
			fmt.Sprintf("HOSTID: %v", fHostID),
			fmt.Sprintf("DOMAINID: %v", fDomainID),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			Username:    fUsername,
			ApiPassword: fApiPassword,
			HostId:      fHostID,
			DomainId:    fDomainID,
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
