//go:build tester

package mattermost_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/notifier/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/mattermost"
)

var (
	fp         = tester.InitArgs("MATTERMOST_")
	fServerUrl string
	fChannelId string
	fUsername  string
	fPassword  string
)

func init() {
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fChannelId, "CHANNELID")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./mattermost_test.go -args \
	--MATTERMOST_SERVERURL="https://example.com/your-server-url" \
	--MATTERMOST_CHANNELID="your-chanel-id" \
	--MATTERMOST_USERNAME="your-username" \
	--MATTERMOST_PASSWORD="your-password"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Notify", func(t *testing.T) {
		provider, err := impl.NewNotifier(&impl.NotifierConfig{
			ServerUrl: fServerUrl,
			ChannelId: fChannelId,
			Username:  fUsername,
			Password:  fPassword,
		})
		require.NoError(t, err)

		tester.Notify(t, provider, tester.NotifyInput{})
	})
}
