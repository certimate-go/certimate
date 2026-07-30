package amzjson_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/stretchr/testify/assert"

	"github.com/certimate-go/certimate/pkg/sdk3rd/aws/zz-shared-smithy/protocol/amzjson"
)

func TestProtocol(t *testing.T) {
	t.Run("GetAPIError", func(t *testing.T) {
		sdkErr, err := amzjson.GetAPIInfo([]byte(`{"__type":"InvalidParameterException","message":"The parameter is invalid."}`))

		assert.NoError(t, err)
		assert.Equal(t, "InvalidParameterException", sdkErr.ErrorCode())
		assert.Equal(t, "The parameter is invalid.", sdkErr.ErrorMessage())
	})

	t.Run("Deserialize_[acm.ListCertificatesOutput]", func(t *testing.T) {
		deserializer := amzjson.NewDeserializer()
		deserializer.UseEpochTime()

		var output *acm.ListCertificatesOutput
		err := deserializer.Deserialize([]byte(`{"CertificateSummaryList":[{"CertificateArn":"arn:aws:acm:us-east-1:000000000000:certificate/2717bc82-c2e4-4377-a3df-0de62c5de348","DomainName":"*.example.com","SubjectAlternativeNameSummaries":["*.example.com","example.com"],"HasAdditionalSubjectAlternativeNames":false,"Status":"ISSUED","Type":"IMPORTED","KeyAlgorithm":"EC_prime256v1","KeyUsages":["DIGITAL_SIGNATURE","KEY_ENCIPHERMENT"],"ExtendedKeyUsages":["TLS_WEB_SERVER_AUTHENTICATION","TLS_WEB_CLIENT_AUTHENTICATION"],"InUse":false,"Exported":false,"RenewalEligibility":"INELIGIBLE","NotBefore":1782835200,"NotAfter":1785513599,"ImportedAt":1784055845},{"CertificateArn":"arn:aws:acm:us-east-1:000000000000:certificate/2717bc82-c2e4-4377-a3df-0de62c5de349","DomainName":"*.example.com","SubjectAlternativeNameSummaries":["*.isafe-tech.com","example.com"],"HasAdditionalSubjectAlternativeNames":false,"Status":"ISSUED","Type":"IMPORTED","KeyAlgorithm":"EC-prime256v1","KeyUsages":["DIGITAL_SIGNATURE","KEY_ENCIPHERMENT"],"ExtendedKeyUsages":["TLS_WEB_SERVER_AUTHENTICATION","TLS_WEB_CLIENT_AUTHENTICATION"],"InUse":false,"Exported":false,"RenewalEligibility":"INELIGIBLE","NotBefore":1782835200.0,"NotAfter":1785513599.0,"ImportedAt":1.784055846E9}]}`), &output)

		assert.NoError(t, err)
		assert.Len(t, output.CertificateSummaryList, 2)
		assert.Equal(t, "arn:aws:acm:us-east-1:000000000000:certificate/2717bc82-c2e4-4377-a3df-0de62c5de348", *output.CertificateSummaryList[0].CertificateArn)
		assert.Equal(t, "*.example.com", *output.CertificateSummaryList[0].DomainName)
		assert.Equal(t, []string{"*.example.com", "example.com"}, output.CertificateSummaryList[0].SubjectAlternativeNameSummaries)
		assert.Equal(t, false, *output.CertificateSummaryList[0].HasAdditionalSubjectAlternativeNames)
		assert.Equal(t, acmtypes.CertificateStatusIssued, output.CertificateSummaryList[0].Status)
		assert.Equal(t, acmtypes.CertificateTypeImported, output.CertificateSummaryList[0].Type)
		assert.Equal(t, acmtypes.KeyAlgorithmEcPrime256v1, output.CertificateSummaryList[0].KeyAlgorithm)
		assert.Equal(t, []acmtypes.KeyUsageName{acmtypes.KeyUsageNameDigitalSignature, acmtypes.KeyUsageNameKeyEncipherment}, output.CertificateSummaryList[0].KeyUsages)
		assert.Equal(t, []acmtypes.ExtendedKeyUsageName{acmtypes.ExtendedKeyUsageNameTlsWebServerAuthentication, acmtypes.ExtendedKeyUsageNameTlsWebClientAuthentication}, output.CertificateSummaryList[0].ExtendedKeyUsages)
		assert.Equal(t, false, *output.CertificateSummaryList[0].InUse)
		assert.Equal(t, false, *output.CertificateSummaryList[0].Exported)
		assert.Equal(t, acmtypes.RenewalEligibilityIneligible, output.CertificateSummaryList[0].RenewalEligibility)
		assert.Equal(t, int64(1782835200), output.CertificateSummaryList[0].NotBefore.Unix())
		assert.Equal(t, int64(1785513599), output.CertificateSummaryList[0].NotAfter.Unix())
		assert.Equal(t, int64(1784055845), output.CertificateSummaryList[0].ImportedAt.Unix())
		assert.Equal(t, "arn:aws:acm:us-east-1:000000000000:certificate/2717bc82-c2e4-4377-a3df-0de62c5de349", *output.CertificateSummaryList[1].CertificateArn)
		assert.Equal(t, int64(1782835200), output.CertificateSummaryList[1].NotBefore.Unix())
		assert.Equal(t, int64(1785513599), output.CertificateSummaryList[1].NotAfter.Unix())
		assert.Equal(t, int64(1784055846), output.CertificateSummaryList[1].ImportedAt.Unix())
	})
}
