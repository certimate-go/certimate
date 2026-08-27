//go:build tester

package baiducloudcert_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/purger/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/purger/providers/baiducloud-cert"
)

var (
	fp               = tester.InitArgs("BAIDUCLOUDCERT_")
	fAccessKeyId     string
	fSecretAccessKey string
)

func init() {
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./baiducloud_cert_test.go -args \
	--BAIDUCLOUDCERT_ACCESSKEYID="your-access-key-id" \
	--BAIDUCLOUDCERT_SECRETACCESSKEY="your-access-key-secret"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Purge", func(t *testing.T) {
		provider, err := impl.NewPurger(&impl.PurgerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
		})
		require.NoError(t, err)

		tester.Purge(t, provider, tester.PurgeInput{})
	})
}
