package restxml_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/certimate-go/certimate/pkg/sdk3rd/aws/zz-shared-smithy/protocol/restxml"
)

func TestProtocol(t *testing.T) {
	t.Run("GetAPIError", func(t *testing.T) {
		sdkErr, err := restxml.GetAPIError([]byte(`<ErrorResponse xmlns="http://webservices.amazon.com/AWSFault/2005-15-09"><Error><Type>Sender</Type><Code>InvalidParameterException</Code><Message>The parameter is invalid.</Message></Error><RequestId>645a617f-73b5-4882-bb23-d68d22d16a76</RequestId></ErrorResponse>`))

		assert.NoError(t, err)
		assert.Equal(t, "InvalidParameterException", sdkErr.ErrorCode())
		assert.Equal(t, "The parameter is invalid.", sdkErr.ErrorMessage())
	})
}
