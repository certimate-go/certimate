package onepanel_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/1panel"
)

var (
	fp             = tester.Args("1PANEL_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fApiVersion    string
	fApiKey        string
	fWebsiteId     int64
	fCertificateId int64
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiVersion, "APIVERSION", "v2")
	fp.DefineString(&fApiKey, "APIKEY")
	fp.DefineInt64(&fWebsiteId, "WEBSITEID")
	fp.DefineInt64(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./1panel_test.go -args \
	--1PANEL_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--1PANEL_TESTKEYPATH="/path/to/your-test-key.pem" \
	--1PANEL_SERVERURL="http://127.0.0.1:20410" \
	--1PANEL_APIVERSION="v2" \
	--1PANEL_APIKEY="your-api-key" \
	--1PANEL_WEBSITEID="your-website-id" \
	--1PANEL_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToWebsite", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiVersion:               fApiVersion,
			ApiKey:                   fApiKey,
			AllowInsecureConnections: true,
			DeployTarget:             impl.DEPLOY_TARGET_WEBSITE,
			WebsiteMatchPattern:      impl.WEBSITE_MATCH_PATTERN_SPECIFIED,
			WebsiteId:                fWebsiteId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_ToCertificate", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiVersion:               fApiVersion,
			ApiKey:                   fApiKey,
			AllowInsecureConnections: true,
			DeployTarget:             impl.DEPLOY_TARGET_CERTIFICATE,
			CertificateId:            fCertificateId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
