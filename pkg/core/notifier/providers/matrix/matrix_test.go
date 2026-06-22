package matrix_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/notifier/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/matrix"
)

var (
	fp           = tester.Args("MATRIX_")
	fServerUrl   string
	fUserId      string
	fAccessToken string
	fRoomId      string
)

func init() {
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fUserId, "USERID")
	fp.DefineString(&fAccessToken, "ACCESSTOKEN")
	fp.DefineString(&fRoomId, "ROOMID")
}

/*
Shell command to run this test:

	go test -v ./matrix_test.go -args \
	--MATRIX_SERVERURL="https://example.com/your-matrix-server" \
	--MATRIX_USERID="@bot:example.org" \
	--MATRIX_ACCESSTOKEN="your-access-token" \
	--MATRIX_ROOMID="!room:example.org"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Notify", func(t *testing.T) {
		provider, err := impl.NewNotifier(&impl.NotifierConfig{
			ServerUrl:   fServerUrl,
			UserId:      fUserId,
			AccessToken: fAccessToken,
			RoomId:      fRoomId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestNotify(t, provider, tester.TestNotifyArgs{})
	})
}
