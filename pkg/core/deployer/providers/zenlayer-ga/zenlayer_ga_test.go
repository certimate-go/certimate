package zenlayerga_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/zenlayer-ga"
)

var (
	fInputCertPath     string
	fInputKeyPath      string
	fAccessKeyId       string
	fAccessKeyPassword string
	fAcceleratorId     string
	fCertificateId     string
)

func init() {
	argsPrefix := "ZENLAYERGA_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fAccessKeyId, argsPrefix+"ACCESSKEYID", "", "")
	flag.StringVar(&fAccessKeyPassword, argsPrefix+"ACCESSKEYPASSWORD", "", "")
	flag.StringVar(&fAcceleratorId, argsPrefix+"ACCELERATORID", "", "")
	flag.StringVar(&fCertificateId, argsPrefix+"CERTIFICATEID", "", "")
}

/*
Shell command to run this test:

	go test -v ./zenlayer_ga_test.go -args \
	--ZENLAYERGA_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--ZENLAYERGA_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--ZENLAYERGA_ACCESSKEYID="your-access-key-id" \
	--ZENLAYERGA_ACCESSKEYPASSWORD="your-access-key-password" \
	--ZENLAYERGA_ACCELERATORID="your-ga-accelerator-id" \
	--ZENLAYERGA_CERTIFICATEID="your-ga-certificate-id"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy_ToAccelerator", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("ACCESSKEYID: %v", fAccessKeyId),
			fmt.Sprintf("ACCESSKEYPASSWORD: %v", fAccessKeyPassword),
			fmt.Sprintf("ACCELERATORID: %v", fAcceleratorId),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			AccessKeyId:       fAccessKeyId,
			AccessKeyPassword: fAccessKeyPassword,
			ResourceType:      provider.RESOURCE_TYPE_ACCELERATOR,
			AcceleratorId:     fAcceleratorId,
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

	t.Run("Deploy_ToCertificate", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("ACCESSKEYID: %v", fAccessKeyId),
			fmt.Sprintf("ACCESSKEYPASSWORD: %v", fAccessKeyPassword),
			fmt.Sprintf("CERTIFICATEID: %v", fCertificateId),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			AccessKeyId:       fAccessKeyId,
			AccessKeyPassword: fAccessKeyPassword,
			ResourceType:      provider.RESOURCE_TYPE_CERTIFICATE,
			CertificateId:     fCertificateId,
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
