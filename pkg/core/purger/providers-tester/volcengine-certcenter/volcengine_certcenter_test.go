//go:build tester

package volcenginecertcenter_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/purger/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/purger/providers/volcengine-certcenter"
)

var (
	fp               = tester.InitArgs("VOLCENGINECERTCENTER_")
	fAccessKeyId     string
	fSecretAccessKey string
)

func init() {
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./volcengine_certcenter_test.go -args \
	--VOLCENGINECERTCENTER_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINECERTCENTER_SECRETACCESSKEY="your-secret-access-key"
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
