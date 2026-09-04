//go:build tester

package aliyunesa_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/purger/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/purger/providers/aliyun-esa"
)

var (
	fp               = tester.InitArgs("ALIYUNESA_")
	fAccessKeyId     string
	fAccessKeySecret string
	fRegion          string
	fSiteId          int64
)

func init() {
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineInt64(&fSiteId, "SITEID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./aliyun_esa_test.go -args \
	--ALIYUNESA_ACCESSKEYID="your-access-key-id" \
	--ALIYUNESA_ACCESSKEYSECRET="your-access-key-secret" \
	--ALIYUNESA_REGION="cn-hangzhou" \
	--ALIYUNESA_SITEID="your-esa-site-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Purge", func(t *testing.T) {
		provider, err := impl.NewPurger(&impl.PurgerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			Region:          fRegion,
			SiteId:          fSiteId,
		})
		require.NoError(t, err)

		tester.Purge(t, provider, tester.PurgeInput{})
	})
}
