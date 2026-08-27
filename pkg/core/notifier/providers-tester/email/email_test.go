//go:build tester

package email_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/notifier/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/email"
)

var (
	fp               = tester.InitArgs("EMAIL_")
	fSmtpHost        string
	fSmtpPort        int64
	fSmtpTLS         bool
	fUsername        string
	fPassword        string
	fSenderAddress   string
	fReceiverAddress string
)

func init() {
	fp.DefineString(&fSmtpHost, "SMTPHOST")
	fp.DefineInt64(&fSmtpPort, "SMTPPORT", 25)
	fp.DefineBool(&fSmtpTLS, "SMTPTLS", false)
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
	fp.DefineString(&fSenderAddress, "SENDERADDRESS")
	fp.DefineString(&fReceiverAddress, "RECEIVERADDRESS")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./email_test.go -args \
	--EMAIL_SMTPHOST="smtp.example.com" \
	--EMAIL_SMTPPORT=465 \
	--EMAIL_SMTPTLS=true \
	--EMAIL_USERNAME="your-username" \
	--EMAIL_PASSWORD="your-password" \
	--EMAIL_SENDERADDRESS="sender@example.com" \
	--EMAIL_RECEIVERADDRESS="receiver@example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Notify_Plain", func(t *testing.T) {
		provider, err := impl.NewNotifier(&impl.NotifierConfig{
			SmtpHost:        fSmtpHost,
			SmtpPort:        int32(fSmtpPort),
			SmtpTls:         fSmtpTLS,
			Username:        fUsername,
			Password:        fPassword,
			SenderAddress:   fSenderAddress,
			ReceiverAddress: fReceiverAddress,
		})
		require.NoError(t, err)

		tester.Notify(t, provider, tester.NotifyInput{})
	})

	t.Run("Notify_Html", func(t *testing.T) {
		provider, err := impl.NewNotifier(&impl.NotifierConfig{
			SmtpHost:        fSmtpHost,
			SmtpPort:        int32(fSmtpPort),
			SmtpTls:         fSmtpTLS,
			Username:        fUsername,
			Password:        fPassword,
			SenderAddress:   fSenderAddress,
			ReceiverAddress: fReceiverAddress,
			MessageFormat:   impl.MESSAGE_FORMAT_HTML,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		const mockHtml = "<h1>Hello Certimate！</h1><a onblur=\"alert(secret)\" href=\"http://www.google.com\">Google</a>"
		tester.Notify(t, provider, tester.NotifyInput{Message: mockHtml})
	})
}
