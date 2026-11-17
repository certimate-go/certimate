package accesscheck

import (
	"github.com/certimate-go/certimate/pkg/core"
	cn35impl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/35cn"
	dns51comimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/51dnscom"
	acmecaimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/acmeca"
	acmednsimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/acmedns"
	acmehttpreqimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/acmehttpreq"
	actalissslimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/actalisssl"
	akamaiimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/akamai"
	aliyunimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/aliyun"
	arvancloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/arvancloud"
	awsimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/aws"
	azureimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/azure"
	baiducloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/baiducloud"
	begetimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/beget"
	bunnyimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/bunny"
	byteplusimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/byteplus"
	cfimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/cloudflare"
	cloudnsimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/cloudns"
	cmcccloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/cmcccloud"
	conohavpsimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/conohavps"
	constelliximpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/constellix"
	cpanelimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/cpanel"
	ctcccloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/ctcccloud"
	desecimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/desec"
	digicertimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/digicert"
	digitaloceanimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/digitalocean"
	dnslaimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/dnsla"
	dnsmadeeasyimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/dnsmadeeasy"
	dynadotimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/dynadot"
	dynuimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/dynu"
	dynv6impl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/dynv6"
	ftpimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/ftp"
	gandinetimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/gandinet"
	gcoreimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/gcore"
	globalsignatlasimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/globalsignatlas"
	gnameimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/gname"
	godaddyimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/godaddy"
	googlecloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/googlecloud"
	googletrustservicesimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/googletrustservices"
	hetznerimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/hetzner"
	hostingdeimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/hostingde"
	hostingerimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/hostinger"
	huaweicloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/huaweicloud"
	infomaniakimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/infomaniak"
	ionosimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/ionos"
	jdcloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/jdcloud"
	linodeimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/linode"
	litesslimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/litessl"
	luadnsimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/luadns"
	namecheapimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/namecheap"
	namedotcomimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/namedotcom"
	namesiloimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/namesilo"
	netcupimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/netcup"
	netlifyimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/netlify"
	ns1impl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/ns1"
	oraclecloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/oraclecloud"
	ovhcloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/ovhcloud"
	porkbunimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/porkbun"
	qingcloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/qingcloud"
	rainyunimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/rainyun"
	regruimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/regru"
	rucenterimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/rucenter"
	s3impl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/s3"
	sectigoimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/sectigo"
	simplycomimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/simplycom"
	spaceshipimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/spaceship"
	sshimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/ssh"
	sslcomimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/sslcom"
	technitiumdnsimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/technitiumdns"
	tencentcloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/tencentcloud"
	todaynicimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/todaynic"
	ucloudimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/ucloud"
	vercelimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/vercel"
	volcengineimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/volcengine"
	vultrimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/vultr"
	webhookimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/webhook"
	westcnimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/westcn"
	xinnetimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/xinnet"
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
	registerCoreChecker(arvancloudimpl.New())
	registerCoreChecker(bunnyimpl.New())
	registerCoreChecker(desecimpl.New())
	registerCoreChecker(digitaloceanimpl.New())
	registerCoreChecker(dynuimpl.New())
	registerCoreChecker(dynv6impl.New())
	registerCoreChecker(gandinetimpl.New())
	registerCoreChecker(gcoreimpl.New())
	registerCoreChecker(hetznerimpl.New())
	registerCoreChecker(hostingdeimpl.New())
	registerCoreChecker(hostingerimpl.New())
	registerCoreChecker(infomaniakimpl.New())
	registerCoreChecker(linodeimpl.New())
	registerCoreChecker(namesiloimpl.New())
	registerCoreChecker(netlifyimpl.New())
	registerCoreChecker(ns1impl.New())
	registerCoreChecker(rainyunimpl.New())
	registerCoreChecker(vultrimpl.New())
	registerCoreChecker(aliyunimpl.New())
	registerCoreChecker(tencentcloudimpl.New())
	registerCoreChecker(huaweicloudimpl.New())
	registerCoreChecker(awsimpl.New())
	registerCoreChecker(azureimpl.New())
	registerCoreChecker(googlecloudimpl.New())
	registerCoreChecker(volcengineimpl.New())
	registerCoreChecker(byteplusimpl.New())
	registerCoreChecker(baiducloudimpl.New())
	registerCoreChecker(cn35impl.New())
	registerCoreChecker(dns51comimpl.New())
	registerCoreChecker(begetimpl.New())
	registerCoreChecker(cloudnsimpl.New())
	registerCoreChecker(constelliximpl.New())
	registerCoreChecker(dnslaimpl.New())
	registerCoreChecker(dnsmadeeasyimpl.New())
	registerCoreChecker(dynadotimpl.New())
	registerCoreChecker(gnameimpl.New())
	registerCoreChecker(godaddyimpl.New())
	registerCoreChecker(ionosimpl.New())
	registerCoreChecker(luadnsimpl.New())
	registerCoreChecker(namecheapimpl.New())
	registerCoreChecker(namedotcomimpl.New())
	registerCoreChecker(netcupimpl.New())
	registerCoreChecker(porkbunimpl.New())
	registerCoreChecker(regruimpl.New())
	registerCoreChecker(rucenterimpl.New())
	registerCoreChecker(simplycomimpl.New())
	registerCoreChecker(spaceshipimpl.New())
	registerCoreChecker(todaynicimpl.New())
	registerCoreChecker(vercelimpl.New())
	registerCoreChecker(westcnimpl.New())
	registerCoreChecker(xinnetimpl.New())
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
	registerCoreChecker(akamaiimpl.New())
	registerCoreChecker(conohavpsimpl.New())
	registerCoreChecker(technitiumdnsimpl.New())
	registerCoreChecker(acmednsimpl.New())
	registerCoreChecker(acmehttpreqimpl.New())
	registerCoreChecker(ctcccloudimpl.New())
	registerCoreChecker(jdcloudimpl.New())
	registerCoreChecker(qingcloudimpl.New())
	registerCoreChecker(ucloudimpl.New())
	registerCoreChecker(yandexcloudimpl.New())
	registerCoreChecker(oraclecloudimpl.New())
	registerCoreChecker(ovhcloudimpl.New())
}

func registerCoreChecker(checker core.AccessChecker) {
	if err := DefaultRegistry().Register(&Registration{
		Source:  SourceCore,
		Checker: checker,
	}); err != nil {
		panic(err)
	}
}
