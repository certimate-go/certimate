package query_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elasticloadbalancingv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/stretchr/testify/assert"

	"github.com/certimate-go/certimate/pkg/sdk3rd/aws/zz-shared-smithy/protocol/query"
)

func TestProtocol(t *testing.T) {
	t.Run("GetAPIError", func(t *testing.T) {
		sdkErr, err := query.GetAPIError([]byte(`<ErrorResponse xmlns="http://webservices.amazon.com/AWSFault/2005-15-09"><Error><Type>Sender</Type><Code>InvalidParameterException</Code><Message>The parameter is invalid.</Message></Error><RequestId>645a617f-73b5-4882-bb23-d68d22d16a76</RequestId></ErrorResponse>`))

		assert.NoError(t, err)
		assert.Equal(t, "InvalidParameterException", sdkErr.ErrorCode())
		assert.Equal(t, "The parameter is invalid.", sdkErr.ErrorMessage())
	})

	t.Run("Deserialize_[iam.ListServerCertificatesOutput]", func(t *testing.T) {
		deserializer := query.NewDeserializer("ListServerCertificates")

		var output *iam.ListServerCertificatesOutput
		err := deserializer.Deserialize([]byte(`<ListServerCertificatesResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/"><ListServerCertificatesResult><IsTruncated>false</IsTruncated><ServerCertificateMetadataList><member><ServerCertificateName>ProdServerCert</ServerCertificateName><Path>/company/servercerts/</Path><Arn>arn:aws:iam::123456789012:server-certificate/company/servercerts/ProdServerCert</Arn><UploadDate>2010-05-08T01:02:03.004Z</UploadDate><ServerCertificateId>ASCACKCEVSQ6CEXAMPLE1</ServerCertificateId><Expiration>2012-05-08T01:02:03.004Z</Expiration></member><member><ServerCertificateName>BetaServerCert</ServerCertificateName><Path>/company/servercerts/</Path><Arn>arn:aws:iam::123456789012:server-certificate/company/servercerts/BetaServerCert</Arn><UploadDate>2010-05-08T02:03:01.004Z</UploadDate><ServerCertificateId>ASCACKCEVSQ6CEXAMPLE2</ServerCertificateId><Expiration>2012-05-08T02:03:01.004Z</Expiration></member><member><ServerCertificateName>TestServerCert</ServerCertificateName><Path>/company/servercerts/</Path><Arn>arn:aws:iam::123456789012:server-certificate/company/servercerts/TestServerCert</Arn><UploadDate>2010-05-08T03:01:02.004Z</UploadDate><ServerCertificateId>ASCACKCEVSQ6CEXAMPLE3</ServerCertificateId><Expiration>2012-05-08T03:01:02.004Z</Expiration></member></ServerCertificateMetadataList></ListServerCertificatesResult><ResponseMetadata><RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId></ResponseMetadata></ListServerCertificatesResponse>`), &output)

		assert.NoError(t, err)
		assert.Len(t, output.ServerCertificateMetadataList, 3)
		assert.Equal(t, "arn:aws:iam::123456789012:server-certificate/company/servercerts/ProdServerCert", *output.ServerCertificateMetadataList[0].Arn)
		assert.Equal(t, "ASCACKCEVSQ6CEXAMPLE1", *output.ServerCertificateMetadataList[0].ServerCertificateId)
		assert.Equal(t, "ProdServerCert", *output.ServerCertificateMetadataList[0].ServerCertificateName)
		assert.Equal(t, "/company/servercerts/", *output.ServerCertificateMetadataList[0].Path)
		assert.Equal(t, int64(1273280523), output.ServerCertificateMetadataList[0].UploadDate.Unix())
		assert.Equal(t, int64(1336438923004), output.ServerCertificateMetadataList[0].Expiration.UnixMilli())
	})

	t.Run("Deserialize_[iam.ListServerCertificatesOutput]_panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			deserializer := query.NewDeserializer("")

			var output *iam.ListServerCertificatesOutput
			err := deserializer.Deserialize([]byte(`<ListServerCertificatesResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/"><ListServerCertificatesResult><IsTruncated>false</IsTruncated><ServerCertificateMetadataList><member><ServerCertificateName>ProdServerCert</ServerCertificateName><Path>/company/servercerts/</Path><Arn>arn:aws:iam::123456789012:server-certificate/company/servercerts/ProdServerCert</Arn><UploadDate>2010-05-08T01:02:03.004Z</UploadDate><ServerCertificateId>ASCACKCEVSQ6CEXAMPLE1</ServerCertificateId><Expiration>2012-05-08T01:02:03.004Z</Expiration></member><member><ServerCertificateName>BetaServerCert</ServerCertificateName><Path>/company/servercerts/</Path><Arn>arn:aws:iam::123456789012:server-certificate/company/servercerts/BetaServerCert</Arn><UploadDate>2010-05-08T02:03:01.004Z</UploadDate><ServerCertificateId>ASCACKCEVSQ6CEXAMPLE2</ServerCertificateId><Expiration>2012-05-08T02:03:01.004Z</Expiration></member><member><ServerCertificateName>TestServerCert</ServerCertificateName><Path>/company/servercerts/</Path><Arn>arn:aws:iam::123456789012:server-certificate/company/servercerts/TestServerCert</Arn><UploadDate>2010-05-08T03:01:02.004Z</UploadDate><ServerCertificateId>ASCACKCEVSQ6CEXAMPLE3</ServerCertificateId><Expiration>2012-05-08T03:01:02.004Z</Expiration></member></ServerCertificateMetadataList></ListServerCertificatesResult><ResponseMetadata><RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId></ResponseMetadata></ListServerCertificatesResponse>`), &output)

			assert.NoError(t, err)
			assert.Len(t, output.ServerCertificateMetadataList, 0)
		})
	})

	t.Run("SerializeToMap_[elasticloadbalancingv2.DescribeLoadBalancers]", func(t *testing.T) {
		serializer := query.NewSerializer()
		serializer.UseOmitEmptyValue()

		input := &elasticloadbalancingv2.DescribeLoadBalancersInput{
			LoadBalancerArns: []string{"arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188"},
		}
		paramsMap, err := serializer.SerializeToMap(input)

		assert.NoError(t, err)
		assert.Len(t, paramsMap, 1)
		assert.Equal(t, "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188", paramsMap["LoadBalancerArns.member.1"])
	})

	t.Run("SerializeToMap_[elasticloadbalancingv2.ModifyListener]", func(t *testing.T) {
		serializer := query.NewSerializer()
		serializer.UseOmitEmptyValue()

		input := &elasticloadbalancingv2.ModifyListenerInput{
			ListenerArn: aws.String("arn:aws:elasticloadbalancing:us-west-2:123456789012:listener/app/my-load-balancer/50dc6c495c0c9188/f2f7dc8efc522ab2"),
			DefaultActions: []elasticloadbalancingv2types.Action{
				{
					Type:           elasticloadbalancingv2types.ActionTypeEnumForward,
					TargetGroupArn: aws.String("arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-new-targets/2453ed029918f21e"),
				},
				{
					Type:           elasticloadbalancingv2types.ActionTypeEnumRedirect,
					TargetGroupArn: aws.String("arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-new-targets/2453ed029918f21f"),
				},
			},
		}
		paramsMap, err := serializer.SerializeToMap(input)

		assert.NoError(t, err)
		assert.Len(t, paramsMap, 5)
		assert.Equal(t, "arn:aws:elasticloadbalancing:us-west-2:123456789012:listener/app/my-load-balancer/50dc6c495c0c9188/f2f7dc8efc522ab2", paramsMap["ListenerArn"])
		assert.Equal(t, "forward", paramsMap["DefaultActions.member.1.Type"])
		assert.Equal(t, "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-new-targets/2453ed029918f21e", paramsMap["DefaultActions.member.1.TargetGroupArn"])
		assert.Equal(t, "redirect", paramsMap["DefaultActions.member.2.Type"])
		assert.Equal(t, "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-new-targets/2453ed029918f21f", paramsMap["DefaultActions.member.2.TargetGroupArn"])
	})
}
