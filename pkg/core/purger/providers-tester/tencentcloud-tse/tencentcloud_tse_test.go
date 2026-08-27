//go:build tester

package tencentcloudtse_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/purger/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/purger/providers/tencentcloud-tse"
)

var (
	fp           = tester.InitArgs("TENCENTCLOUDTSE_")
	fSecretId    string
	fSecretKey   string
	fRegion      string
	fServiceType string
	fGatewayId   string
)

func init() {
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fServiceType, "SERVICETYPE")
	fp.DefineString(&fGatewayId, "GATEWAYID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./tencentcloud_tse_test.go -args \
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

	t.Run("Purge", func(t *testing.T) {
		provider, err := impl.NewPurger(&impl.PurgerConfig{
			SecretId:    fSecretId,
			SecretKey:   fSecretKey,
			Region:      fRegion,
			ServiceType: fServiceType,
			GatewayId:   fGatewayId,
		})
		require.NoError(t, err)

		tester.Purge(t, provider, tester.PurgeInput{})
	})
}
