package restjson_test

import (
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amplify"
	amplifytypes "github.com/aws/aws-sdk-go-v2/service/amplify/types"
	"github.com/stretchr/testify/assert"

	"github.com/certimate-go/certimate/pkg/sdk3rd/aws/zz-shared-smithy/protocol/restjson"
)

func TestProtocol(t *testing.T) {
	t.Run("GetErrorInfo", func(t *testing.T) {
		mockResponse := &http.Response{Header: http.Header{}}
		mockResponse.StatusCode = http.StatusBadRequest
		mockResponse.Header.Set("x-amzn-ErrorType", "InvalidParameterException")
		sdkErr, err := restjson.GetAPIErrorWithRawResponse([]byte(`{"message":"The parameter is invalid."}`), mockResponse)

		assert.NoError(t, err)
		assert.Equal(t, "InvalidParameterException", sdkErr.ErrorCode())
		assert.Equal(t, "The parameter is invalid.", sdkErr.ErrorMessage())
	})

	t.Run("Serialize_[amplify.UpdateDomainAssociationInput]", func(t *testing.T) {
		serializer := restjson.NewSerializer()
		serializer.UseCamelCaseNamePolicy()

		intput := amplify.UpdateDomainAssociationInput{
			AppId:      aws.String("a1b2c3d4e5"),
			DomainName: aws.String("example.com"),
			CertificateSettings: &amplifytypes.CertificateSettings{
				Type:                 amplifytypes.CertificateTypeCustom,
				CustomCertificateArn: aws.String("arn:aws:acm:us-east-1:000000000000:certificate/2717bc82-c2e4-4377-a3df-0de62c5de348"),
			},
		}
		jsonb, err := serializer.Serialize(intput)

		assert.NoError(t, err)
		assert.NotContains(t, string(jsonb), "AppId")
		assert.NotContains(t, string(jsonb), "DomainName")
		assert.NotContains(t, string(jsonb), "CertificateSettings")
		assert.NotContains(t, string(jsonb), "CustomCertificateArn")
		assert.Contains(t, string(jsonb), "appId")
		assert.Contains(t, string(jsonb), "domainName")
		assert.Contains(t, string(jsonb), "certificateSettings")
		assert.Contains(t, string(jsonb), "customCertificateArn")
	})
}
