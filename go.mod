module github.com/certimate-go/certimate

go 1.25.0

toolchain go1.25.5

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.20.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.13.1
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azcertificates v1.4.0
	github.com/G-Core/gcorelabscdn-go v1.0.35
	github.com/KscSDK/ksc-sdk-go v0.17.0
	github.com/akamai/AkamaiOPEN-edgegrid-golang/v11 v11.1.0
	github.com/alibabacloud-go/alb-20200616/v2 v2.2.10
	github.com/alibabacloud-go/apig-20240327/v5 v5.4.1
	github.com/alibabacloud-go/cas-20200407/v4 v4.0.4
	github.com/alibabacloud-go/cdn-20180510/v9 v9.0.0
	github.com/alibabacloud-go/cloudapi-20160714/v5 v5.7.8
	github.com/alibabacloud-go/darabonba-openapi/v2 v2.1.13
	github.com/alibabacloud-go/dcdn-20180115/v4 v4.0.0
	github.com/alibabacloud-go/ddoscoo-20200101/v4 v4.1.1
	github.com/alibabacloud-go/esa-20240910/v2 v2.42.0
	github.com/alibabacloud-go/fc-20230330/v4 v4.6.6
	github.com/alibabacloud-go/fc-open-20210406/v2 v2.0.12
	github.com/alibabacloud-go/ga-20191120/v3 v3.1.8
	github.com/alibabacloud-go/live-20161101/v2 v2.5.0
	github.com/alibabacloud-go/nlb-20220430/v4 v4.1.0
	github.com/alibabacloud-go/openapi-util v0.1.1
	github.com/alibabacloud-go/slb-20140515/v4 v4.0.13
	github.com/alibabacloud-go/tea v1.3.14
	github.com/alibabacloud-go/tea-utils/v2 v2.0.9
	github.com/alibabacloud-go/vod-20170321/v4 v4.11.0
	github.com/alibabacloud-go/waf-openapi-20211001/v7 v7.1.1
	github.com/aliyun/alibabacloud-oss-go-sdk-v2 v1.3.0
	github.com/aws/aws-sdk-go-v2 v1.41.0
	github.com/aws/aws-sdk-go-v2/config v1.32.6
	github.com/aws/aws-sdk-go-v2/credentials v1.19.6
	github.com/aws/aws-sdk-go-v2/service/acm v1.37.18
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.58.3
	github.com/aws/aws-sdk-go-v2/service/iam v1.53.1
	github.com/baidubce/bce-sdk-go v0.9.256
	github.com/byteplus-sdk/byteplus-sdk-golang v1.0.60
	github.com/go-acme/lego/v4 v4.30.1
	github.com/go-cmd/cmd v1.4.3
	github.com/go-resty/resty/v2 v2.17.1
	github.com/go-viper/mapstructure/v2 v2.4.0
	github.com/google/go-querystring v1.1.0
	github.com/holubovskyi/apisix-client-go v1.4.1
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.1.181
	github.com/jdcloud-api/jdcloud-sdk-go v1.64.0
	github.com/kong/go-kong v0.71.0
	github.com/libdns/dynv6 v1.1.1
	github.com/libdns/libdns v1.1.1
	github.com/luthermonson/go-proxmox v0.2.4
	github.com/minio/minio-go/v7 v7.0.97
	github.com/mohuatech/mohuacloud-go-sdk v0.0.0-20251115182757-6fba4d0a4c47
	github.com/pavlo-v-chernykh/keystore-go/v4 v4.5.0
	github.com/pkg/sftp v1.13.10
	github.com/pocketbase/dbx v1.11.0
	github.com/pocketbase/pocketbase v0.35.0
	github.com/povsister/scp v0.0.0-20250701154629-777cf82de5df
	github.com/qiniu/go-sdk/v7 v7.25.5
	github.com/samber/lo v1.52.0
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn v1.3.20
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb v1.3.13
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.3.20
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/gaap v1.1.45
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/live v1.3.20
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/scf v1.3.4
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl v1.3.20
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo v1.3.18
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vod v1.3.20
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/waf v1.3.20
	github.com/ucloud/ucloud-sdk-go v0.22.58
	github.com/volcengine/ve-tos-golang-sdk/v2 v2.8.0
	github.com/volcengine/volc-sdk-golang v1.0.232
	github.com/volcengine/volcengine-go-sdk v1.1.55
	github.com/wneessen/go-mail v0.7.2
	github.com/xhit/go-str2duration/v2 v2.1.0
	gitlab.ecloud.com/ecloud/ecloudsdkclouddns v1.0.1
	gitlab.ecloud.com/ecloud/ecloudsdkcore v1.0.0
	golang.org/x/crypto v0.46.0
	golang.org/x/sync v0.19.0
	golang.org/x/sys v0.39.0
	golang.org/x/text v0.32.0
	k8s.io/api v0.35.0
	k8s.io/apimachinery v0.35.0
	k8s.io/client-go v0.35.0
	software.sslmate.com/src/go-pkcs12 v0.6.0
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns v1.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph v0.9.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal v1.2.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.6.0 // indirect
	github.com/alibabacloud-go/alibabacloud-gateway-fc-util v0.0.7 // indirect
	github.com/avast/retry-go v3.0.0+incompatible // indirect
	github.com/aws/aws-sdk-go v1.40.45 // indirect
	github.com/aws/aws-sdk-go-v2/service/route53 v1.62.0 // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/buger/goterm v1.0.4 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/diskfs/go-diskfs v1.7.0 // indirect
	github.com/djherbis/times v1.6.0 // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-acme/alidns-20150109/v4 v4.7.0 // indirect
	github.com/go-acme/tencentclouddnspod v1.1.25 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.23.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/goccy/go-yaml v1.9.8 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kong/semver/v4 v4.0.1 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/linode/linodego v1.62.0 // indirect
	github.com/magefile/mage v1.15.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/namedotcom/go/v4 v4.0.2 // indirect
	github.com/nrdcg/bunny-go v0.1.0 // indirect
	github.com/nrdcg/desec v0.11.1 // indirect
	github.com/nrdcg/goacmedns v0.2.0 // indirect
	github.com/nrdcg/porkbun v0.4.0 // indirect
	github.com/ovh/go-ovh v1.9.0 // indirect
	github.com/peterhellberg/link v1.2.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/qiniu/dyn v1.3.0 // indirect
	github.com/qiniu/x v1.10.5 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/sirupsen/logrus v1.9.4-0.20230606125235-dd1b4c2e81af // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/vultr/govultr/v3 v3.26.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.mongodb.org/mongo-driver v1.17.2 // indirect
	go.uber.org/ratelimit v0.3.1 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp v0.0.0-20251219203646-944ab1f22d93 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ns1/ns1-go.v2 v2.16.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250910181357-589584f1c912 // indirect
	k8s.io/utils v0.0.0-20251002143259-bc988d571ff4 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/alibabacloud-go/alibabacloud-gateway-spi v0.0.5 // indirect
	github.com/alibabacloud-go/debug v1.0.1 // indirect
	github.com/alibabacloud-go/endpoint-util v1.1.1 // indirect
	github.com/aliyun/credentials-go v1.4.7 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.16 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.16 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.16 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.5 // indirect
	github.com/aws/smithy-go v1.24.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/clbanning/mxj/v2 v2.7.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/domodwyer/mailyak/v3 v3.6.2 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/ganigeorgiev/fexpr v0.5.0 // indirect
	github.com/go-acme/esa-20240910/v2 v2.40.3 // indirect
	github.com/go-acme/tencentedgdeone v1.1.48 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.13-0.20220915233716-71ac16282d12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/klauspost/crc32 v1.3.0 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/miekg/dns v1.1.69 // indirect
	github.com/minio/crc64nvme v1.1.0 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/nrdcg/namesilo v0.5.0 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/tinylib/msgp v1.3.0 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	golang.org/x/image v0.34.0 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/oauth2 v0.34.0 // indirect
	golang.org/x/term v0.38.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/libc v1.66.10 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.41.0 // indirect
)

replace gitlab.ecloud.com/ecloud/ecloudsdkcore v1.0.0 => ./pkg/forks/gitlab.ecloud.com/ecloud/ecloudsdkcore@v1.0.0

replace gitlab.ecloud.com/ecloud/ecloudsdkclouddns v1.0.1 => ./pkg/forks/gitlab.ecloud.com/ecloud/ecloudsdkclouddns@v1.0.1
