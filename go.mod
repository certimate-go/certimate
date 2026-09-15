module github.com/certimate-go/certimate

go 1.25.14

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.23.1
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.14.1
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azcertificates v1.5.0
	github.com/G-Core/gcorelabscdn-go v1.0.37
	github.com/akamai/AkamaiOPEN-edgegrid-golang/v13 v13.4.0
	github.com/alibabacloud-go/alb-20200616/v2 v2.3.2
	github.com/alibabacloud-go/apig-20240327/v11 v11.2.3
	github.com/alibabacloud-go/cas-20200407/v4 v4.7.1
	github.com/alibabacloud-go/cdn-20180510/v10 v10.2.0
	github.com/alibabacloud-go/cloudapi-20160714/v5 v5.8.0
	github.com/alibabacloud-go/darabonba-openapi/v2 v2.2.4
	github.com/alibabacloud-go/dcdn-20180115/v4 v4.1.0
	github.com/alibabacloud-go/ddoscoo-20200101/v5 v5.0.2
	github.com/alibabacloud-go/esa-20240910/v3 v3.13.0
	github.com/alibabacloud-go/fc-20230330/v4 v4.8.2
	github.com/alibabacloud-go/fc-open-20210406/v2 v2.0.12
	github.com/alibabacloud-go/ga-20191120/v4 v4.0.2
	github.com/alibabacloud-go/live-20161101/v3 v3.0.1
	github.com/alibabacloud-go/nlb-20220430/v4 v4.1.3
	github.com/alibabacloud-go/openapi-util v0.1.2
	github.com/alibabacloud-go/slb-20140515/v4 v4.0.14
	github.com/alibabacloud-go/tea v1.5.3
	github.com/alibabacloud-go/tea-utils/v2 v2.0.9
	github.com/alibabacloud-go/vod-20170321/v4 v4.12.1
	github.com/alibabacloud-go/waf-openapi-20211001/v7 v7.8.5
	github.com/aws/aws-sdk-go-v2 v1.47.0
	github.com/aws/aws-sdk-go-v2/config v1.33.5
	github.com/aws/aws-sdk-go-v2/credentials v1.20.5
	github.com/aws/aws-sdk-go-v2/service/acm v1.50.0
	github.com/aws/aws-sdk-go-v2/service/amplify v1.48.0
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.42.0
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.73.0
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.41.0
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.63.0
	github.com/aws/aws-sdk-go-v2/service/iam v1.64.0
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.66.0
	github.com/aws/smithy-go v1.28.1
	github.com/baidubce/bce-sdk-go v0.9.276
	github.com/byteplus-sdk/byteplus-go-sdk-v2 v1.0.78
	github.com/byteplus-sdk/byteplus-sdk-golang v1.0.72
	github.com/fsnotify/fsnotify v1.10.1
	github.com/go-acme/lego/v5 v5.3.1
	github.com/go-cmd/cmd v1.4.3
	github.com/go-resty/resty/v2 v2.17.2
	github.com/go-viper/mapstructure/v2 v2.5.0
	github.com/google/go-querystring v1.2.0
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-plugin v1.8.0
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.1.215
	github.com/jdcloud-api/jdcloud-sdk-go v1.67.0
	github.com/jlaffaye/ftp v0.2.4
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/minio/minio-go/v7 v7.3.0
	github.com/nrdcg/oci-go-sdk/certificatesmanagement/v1065 v1065.124.2
	github.com/nrdcg/oci-go-sdk/common/v1065 v1065.124.2
	github.com/pavlo-v-chernykh/keystore-go/v4 v4.5.0
	github.com/pkg/sftp v1.13.11
	github.com/pocketbase/dbx v1.12.0
	github.com/pocketbase/pocketbase v0.39.10
	github.com/povsister/scp v0.0.0-20250701154629-777cf82de5df
	github.com/pquerna/otp v1.5.0
	github.com/samber/lo v1.53.0
	github.com/spf13/cobra v1.10.2
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/antiddos v1.3.131
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn v1.3.173
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb v1.3.176
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.3.180
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ga2 v1.3.180
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/gaap v1.3.34
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/live v1.3.175
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/scf v1.3.101
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl v1.3.147
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo v1.3.180
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tse v1.3.168
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vod v1.3.179
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/waf v1.3.171
	github.com/ucloud/ucloud-sdk-go v0.22.124
	github.com/volcengine/volc-sdk-golang v1.0.256
	github.com/volcengine/volcengine-go-sdk v1.2.52
	github.com/wneessen/go-mail v0.8.1
	github.com/xhit/go-str2duration/v2 v2.2.0
	github.com/yandex-cloud/go-genproto v0.121.0
	github.com/yandex-cloud/go-sdk/services/certificatemanager v0.0.97
	github.com/yandex-cloud/go-sdk/v2 v2.176.0
	github.com/yunify/qingcloud-sdk-go v2.0.0-alpha.38+incompatible
	gitlab.ecloud.com/ecloud/ecloudsdkcloudcore v1.0.0
	gitlab.ecloud.com/ecloud/ecloudsdkclouddns v1.0.1
	gitlab.ecloud.com/ecloud/ecloudsdkcmcdn v1.0.0
	gitlab.ecloud.com/ecloud/ecloudsdkcore v1.0.6
	gitlab.ecloud.com/ecloud/ecloudsdkvlb v1.0.7
	golang.org/x/crypto v0.55.0
	golang.org/x/oauth2 v0.36.0
	golang.org/x/sync v0.22.0
	golang.org/x/sys v0.47.0
	google.golang.org/api v0.295.0
	google.golang.org/grpc v1.83.2
	google.golang.org/protobuf v1.36.12
	software.sslmate.com/src/go-pkcs12 v0.7.3
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.12.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns v1.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph v0.10.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal v1.2.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.8.0 // indirect
	github.com/alibabacloud-go/alibabacloud-gateway-fc-util v0.0.7 // indirect
	github.com/avast/retry-go v3.0.0+incompatible // indirect
	github.com/aws/aws-sdk-go-v2/service/route53 v1.64.1 // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/go-acme/tencentclouddnspod v1.3.24 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/namedotcom/go/v4 v4.0.2 // indirect
	github.com/nrdcg/bunny-go v0.1.0 // indirect
	github.com/nrdcg/desec v0.11.1 // indirect
	github.com/nrdcg/goacmedns v0.2.0 // indirect
	github.com/nrdcg/porkbun v0.4.0 // indirect
	github.com/ovh/go-ovh v1.9.0 // indirect
	github.com/peterhellberg/link v1.2.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/stretchr/testify v1.12.1
	github.com/vultr/govultr/v3 v3.31.2 // indirect
	go.mongodb.org/mongo-driver v1.17.9 // indirect
	go.uber.org/ratelimit v0.3.1 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	gopkg.in/ns1/ns1-go.v2 v2.18.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

