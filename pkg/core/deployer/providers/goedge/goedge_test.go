package goedge_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/goedge"
)

var (
	fp             = tester.Args("GOEDGE_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fAccessKeyId   string
	fAccessKey     string
	fCertificateId int64
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKey, "ACCESSKEY")
	fp.DefineInt64(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./goedge_test.go -args \
	--GOEDGE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--GOEDGE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--GOEDGE_SERVERURL="http://127.0.0.1:7788" \
	--GOEDGE_ACCESSKEYID="your-access-key-id" \
	--GOEDGE_ACCESSKEY="your-access-key" \
	--GOEDGE_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToCertificate", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiRole:                  "user",
			AccessKeyId:              fAccessKeyId,
			AccessKey:                fAccessKey,
			AllowInsecureConnections: true,
			ResourceType:             impl.RESOURCE_TYPE_CERTIFICATE,
			CertificateId:            fCertificateId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
