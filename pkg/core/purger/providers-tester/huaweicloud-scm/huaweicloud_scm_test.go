//go:build tester

package huaweicloudscm_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/purger/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/purger/providers/huaweicloud-scm"
)

var (
	fp               = tester.InitArgs("HUAWEICLOUDSCM_")
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
)

func init() {
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./huaweicloud_scm_test.go -args \
	--HUAWEICLOUDSCM_ACCESSKEYID="your-access-key-id" \
	--HUAWEICLOUDSCM_SECRETACCESSKEY="your-access-key-secret" \
	--HUAWEICLOUDSCM_REGION="cn-north-4"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Purge", func(t *testing.T) {
		provider, err := impl.NewPurger(&impl.PurgerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			Region:          fRegion,
		})
		require.NoError(t, err)

		tester.Purge(t, provider, tester.PurgeInput{})
	})
}