require (
	cloud.google.com/go/auth v0.23.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	github.com/alexbrainman/sspi v0.0.0-20250919150558-7d374ff0d59e // indirect
	github.com/alibabacloud-go/alibabacloud-gateway-spi v0.0.5 // indirect
	github.com/alibabacloud-go/debug v1.0.1 // indirect
	github.com/alibabacloud-go/endpoint-util v1.1.1 // indirect
	github.com/aliyun/credentials-go v1.4.12 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.20.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.10.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.38.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.43.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.51.0 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/bodgit/gssapi v0.0.4 // indirect
	github.com/bodgit/tsig v1.3.1 // indirect
	github.com/boombuler/barcode v1.1.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clbanning/mxj/v2 v2.7.0 // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/domodwyer/mailyak/v3 v3.6.2 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/ganigeorgiev/fexpr v0.6.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-acme/alidns-20150109/v5 v5.5.0 // indirect
	github.com/go-acme/esa-20240910/v3 v3.4.0 // indirect
	github.com/go-acme/jdcloud-sdk-go v1.64.0 // indirect
	github.com/go-acme/tencentedgdeone v1.3.38 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0 // indirect
	github.com/go-test/deep v1.1.1 // indirect
	github.com/gofrs/flock v0.13.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.20 // indirect
	github.com/googleapis/gax-go/v2 v2.24.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jarcoal/httpmock v1.4.1 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.13-0.20220915233716-71ac16282d12 // indirect
	github.com/klauspost/compress v1.19.2 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/klauspost/crc32 v1.3.0 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.23 // indirect
	github.com/maxatome/go-testdeep v1.14.0 // indirect
	github.com/miekg/dns v1.1.72 // indirect
	github.com/minio/crc64nvme v1.1.1 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/nrdcg/namesilo v0.5.0 // indirect
	github.com/nrdcg/oci-go-sdk/dns/v1065 v1065.120.0 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pocketbase/ozzo-validation/v4 v4.3.0 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/sony/gobreaker/v2 v2.4.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/tinylib/msgp v1.6.4 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/yandex-cloud/go-sdk/services/dns v0.0.65 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.67.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/image v0.44.0 // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260715232425-e75dac1f907d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260819154853-08b0e4226688 // indirect
	gopkg.in/ini.v1 v1.67.3 // indirect
	modernc.org/libc v1.74.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.55.0 // indirect
)

replace gitlab.ecloud.com/ecloud/ecloudsdkcloudcore v1.0.0 => ./pkg/sdk3rd-forked/gitlab.ecloud.com/ecloud/ecloudsdkcloudcore@v1.0.0+mod

replace gitlab.ecloud.com/ecloud/ecloudsdkclouddns v1.0.1 => ./pkg/sdk3rd-forked/gitlab.ecloud.com/ecloud/ecloudsdkclouddns@v1.0.1+mod

replace gitlab.ecloud.com/ecloud/ecloudsdkcmcdn v1.0.0 => ./pkg/sdk3rd-forked/gitlab.ecloud.com/ecloud/ecloudsdkcmcdn@v1.0.0

replace gitlab.ecloud.com/ecloud/ecloudsdkcore v1.0.2 => ./pkg/sdk3rd-forked/gitlab.ecloud.com/ecloud/ecloudsdkcore@v1.0.2

replace gitlab.ecloud.com/ecloud/ecloudsdkcore v1.0.6 => ./pkg/sdk3rd-forked/gitlab.ecloud.com/ecloud/ecloudsdkcore@v1.0.6

replace gitlab.ecloud.com/ecloud/ecloudsdkvlb v1.0.7 => ./pkg/sdk3rd-forked/gitlab.ecloud.com/ecloud/ecloudsdkvlb@v1.0.7
