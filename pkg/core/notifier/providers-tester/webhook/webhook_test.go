//go:build tester

package webhook_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/notifier/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/webhook"
)

var (
	fp                  = tester.InitArgs("WEBHOOK_")
	fWebhookUrl         string
	fWebhookContentType string
)

func init() {
	fp.DefineString(&fWebhookUrl, "URL")
	fp.DefineString(&fWebhookContentType, "CONTENTTYPE", "application/json")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./webhook_test.go -args \
	--WEBHOOK_URL="https://example.com/your-webhook-url" \
	--WEBHOOK_CONTENTTYPE="application/json"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Notify", func(t *testing.T) {
		provider, err := impl.NewNotifier(&impl.NotifierConfig{
			WebhookUrl: fWebhookUrl,
			Method:     "POST",
			Headers: map[string]string{
				"Content-Type": fWebhookContentType,
			},
			AllowInsecureConnections: true,
		})
		require.NoError(t, err)

		tester.Notify(t, provider, tester.NotifyInput{})
	})
}
