package accesscheck

import (
	"github.com/certimate-go/certimate/pkg/core"

	acmecaimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/acmeca"

	actalissslimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/actalisssl"

	aliyunimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/aliyun"

	awsimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/aws"
	azureimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/azure"
	baiducloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/baiducloud"

	bunnyimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/bunny"
	byteplusimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/byteplus"
	cfimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/cloudflare"

	cmcccloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/cmcccloud"

	cpanelimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/cpanel"
	ctcccloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/ctcccloud"

	digicertimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/digicert"
	digitaloceanimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/digitalocean"

	ftpimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/ftp"

	gcoreimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/gcore"
	globalsignatlasimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/globalsignatlas"

	googlecloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/googlecloud"
	googletrustservicesimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/googletrustservices"

	huaweicloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/huaweicloud"

	jdcloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/jdcloud"
	linodeimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/linode"
	litesslimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/litessl"

	netlifyimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/netlify"

	oraclecloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/oraclecloud"

	qingcloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/qingcloud"
	rainyunimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/rainyun"

	s3impl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/s3"
	sectigoimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/sectigo"

	sshimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/ssh"
	sslcomimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/sslcom"

	tencentcloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/tencentcloud"

	ucloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/ucloud"
	vercelimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/vercel"
	volcengineimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/volcengine"

	webhookimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/webhook"

	yandexcloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/yandexcloud"
	zerosslimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/zerossl"
)

func init() {
	registerCoreChecker(sshimpl.New())
	registerCoreChecker(cfimpl.New())
	registerCoreChecker(ftpimpl.New())
	registerCoreChecker(s3impl.New())
	registerCoreChecker(webhookimpl.New())
	registerCoreChecker(cpanelimpl.New())
	registerCoreChecker(bunnyimpl.New())
	registerCoreChecker(digitaloceanimpl.New())
	registerCoreChecker(gcoreimpl.New())
	registerCoreChecker(linodeimpl.New())
	registerCoreChecker(netlifyimpl.New())
	registerCoreChecker(rainyunimpl.New())
	registerCoreChecker(aliyunimpl.New())
	registerCoreChecker(tencentcloudimpl.New())
	registerCoreChecker(huaweicloudimpl.New())
	registerCoreChecker(awsimpl.New())
	registerCoreChecker(azureimpl.New())
	registerCoreChecker(googlecloudimpl.New())
	registerCoreChecker(volcengineimpl.New())
	registerCoreChecker(byteplusimpl.New())
	registerCoreChecker(baiducloudimpl.New())
	registerCoreChecker(vercelimpl.New())
	registerCoreChecker(acmecaimpl.New())
	registerCoreChecker(actalissslimpl.New())
	registerCoreChecker(digicertimpl.New())
	registerCoreChecker(globalsignatlasimpl.New())
	registerCoreChecker(googletrustservicesimpl.New())
	registerCoreChecker(litesslimpl.New())
	registerCoreChecker(sectigoimpl.New())
	registerCoreChecker(sslcomimpl.New())
	registerCoreChecker(zerosslimpl.New())
	registerCoreChecker(cmcccloudimpl.New())
	registerCoreChecker(ctcccloudimpl.New())
	registerCoreChecker(jdcloudimpl.New())
	registerCoreChecker(qingcloudimpl.New())
	registerCoreChecker(ucloudimpl.New())
	registerCoreChecker(yandexcloudimpl.New())
	registerCoreChecker(oraclecloudimpl.New())
}

func registerCoreChecker(checker core.AccessChecker) {
	if err := DefaultRegistry().Register(&Registration{
		Source:  SourceCore,
		Checker: checker,
	}); err != nil {
		panic(err)
	}
}
