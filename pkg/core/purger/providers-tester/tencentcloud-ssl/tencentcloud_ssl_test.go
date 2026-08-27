//go:build tester

package tencentcloudssl_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/purger/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/purger/providers/tencentcloud-ssl"
)

var (
	fp         = tester.InitArgs("TENCENTCLOUDSSL_")
	fSecretId  string
	fSecretKey string
)

func init() {
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./tencentcloud_ssl_test.go -args \
	--TENCENTCLOUDSSL_SECRETID="your-secret-id" \
	--TENCENTCLOUDSSL_SECRETKEY="your-secret-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Purge", func(t *testing.T) {
		provider, err := impl.NewPurger(&impl.PurgerConfig{
			SecretId:  fSecretId,
			SecretKey: fSecretKey,
		})
		require.NoError(t, err)

		tester.Purge(t, provider, tester.PurgeInput{})
	})
}
