package webhook_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/notifier/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/webhook"
)

var (
	fp                  = tester.Args("WEBHOOK_")
	fWebhookUrl         string
	fWebhookContentType string
)

func init() {
	fp.DefineString(&fWebhookUrl, "URL")
	fp.DefineString(&fWebhookContentType, "CONTENTTYPE", "application/json")
}

/*
Shell command to run this test:

	go test -v ./webhook_test.go -args \
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
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestNotify(t, provider, tester.TestNotifyArgs{})
	})
}
