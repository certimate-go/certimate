package tencentcloudtse_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-tse"
)

var (
	fp            = tester.Args("TENCENTCLOUDTSE_")
	fTestCertPath string
	fTestKeyPath  string
	fSecretId     string
	fSecretKey    string
	fRegion       string
	fServiceType  string
	fGatewayId    string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fServiceType, "SERVICETYPE")
	fp.DefineString(&fGatewayId, "GATEWAYID")
}

/*
Shell command to run this test:

	go test -v ./tencentcloud_tse_test.go -args \
	--TENCENTCLOUDTSE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDTSE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDTSE_SECRETID="your-secret-id" \
	--TENCENTCLOUDTSE_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDTSE_REGION="ap-guangzhou" \
	--TENCENTCLOUDTSE_SERVICETYPE="cloudnative" \
	--TENCENTCLOUDTSE_GATEWAYID="your-gateway-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			SecretId:    fSecretId,
			SecretKey:   fSecretKey,
			Region:      fRegion,
			ServiceType: fServiceType,
			GatewayId:   fGatewayId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
