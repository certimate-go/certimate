//go:build tester

package awsacm_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/purger/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/purger/providers/aws-acm"
)

var (
	fp               = tester.InitArgs("AWSACM_")
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

	go test -tags=tester -v ./aws_acm_test.go -args \
	--AWSACM_ACCESSKEYID="your-access-key-id" \
	--AWSACM_SECRETACCESSKEY="your-access-key-secret" \
	--AWSACM_REGION="us-east-1"
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
