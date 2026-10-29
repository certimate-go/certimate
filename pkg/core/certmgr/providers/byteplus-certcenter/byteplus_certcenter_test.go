package bytepluscertcenter

import (
	"testing"

	"github.com/stretchr/testify/require"

	bpcertificateservice "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/byteplus-sdk/byteplus-go-sdk-v2/service/certificateservice"
)

func TestCreateSDKClient(t *testing.T) {
	for _, region := range []string{"", "ap-southeast-1"} {
		client, err := createSDKClient("test-ak", "test-sk", region)
		require.NoError(t, err)
		require.Equal(t, "https://certificate-service.byteplusapi.com", client.Endpoint)

		req, _ := client.UploadCertificateRequest(&bpcertificateservice.UploadCertificateInput{})
		require.NoError(t, req.Build())
		require.Equal(t, "certificate-service.byteplusapi.com", req.HTTPRequest.URL.Host)
	}
}
