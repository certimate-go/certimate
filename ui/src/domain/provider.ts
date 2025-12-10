interface BaseProvider<P> {
  type: P;
  name: string;
  icon: string;
  builtin: boolean;
}

interface BaseProviderWithAccess<P> extends BaseProvider<P> {
  provider: AccessProviderType;
}

// #region AccessProvider
/*
  注意：如果追加新的常量值，请保持以 ASCII 排序。
  NOTICE: If you add new constant, please keep ASCII order.
 */
export const ACCESS_PROVIDERS = Object.freeze({
  ["1PANEL"]: "1panel",
  ["35CN"]: "35cn",
  ["51DNSCOM"]: "51dnscom",
  ACMECA: "acmeca",
  ACMEDNS: "acmedns",
  ACMEHTTPREQ: "acmehttpreq",
  ACTALISSSL: "actalisssl",
  AKAMAI: "akamai",
  ALIYUN: "aliyun",
  APISIX: "apisix",
  ARVANCLOUD: "arvancloud",
  AWS: "aws",
  AZURE: "azure",
  BAIDUCLOUD: "baiducloud",
  BAISHAN: "baishan",
  BAOTAPANEL: "baotapanel",
  BAOTAPANELGO: "baotapanelgo",
  BAOTAWAF: "baotawaf",
  BOOKMYNAME: "bookmyname",
  BUNNY: "bunny",
  BYTEPLUS: "byteplus",
  CACHEFLY: "cachefly",
  CDNFLY: "cdnfly",
  CLOUDFLARE: "cloudflare",
  CLOUDNS: "cloudns",
  CMCCCLOUD: "cmcccloud",
  CONSTELLIX: "constellix",
  CPANEL: "cpanel",
  CTCCCLOUD: "ctcccloud",
  DESEC: "desec",
  DIGITALOCEAN: "digitalocean",
  DINGTALKBOT: "dingtalkbot",
  DISCORDBOT: "discordbot",
  DNSEXIT: "dnsexit",
  DNSLA: "dnsla",
  DNSMADEEASY: "dnsmadeeasy",
  DOGECLOUD: "dogecloud",
  DUCKDNS: "duckdns",
  DYNU: "dynu",
  DYNV6: "dynv6",
  EMAIL: "email",
  FLEXCDN: "flexcdn",
  GANDINET: "gandinet",
  GCORE: "gcore",
  GLOBALSIGNATLAS: "globalsignatlas",
  GNAME: "gname",
  GODADDY: "godaddy",
  GOEDGE: "goedge",
  GOOGLETRUSTSERVICES: "googletrustservices",
  HETZNER: "hetzner",
  HOSTINGDE: "hostingde",
  HOSTINGER: "hostinger",
  HUAWEICLOUD: "huaweicloud",
  INFOMANIAK: "infomaniak",
  IONOS: "ionos",
  JDCLOUD: "jdcloud",
  KONG: "kong",
  KUBERNETES: "k8s",
  KSYUN: "ksyun",
  LARKBOT: "larkbot",
  LECDN: "lecdn",
  LETSENCRYPT: "letsencrypt",
  LETSENCRYPTSTAGING: "letsencryptstaging",
  LINODE: "linode",
  LITESSL: "litessl",
  LOCAL: "local",
  MATTERMOST: "mattermost",
  MOHUA: "mohua",
  NAMECHEAP: "namecheap",
  NAMEDOTCOM: "namedotcom",
  NAMESILO: "namesilo",
  NETCUP: "netcup",
  NETLIFY: "netlify",
  NS1: "ns1",
  OVHCLOUD: "ovhcloud",
  PORKBUN: "porkbun",
  POWERDNS: "powerdns",
  PROXMOXVE: "proxmoxve",
  QINGCLOUD: "qingcloud",
  QINIU: "qiniu",
  RAINYUN: "rainyun",
  RATPANEL: "ratpanel",
  RFC2136: "rfc2136",
  SAFELINE: "safeline",
  SECTIGO: "sectigo",
  SLACKBOT: "slackbot",
  SPACESHIP: "spaceship",
  SSH: "ssh",
  SSLCOM: "sslcom",
  TECHNITIUMDNS: "technitiumdns",
  TELEGRAMBOT: "telegrambot",
  TENCENTCLOUD: "tencentcloud",
  UCLOUD: "ucloud",
  UNICLOUD: "unicloud",
  UPYUN: "upyun",
  VERCEL: "vercel",
  VOLCENGINE: "volcengine",
  VULTR: "vultr",
  WANGSU: "wangsu",
  WEBHOOK: "webhook",
  WECOMBOT: "wecombot",
  WESTCN: "westcn",
  XINNET: "xinnet",
  ZEROSSL: "zerossl",
} as const);

export type AccessProviderType = (typeof ACCESS_PROVIDERS)[keyof typeof ACCESS_PROVIDERS];

export const ACCESS_USAGES = Object.freeze({
  DNS: "dns",
  HOSTING: "hosting",
  CA: "ca",
  NOTIFICATION: "notification",
} as const);

export type AccessUsageType = (typeof ACCESS_USAGES)[keyof typeof ACCESS_USAGES];

export interface AccessProvider extends BaseProvider<AccessProviderType> {
  usages: AccessUsageType[];
}

export const accessProvidersMap: Map<AccessProvider["type"] | string, AccessProvider> = new Map(
  /*
    注意：此处的顺序决定显示在前端的顺序。
    NOTICE: The following order determines the order displayed at the frontend.
  */
  (
    [
      [ACCESS_PROVIDERS.LOCAL, "provider.local", "/imgs/providers/local.svg", [ACCESS_USAGES.HOSTING], "builtin"],
      [ACCESS_PROVIDERS.SSH, "provider.ssh", "/imgs/providers/ssh.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.WEBHOOK, "provider.webhook", "/imgs/providers/webhook.svg", [ACCESS_USAGES.HOSTING, ACCESS_USAGES.NOTIFICATION]],
      [ACCESS_PROVIDERS.KUBERNETES, "provider.kubernetes", "/imgs/providers/kubernetes.svg", [ACCESS_USAGES.HOSTING]],

      [ACCESS_PROVIDERS.ALIYUN, "provider.aliyun", "/imgs/providers/aliyun.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.TENCENTCLOUD, "provider.tencentcloud", "/imgs/providers/tencentcloud.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.BAIDUCLOUD, "provider.baiducloud", "/imgs/providers/baiducloud.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.HUAWEICLOUD, "provider.huaweicloud", "/imgs/providers/huaweicloud.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.VOLCENGINE, "provider.volcengine", "/imgs/providers/volcengine.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.JDCLOUD, "provider.jdcloud", "/imgs/providers/jdcloud.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.AWS, "provider.aws", "/imgs/providers/aws.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.AZURE, "provider.azure", "/imgs/providers/azure.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.BUNNY, "provider.bunny", "/imgs/providers/bunny.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.GCORE, "provider.gcore", "/imgs/providers/gcore.png", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.NETLIFY, "provider.netlify", "/imgs/providers/netlify.png", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.RAINYUN, "provider.rainyun", "/imgs/providers/rainyun.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.UCLOUD, "provider.ucloud", "/imgs/providers/ucloud.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.CTCCCLOUD, "provider.ctcccloud", "/imgs/providers/ctcccloud.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.CPANEL, "provider.cpanel", "/imgs/providers/cpanel.svg", [ACCESS_USAGES.DNS, ACCESS_USAGES.HOSTING]],

      [ACCESS_PROVIDERS.QINIU, "provider.qiniu", "/imgs/providers/qiniu.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.UPYUN, "provider.upyun", "/imgs/providers/upyun.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.BAISHAN, "provider.baishan", "/imgs/providers/baishan.png", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.WANGSU, "provider.wangsu", "/imgs/providers/wangsu.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.DOGECLOUD, "provider.dogecloud", "/imgs/providers/dogecloud.png", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.KSYUN, "provider.ksyun", "/imgs/providers/ksyun.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.BYTEPLUS, "provider.byteplus", "/imgs/providers/byteplus.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.UNICLOUD, "provider.unicloud", "/imgs/providers/unicloud.png", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.MOHUA, "provider.mohua", "/imgs/providers/mohua.png", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS["1PANEL"], "provider.1panel", "/imgs/providers/1panel.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.BAOTAPANEL, "provider.baotapanel", "/imgs/providers/baotapanel.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.BAOTAPANELGO, "provider.baotapanelgo", "/imgs/providers/baotapanel.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.BAOTAWAF, "provider.baotawaf", "/imgs/providers/baotawaf.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.RATPANEL, "provider.ratpanel", "/imgs/providers/ratpanel.png", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.SAFELINE, "provider.safeline", "/imgs/providers/safeline.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.CDNFLY, "provider.cdnfly", "/imgs/providers/cdnfly.png", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.FLEXCDN, "provider.flexcdn", "/imgs/providers/flexcdn.png", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.GOEDGE, "provider.goedge", "/imgs/providers/goedge.png", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.LECDN, "provider.lecdn", "/imgs/providers/lecdn.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.CACHEFLY, "provider.cachefly", "/imgs/providers/cachefly.png", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.APISIX, "provider.apisix", "/imgs/providers/apisix.svg", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.KONG, "provider.kong", "/imgs/providers/kong.png", [ACCESS_USAGES.HOSTING]],
      [ACCESS_PROVIDERS.PROXMOXVE, "provider.proxmoxve", "/imgs/providers/proxmoxve.svg", [ACCESS_USAGES.HOSTING]],

      [ACCESS_PROVIDERS.AKAMAI, "provider.akamai", "/imgs/providers/akamai.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.ARVANCLOUD, "provider.arvancloud", "/imgs/providers/arvancloud.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.BOOKMYNAME, "provider.bookmyname", "/imgs/providers/bookmyname.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.CLOUDFLARE, "provider.cloudflare", "/imgs/providers/cloudflare.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.CLOUDNS, "provider.cloudns", "/imgs/providers/cloudns.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.CONSTELLIX, "provider.constellix", "/imgs/providers/constellix.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.DESEC, "provider.desec", "/imgs/providers/desec.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.DIGITALOCEAN, "provider.digitalocean", "/imgs/providers/digitalocean.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.DNSEXIT, "provider.dnsexit", "/imgs/providers/dnsexit.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.DNSMADEEASY, "provider.dnsmadeeasy", "/imgs/providers/dnsmadeeasy.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.DUCKDNS, "provider.duckdns", "/imgs/providers/duckdns.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.DYNU, "provider.dynu", "/imgs/providers/dynu.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.DYNV6, "provider.dynv6", "/imgs/providers/dynv6.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.GANDINET, "provider.gandinet", "/imgs/providers/gandinet.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.GNAME, "provider.gname", "/imgs/providers/gname.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.GODADDY, "provider.godaddy", "/imgs/providers/godaddy.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.HETZNER, "provider.hetzner", "/imgs/providers/hetzner.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.HOSTINGDE, "provider.hostingde", "/imgs/providers/hostingde.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.HOSTINGER, "provider.hostinger", "/imgs/providers/hostinger.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.INFOMANIAK, "provider.infomaniak", "/imgs/providers/infomaniak.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.IONOS, "provider.ionos", "/imgs/providers/ionos.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.LINODE, "provider.linode", "/imgs/providers/linode.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.NAMECHEAP, "provider.namecheap", "/imgs/providers/namecheap.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.NAMEDOTCOM, "provider.namedotcom", "/imgs/providers/namedotcom.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.NAMESILO, "provider.namesilo", "/imgs/providers/namesilo.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.NETCUP, "provider.netcup", "/imgs/providers/netcup.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.NS1, "provider.ns1", "/imgs/providers/ns1.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.OVHCLOUD, "provider.ovhcloud", "/imgs/providers/ovhcloud.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.PORKBUN, "provider.porkbun", "/imgs/providers/porkbun.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.SPACESHIP, "provider.spaceship", "/imgs/providers/spaceship.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.VERCEL, "provider.vercel", "/imgs/providers/vercel.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.VULTR, "provider.vultr", "/imgs/providers/vultr.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.CMCCCLOUD, "provider.cmcccloud", "/imgs/providers/cmcccloud.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.QINGCLOUD, "provider.qingcloud", "/imgs/providers/qingcloud.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.WESTCN, "provider.westcn", "/imgs/providers/westcn.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS["35CN"], "provider.35cn", "/imgs/providers/35cn.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.XINNET, "provider.xinnet", "/imgs/providers/xinnet.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS["51DNSCOM"], "provider.51dnscom", "/imgs/providers/51dnscom.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.DNSLA, "provider.dnsla", "/imgs/providers/dnsla.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.POWERDNS, "provider.powerdns", "/imgs/providers/powerdns.svg", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.TECHNITIUMDNS, "provider.technitiumdns", "/imgs/providers/technitiumdns.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.RFC2136, "provider.rfc2136", "/imgs/providers/rfc.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.ACMEDNS, "provider.acmedns", "/imgs/providers/acmedns.png", [ACCESS_USAGES.DNS]],
      [ACCESS_PROVIDERS.ACMEHTTPREQ, "provider.acmehttpreq", "/imgs/providers/acmehttpreq.svg", [ACCESS_USAGES.DNS]],

      [ACCESS_PROVIDERS.LETSENCRYPT, "provider.letsencrypt", "/imgs/providers/letsencrypt.svg", [ACCESS_USAGES.CA], "builtin"],
      [ACCESS_PROVIDERS.LETSENCRYPTSTAGING, "provider.letsencryptstaging", "/imgs/providers/letsencrypt.svg", [ACCESS_USAGES.CA], "builtin"],
      [ACCESS_PROVIDERS.ACTALISSSL, "provider.actalisssl", "/imgs/providers/actalisssl.png", [ACCESS_USAGES.CA]],
      [ACCESS_PROVIDERS.GLOBALSIGNATLAS, "provider.globalsignatlas", "/imgs/providers/globalsignatlas.png", [ACCESS_USAGES.CA]],
      [ACCESS_PROVIDERS.GOOGLETRUSTSERVICES, "provider.googletrustservices", "/imgs/providers/google.svg", [ACCESS_USAGES.CA]],
      [ACCESS_PROVIDERS.LITESSL, "provider.litessl", "/imgs/providers/litessl.svg", [ACCESS_USAGES.CA]],
      [ACCESS_PROVIDERS.SECTIGO, "provider.sectigo", "/imgs/providers/sectigo.svg", [ACCESS_USAGES.CA]],
      [ACCESS_PROVIDERS.SSLCOM, "provider.sslcom", "/imgs/providers/sslcom.svg", [ACCESS_USAGES.CA]],
      [ACCESS_PROVIDERS.ZEROSSL, "provider.zerossl", "/imgs/providers/zerossl.svg", [ACCESS_USAGES.CA]],
      [ACCESS_PROVIDERS.ACMECA, "provider.acmeca", "/imgs/providers/acmeca.svg", [ACCESS_USAGES.CA]],

      [ACCESS_PROVIDERS.EMAIL, "provider.email", "/imgs/providers/email.svg", [ACCESS_USAGES.NOTIFICATION]],
      [ACCESS_PROVIDERS.DINGTALKBOT, "provider.dingtalkbot", "/imgs/providers/dingtalk.svg", [ACCESS_USAGES.NOTIFICATION]],
      [ACCESS_PROVIDERS.LARKBOT, "provider.larkbot", "/imgs/providers/lark.svg", [ACCESS_USAGES.NOTIFICATION]],
      [ACCESS_PROVIDERS.WECOMBOT, "provider.wecombot", "/imgs/providers/wecom.svg", [ACCESS_USAGES.NOTIFICATION]],
      [ACCESS_PROVIDERS.DISCORDBOT, "provider.discordbot", "/imgs/providers/discord.svg", [ACCESS_USAGES.NOTIFICATION]],
      [ACCESS_PROVIDERS.SLACKBOT, "provider.slackbot", "/imgs/providers/slack.svg", [ACCESS_USAGES.NOTIFICATION]],
      [ACCESS_PROVIDERS.TELEGRAMBOT, "provider.telegrambot", "/imgs/providers/telegram.svg", [ACCESS_USAGES.NOTIFICATION]],
      [ACCESS_PROVIDERS.MATTERMOST, "provider.mattermost", "/imgs/providers/mattermost.svg", [ACCESS_USAGES.NOTIFICATION]],
    ] satisfies Array<[AccessProviderType, string, string, AccessUsageType[], "builtin"] | [AccessProviderType, string, string, AccessUsageType[]]>
  ).map(([type, name, icon, usages, builtin]) => [
    type,
    {
      type: type,
      name: name,
      icon: icon,
      usages: usages,
      builtin: builtin === "builtin",
    },
  ])
);
// #endregion

// #region CAProvider
/*
  注意：如果追加新的常量值，请保持以 ASCII 排序。
  NOTICE: If you add new constant, please keep ASCII order.
 */
export const CA_PROVIDERS = Object.freeze({
  ACMECA: `${ACCESS_PROVIDERS.ACMECA}`,
  ACTALISSSL: `${ACCESS_PROVIDERS.ACTALISSSL}`,
  GLOBALSIGNATLAS: `${ACCESS_PROVIDERS.GLOBALSIGNATLAS}`,
  GOOGLETRUSTSERVICES: `${ACCESS_PROVIDERS.GOOGLETRUSTSERVICES}`,
  LETSENCRYPT: `${ACCESS_PROVIDERS.LETSENCRYPT}`,
  LETSENCRYPTSTAGING: `${ACCESS_PROVIDERS.LETSENCRYPTSTAGING}`,
  LITESSL: `${ACCESS_PROVIDERS.LITESSL}`,
  SECTIGO: `${ACCESS_PROVIDERS.SECTIGO}`,
  SSLCOM: `${ACCESS_PROVIDERS.SSLCOM}`,
  ZEROSSL: `${ACCESS_PROVIDERS.ZEROSSL}`,
} as const);

export type CAProviderType = (typeof CA_PROVIDERS)[keyof typeof CA_PROVIDERS];

export interface CAProvider extends BaseProviderWithAccess<CAProviderType> {}

export const caProvidersMap: Map<CAProvider["type"] | string, CAProvider> = new Map(
  /*
    注意：此处的顺序决定显示在前端的顺序。
    NOTICE: The following order determines the order displayed at the frontend.
  */
  (
    [
      [CA_PROVIDERS.LETSENCRYPT, "builtin"],
      [CA_PROVIDERS.LETSENCRYPTSTAGING, "builtin"],
      [CA_PROVIDERS.ACTALISSSL],
      [CA_PROVIDERS.GLOBALSIGNATLAS],
      [CA_PROVIDERS.GOOGLETRUSTSERVICES],
      [CA_PROVIDERS.SECTIGO],
      [CA_PROVIDERS.SSLCOM],
      [CA_PROVIDERS.ZEROSSL],
      [CA_PROVIDERS.LITESSL],
      [CA_PROVIDERS.ACMECA],
    ] satisfies Array<[CAProviderType, "builtin"] | [CAProviderType]>
  ).map(([type, builtin]) => [
    type,
    {
      type: type,
      name: accessProvidersMap.get(type.split("-")[0])!.name,
      icon: accessProvidersMap.get(type.split("-")[0])!.icon,
      provider: type.split("-")[0] as AccessProviderType,
      builtin: builtin === "builtin",
    },
  ])
);
// #endregion

// #region ACMEDNS01Provider
/*
  注意：如果追加新的常量值，请保持以 ASCII 排序。
  NOTICE: If you add new constant, please keep ASCII order.
 */
export const ACME_DNS01_PROVIDERS = Object.freeze({
  ["35CN"]: `${ACCESS_PROVIDERS["35CN"]}`,
  ["51DNSCOM"]: `${ACCESS_PROVIDERS["51DNSCOM"]}`,
  ACMEDNS: `${ACCESS_PROVIDERS.ACMEDNS}`,
  ACMEHTTPREQ: `${ACCESS_PROVIDERS.ACMEHTTPREQ}`,
  AKAMAI: `${ACCESS_PROVIDERS.AKAMAI}`, // 兼容旧值，等同于 `AKAMAI_EDGEDNS`
  AKAMAI_EDGEDNS: `${ACCESS_PROVIDERS.AKAMAI}-edgedns`,
  ALIYUN: `${ACCESS_PROVIDERS.ALIYUN}`, // 兼容旧值，等同于 `ALIYUN_DNS`
  ALIYUN_DNS: `${ACCESS_PROVIDERS.ALIYUN}-dns`,
  ALIYUN_ESA: `${ACCESS_PROVIDERS.ALIYUN}-esa`,
  ARVANCLOUD: `${ACCESS_PROVIDERS.ARVANCLOUD}`,
  AWS: `${ACCESS_PROVIDERS.AWS}`, // 兼容旧值，等同于 `AWS_ROUTE53`
  AWS_ROUTE53: `${ACCESS_PROVIDERS.AWS}-route53`,
  AZURE: `${ACCESS_PROVIDERS.AZURE}`, // 兼容旧值，等同于 `AZURE_DNS`
  AZURE_DNS: `${ACCESS_PROVIDERS.AZURE}-dns`,
  BAIDUCLOUD: `${ACCESS_PROVIDERS.BAIDUCLOUD}`, // 兼容旧值，等同于 `BAIDUCLOUD_DNS`
  BAIDUCLOUD_DNS: `${ACCESS_PROVIDERS.BAIDUCLOUD}-dns`,
  BOOKMYNAME: `${ACCESS_PROVIDERS.BOOKMYNAME}`,
  BUNNY: `${ACCESS_PROVIDERS.BUNNY}`,
  CLOUDFLARE: `${ACCESS_PROVIDERS.CLOUDFLARE}`,
  CLOUDNS: `${ACCESS_PROVIDERS.CLOUDNS}`,
  CMCCCLOUD: `${ACCESS_PROVIDERS.CMCCCLOUD}`, // 兼容旧值，等同于 `CMCCCLOUD_DNS`
  CMCCCLOUD_DNS: `${ACCESS_PROVIDERS.CMCCCLOUD}-dns`,
  CONSTELLIX: `${ACCESS_PROVIDERS.CONSTELLIX}`,
  CPANEL: `${ACCESS_PROVIDERS.CPANEL}`,
  CTCCCLOUD: `${ACCESS_PROVIDERS.CTCCCLOUD}`, // 兼容旧值，等同于 `CTCCCLOUD_SMARTDNS`
  CTCCCLOUD_SMARTDNS: `${ACCESS_PROVIDERS.CTCCCLOUD}-smartdns`,
  DESEC: `${ACCESS_PROVIDERS.DESEC}`,
  DIGITALOCEAN: `${ACCESS_PROVIDERS.DIGITALOCEAN}`,
  DNSEXIT: `${ACCESS_PROVIDERS.DNSEXIT}`,
  DNSLA: `${ACCESS_PROVIDERS.DNSLA}`,
  DNSMADEEASY: `${ACCESS_PROVIDERS.DNSMADEEASY}`,
  DUCKDNS: `${ACCESS_PROVIDERS.DUCKDNS}`,
  DYNU: `${ACCESS_PROVIDERS.DYNU}`,
  DYNV6: `${ACCESS_PROVIDERS.DYNV6}`,
  GANDINET: `${ACCESS_PROVIDERS.GANDINET}`,
  GCORE: `${ACCESS_PROVIDERS.GCORE}`,
  GNAME: `${ACCESS_PROVIDERS.GNAME}`,
  GODADDY: `${ACCESS_PROVIDERS.GODADDY}`,
  HETZNER: `${ACCESS_PROVIDERS.HETZNER}`,
  HOSTINGDE: `${ACCESS_PROVIDERS.HOSTINGDE}`,
  HOSTINGER: `${ACCESS_PROVIDERS.HOSTINGER}`,
  HUAWEICLOUD: `${ACCESS_PROVIDERS.HUAWEICLOUD}`, // 兼容旧值，等同于 `HUAWEICLOUD_DNS`
  HUAWEICLOUD_DNS: `${ACCESS_PROVIDERS.HUAWEICLOUD}-dns`,
  INFOMANIAK: `${ACCESS_PROVIDERS.INFOMANIAK}`,
  IONOS: `${ACCESS_PROVIDERS.IONOS}`,
  JDCLOUD: `${ACCESS_PROVIDERS.JDCLOUD}`, // 兼容旧值，等同于 `JDCLOUD_DNS`
  JDCLOUD_DNS: `${ACCESS_PROVIDERS.JDCLOUD}-dns`,
  LINODE: `${ACCESS_PROVIDERS.LINODE}`,
  NAMECHEAP: `${ACCESS_PROVIDERS.NAMECHEAP}`,
  NAMEDOTCOM: `${ACCESS_PROVIDERS.NAMEDOTCOM}`,
  NAMESILO: `${ACCESS_PROVIDERS.NAMESILO}`,
  NETCUP: `${ACCESS_PROVIDERS.NETCUP}`,
  NETLIFY: `${ACCESS_PROVIDERS.NETLIFY}`,
  NS1: `${ACCESS_PROVIDERS.NS1}`,
  OVHCLOUD: `${ACCESS_PROVIDERS.OVHCLOUD}`,
  PORKBUN: `${ACCESS_PROVIDERS.PORKBUN}`,
  POWERDNS: `${ACCESS_PROVIDERS.POWERDNS}`,
  QINGCLOUD: `${ACCESS_PROVIDERS.QINGCLOUD}`, // 兼容旧值，等同于 `QINGCLOUD_DNS`
  QINGCLOUD_DNS: `${ACCESS_PROVIDERS.QINGCLOUD}-dns`,
  RAINYUN: `${ACCESS_PROVIDERS.RAINYUN}`,
  RFC2136: `${ACCESS_PROVIDERS.RFC2136}`,
  SPACESHIP: `${ACCESS_PROVIDERS.SPACESHIP}`,
  UCLOUD: `${ACCESS_PROVIDERS.UCLOUD}`, // 兼容旧值，等同于 `UCLOUD_UDNR`
  UCLOUD_UDNR: `${ACCESS_PROVIDERS.UCLOUD}-udnr`,
  TECHNITIUMDNS: `${ACCESS_PROVIDERS.TECHNITIUMDNS}`,
  TENCENTCLOUD: `${ACCESS_PROVIDERS.TENCENTCLOUD}`, // 兼容旧值，等同于 `TENCENTCLOUD_DNS`
  TENCENTCLOUD_DNS: `${ACCESS_PROVIDERS.TENCENTCLOUD}-dns`,
  TENCENTCLOUD_EO: `${ACCESS_PROVIDERS.TENCENTCLOUD}-eo`,
  VERCEL: `${ACCESS_PROVIDERS.VERCEL}`,
  VOLCENGINE: `${ACCESS_PROVIDERS.VOLCENGINE}`, // 兼容旧值，等同于 `VOLCENGINE_DNS`
  VOLCENGINE_DNS: `${ACCESS_PROVIDERS.VOLCENGINE}-dns`,
  VULTR: `${ACCESS_PROVIDERS.VULTR}`,
  WESTCN: `${ACCESS_PROVIDERS.WESTCN}`,
  XINNET: `${ACCESS_PROVIDERS.XINNET}`,
} as const);

export type ACMEDns01ProviderType = (typeof ACME_DNS01_PROVIDERS)[keyof typeof ACME_DNS01_PROVIDERS];

export interface ACMEDns01Provider extends BaseProviderWithAccess<ACMEDns01ProviderType> {}

export const acmeDns01ProvidersMap: Map<ACMEDns01Provider["type"] | string, ACMEDns01Provider> = new Map(
  /*
    注意：此处的顺序决定显示在前端的顺序。
    NOTICE: The following order determines the order displayed at the frontend.
   */
  (
    [
      [ACME_DNS01_PROVIDERS.ALIYUN_DNS, "provider.aliyun_dns"],
      [ACME_DNS01_PROVIDERS.ALIYUN_ESA, "provider.aliyun_esa"],
      [ACME_DNS01_PROVIDERS.TENCENTCLOUD_DNS, "provider.tencentcloud_dns"],
      [ACME_DNS01_PROVIDERS.TENCENTCLOUD_EO, "provider.tencentcloud_eo"],
      [ACME_DNS01_PROVIDERS.BAIDUCLOUD_DNS, "provider.baiducloud_dns"],
      [ACME_DNS01_PROVIDERS.HUAWEICLOUD_DNS, "provider.huaweicloud_dns"],
      [ACME_DNS01_PROVIDERS.VOLCENGINE_DNS, "provider.volcengine_dns"],
      [ACME_DNS01_PROVIDERS.JDCLOUD_DNS, "provider.jdcloud_dns"],
      [ACME_DNS01_PROVIDERS.AWS_ROUTE53, "provider.aws_route53"],
      [ACME_DNS01_PROVIDERS.AZURE_DNS, "provider.azure_dns"],
      [ACME_DNS01_PROVIDERS.AKAMAI_EDGEDNS, "provider.akamai_edgedns"],
      [ACME_DNS01_PROVIDERS.ARVANCLOUD, "provider.arvancloud"],
      [ACME_DNS01_PROVIDERS.BOOKMYNAME, "provider.bookmyname"],
      [ACME_DNS01_PROVIDERS.BUNNY, "provider.bunny"],
      [ACME_DNS01_PROVIDERS.CLOUDFLARE, "provider.cloudflare"],
      [ACME_DNS01_PROVIDERS.CLOUDNS, "provider.cloudns"],
      [ACME_DNS01_PROVIDERS.CONSTELLIX, "provider.constellix"],
      [ACME_DNS01_PROVIDERS.DESEC, "provider.desec"],
      [ACME_DNS01_PROVIDERS.DIGITALOCEAN, "provider.digitalocean"],
      [ACME_DNS01_PROVIDERS.DNSEXIT, "provider.dnsexit"],
      [ACME_DNS01_PROVIDERS.DNSMADEEASY, "provider.dnsmadeeasy"],
      [ACME_DNS01_PROVIDERS.DUCKDNS, "provider.duckdns"],
      [ACME_DNS01_PROVIDERS.DYNU, "provider.dynu"],
      [ACME_DNS01_PROVIDERS.DYNV6, "provider.dynv6"],
      [ACME_DNS01_PROVIDERS.GANDINET, "provider.gandinet"],
      [ACME_DNS01_PROVIDERS.GCORE, "provider.gcore"],
      [ACME_DNS01_PROVIDERS.GNAME, "provider.gname"],
      [ACME_DNS01_PROVIDERS.GODADDY, "provider.godaddy"],
      [ACME_DNS01_PROVIDERS.HETZNER, "provider.hetzner"],
      [ACME_DNS01_PROVIDERS.HOSTINGDE, "provider.hostingde"],
      [ACME_DNS01_PROVIDERS.HOSTINGER, "provider.hostinger"],
      [ACME_DNS01_PROVIDERS.INFOMANIAK, "provider.infomaniak"],
      [ACME_DNS01_PROVIDERS.IONOS, "provider.ionos"],
      [ACME_DNS01_PROVIDERS.LINODE, "provider.linode"],
      [ACME_DNS01_PROVIDERS.NAMECHEAP, "provider.namecheap"],
      [ACME_DNS01_PROVIDERS.NAMEDOTCOM, "provider.namedotcom"],
      [ACME_DNS01_PROVIDERS.NAMESILO, "provider.namesilo"],
      [ACME_DNS01_PROVIDERS.NETCUP, "provider.netcup"],
      [ACME_DNS01_PROVIDERS.NETLIFY, "provider.netlify"],
      [ACME_DNS01_PROVIDERS.NS1, "provider.ns1"],
      [ACME_DNS01_PROVIDERS.OVHCLOUD, "provider.ovhcloud"],
      [ACME_DNS01_PROVIDERS.PORKBUN, "provider.porkbun"],
      [ACME_DNS01_PROVIDERS.SPACESHIP, "provider.spaceship"],
      [ACME_DNS01_PROVIDERS.VERCEL, "provider.vercel"],
      [ACME_DNS01_PROVIDERS.VULTR, "provider.vultr"],
      [ACME_DNS01_PROVIDERS.CMCCCLOUD_DNS, "provider.cmcccloud_dns"],
      [ACME_DNS01_PROVIDERS.CTCCCLOUD_SMARTDNS, "provider.ctcccloud_smartdns"],
      [ACME_DNS01_PROVIDERS.RAINYUN, "provider.rainyun"],
      [ACME_DNS01_PROVIDERS.UCLOUD_UDNR, "provider.ucloud_udnr"],
      [ACME_DNS01_PROVIDERS.QINGCLOUD_DNS, "provider.qingcloud_dns"],
      [ACME_DNS01_PROVIDERS.WESTCN, "provider.westcn"],
      [ACME_DNS01_PROVIDERS["35CN"], "provider.35cn"],
      [ACME_DNS01_PROVIDERS.XINNET, "provider.xinnet"],
      [ACME_DNS01_PROVIDERS["51DNSCOM"], "provider.51dnscom"],
      [ACME_DNS01_PROVIDERS.DNSLA, "provider.dnsla"],
      [ACME_DNS01_PROVIDERS.CPANEL, "provider.cpanel"],
      [ACME_DNS01_PROVIDERS.POWERDNS, "provider.powerdns"],
      [ACME_DNS01_PROVIDERS.TECHNITIUMDNS, "provider.technitiumdns"],
      [ACME_DNS01_PROVIDERS.RFC2136, "provider.rfc2136"],
      [ACME_DNS01_PROVIDERS.ACMEDNS, "provider.acmedns"],
      [ACME_DNS01_PROVIDERS.ACMEHTTPREQ, "provider.acmehttpreq"],
    ] satisfies Array<[ACMEDns01ProviderType, string]>
  ).map(([type, name]) => [
    type,
    {
      type: type,
      name: name,
      icon: accessProvidersMap.get(type.split("-")[0])!.icon,
      provider: type.split("-")[0] as AccessProviderType,
      builtin: false,
    },
  ])
);
// #endregion

// #region ACMEHTTP01Provider
/*
  注意：如果追加新的常量值，请保持以 ASCII 排序。
  NOTICE: If you add new constant, please keep ASCII order.
 */
export const ACME_HTTP01_PROVIDERS = Object.freeze({
  LOCAL: `${ACCESS_PROVIDERS.LOCAL}`,
  SSH: `${ACCESS_PROVIDERS.SSH}`,
} as const);

export type ACMEHttp01ProviderType = (typeof ACME_HTTP01_PROVIDERS)[keyof typeof ACME_HTTP01_PROVIDERS];

export interface ACMEHttp01Provider extends BaseProviderWithAccess<ACMEHttp01ProviderType> {}

export const acmeHttp01ProvidersMap: Map<ACMEHttp01Provider["type"] | string, ACMEHttp01Provider> = new Map(
  /*
    注意：此处的顺序决定显示在前端的顺序。
    NOTICE: The following order determines the order displayed at the frontend.
   */
  (
    [
      [ACME_HTTP01_PROVIDERS.LOCAL, "provider.local", "builtin"],
      [ACME_HTTP01_PROVIDERS.SSH, "provider.ssh"],
    ] satisfies Array<[ACMEHttp01ProviderType, string, "builtin"] | [ACMEHttp01ProviderType, string]>
  ).map(([type, name, builtin]) => [
    type,
    {
      type: type,
      name: name,
      icon: accessProvidersMap.get(type.split("-")[0])!.icon,
      provider: type.split("-")[0] as AccessProviderType,
      builtin: builtin === "builtin",
    },
  ])
);
// #endregion

// #region DeploymentProvider
/*
  注意：如果追加新的常量值，请保持以 ASCII 排序。
  NOTICE: If you add new constant, please keep ASCII order.
 */
export const DEPLOYMENT_PROVIDERS = Object.freeze({
  ["1PANEL_CONSOLE"]: `${ACCESS_PROVIDERS["1PANEL"]}-console`,
  ["1PANEL_SITE"]: `${ACCESS_PROVIDERS["1PANEL"]}-site`,
  ALIYUN_ALB: `${ACCESS_PROVIDERS.ALIYUN}-alb`,
  ALIYUN_APIGW: `${ACCESS_PROVIDERS.ALIYUN}-apigw`,
  ALIYUN_CAS: `${ACCESS_PROVIDERS.ALIYUN}-cas`,
  ALIYUN_CAS_DEPLOY: `${ACCESS_PROVIDERS.ALIYUN}-casdeploy`,
  ALIYUN_CDN: `${ACCESS_PROVIDERS.ALIYUN}-cdn`,
  ALIYUN_CLB: `${ACCESS_PROVIDERS.ALIYUN}-clb`,
  ALIYUN_DCDN: `${ACCESS_PROVIDERS.ALIYUN}-dcdn`,
  ALIYUN_DDOSPRO: `${ACCESS_PROVIDERS.ALIYUN}-ddospro`,
  ALIYUN_ESA: `${ACCESS_PROVIDERS.ALIYUN}-esa`,
  ALIYUN_FC: `${ACCESS_PROVIDERS.ALIYUN}-fc`,
  ALIYUN_GA: `${ACCESS_PROVIDERS.ALIYUN}-ga`,
  ALIYUN_LIVE: `${ACCESS_PROVIDERS.ALIYUN}-live`,
  ALIYUN_NLB: `${ACCESS_PROVIDERS.ALIYUN}-nlb`,
  ALIYUN_OSS: `${ACCESS_PROVIDERS.ALIYUN}-oss`,
  ALIYUN_VOD: `${ACCESS_PROVIDERS.ALIYUN}-vod`,
  ALIYUN_WAF: `${ACCESS_PROVIDERS.ALIYUN}-waf`,
  APISIX: `${ACCESS_PROVIDERS.APISIX}`,
  AWS_ACM: `${ACCESS_PROVIDERS.AWS}-acm`,
  AWS_CLOUDFRONT: `${ACCESS_PROVIDERS.AWS}-cloudfront`,
  AWS_IAM: `${ACCESS_PROVIDERS.AWS}-iam`,
  AZURE_KEYVAULT: `${ACCESS_PROVIDERS.AZURE}-keyvault`,
  BAIDUCLOUD_APPBLB: `${ACCESS_PROVIDERS.BAIDUCLOUD}-appblb`,
  BAIDUCLOUD_BLB: `${ACCESS_PROVIDERS.BAIDUCLOUD}-blb`,
  BAIDUCLOUD_CDN: `${ACCESS_PROVIDERS.BAIDUCLOUD}-cdn`,
  BAIDUCLOUD_CERT: `${ACCESS_PROVIDERS.BAIDUCLOUD}-cert`,
  BAISHAN_CDN: `${ACCESS_PROVIDERS.BAISHAN}-cdn`,
  BAOTAPANEL_CONSOLE: `${ACCESS_PROVIDERS.BAOTAPANEL}-console`,
  BAOTAPANEL_SITE: `${ACCESS_PROVIDERS.BAOTAPANEL}-site`,
  BAOTAPANELGO_CONSOLE: `${ACCESS_PROVIDERS.BAOTAPANELGO}-console`,
  BAOTAPANELGO_SITE: `${ACCESS_PROVIDERS.BAOTAPANELGO}-site`,
  BAOTAWAF_CONSOLE: `${ACCESS_PROVIDERS.BAOTAWAF}-console`,
  BAOTAWAF_SITE: `${ACCESS_PROVIDERS.BAOTAWAF}-site`,
  BUNNY_CDN: `${ACCESS_PROVIDERS.BUNNY}-cdn`,
  BYTEPLUS_CDN: `${ACCESS_PROVIDERS.BYTEPLUS}-cdn`,
  CACHEFLY: `${ACCESS_PROVIDERS.CACHEFLY}`,
  CDNFLY: `${ACCESS_PROVIDERS.CDNFLY}`,
  CPANEL_SITE: `${ACCESS_PROVIDERS.CPANEL}-site`,
  CTCCCLOUD_AO: `${ACCESS_PROVIDERS.CTCCCLOUD}-ao`,
  CTCCCLOUD_CDN: `${ACCESS_PROVIDERS.CTCCCLOUD}-cdn`,
  CTCCCLOUD_CMS: `${ACCESS_PROVIDERS.CTCCCLOUD}-cms`,
  CTCCCLOUD_ELB: `${ACCESS_PROVIDERS.CTCCCLOUD}-elb`,
  CTCCCLOUD_ICDN: `${ACCESS_PROVIDERS.CTCCCLOUD}-icdn`,
  CTCCCLOUD_LVDN: `${ACCESS_PROVIDERS.CTCCCLOUD}-lvdn`,
  DOGECLOUD_CDN: `${ACCESS_PROVIDERS.DOGECLOUD}-cdn`,
  FLEXCDN: `${ACCESS_PROVIDERS.FLEXCDN}`,
  GCORE_CDN: `${ACCESS_PROVIDERS.GCORE}-cdn`,
  GOEDGE: `${ACCESS_PROVIDERS.GOEDGE}`,
  HUAWEICLOUD_CDN: `${ACCESS_PROVIDERS.HUAWEICLOUD}-cdn`,
  HUAWEICLOUD_ELB: `${ACCESS_PROVIDERS.HUAWEICLOUD}-elb`,
  HUAWEICLOUD_SCM: `${ACCESS_PROVIDERS.HUAWEICLOUD}-scm`,
  HUAWEICLOUD_OBS: `${ACCESS_PROVIDERS.HUAWEICLOUD}-obs`,
  HUAWEICLOUD_WAF: `${ACCESS_PROVIDERS.HUAWEICLOUD}-waf`,
  JDCLOUD_ALB: `${ACCESS_PROVIDERS.JDCLOUD}-alb`,
  JDCLOUD_CDN: `${ACCESS_PROVIDERS.JDCLOUD}-cdn`,
  JDCLOUD_LIVE: `${ACCESS_PROVIDERS.JDCLOUD}-live`,
  JDCLOUD_VOD: `${ACCESS_PROVIDERS.JDCLOUD}-vod`,
  KONG: `${ACCESS_PROVIDERS.KONG}`,
  KUBERNETES_SECRET: `${ACCESS_PROVIDERS.KUBERNETES}-secret`,
  KSYUN_CDN: `${ACCESS_PROVIDERS.KSYUN}-cdn`,
  LECDN: `${ACCESS_PROVIDERS.LECDN}`,
  LOCAL: `${ACCESS_PROVIDERS.LOCAL}`,
  MOHUA_MVH: `${ACCESS_PROVIDERS.MOHUA}-mvh`,
  NETLIFY_SITE: `${ACCESS_PROVIDERS.NETLIFY}-site`,
  PROXMOXVE: `${ACCESS_PROVIDERS.PROXMOXVE}`,
  QINIU_CDN: `${ACCESS_PROVIDERS.QINIU}-cdn`,
  QINIU_KODO: `${ACCESS_PROVIDERS.QINIU}-kodo`,
  QINIU_PILI: `${ACCESS_PROVIDERS.QINIU}-pili`,
  RAINYUN_RCDN: `${ACCESS_PROVIDERS.RAINYUN}-rcdn`,
  RATPANEL_CONSOLE: `${ACCESS_PROVIDERS.RATPANEL}-console`,
  RATPANEL_SITE: `${ACCESS_PROVIDERS.RATPANEL}-site`,
  SAFELINE_SITE: `${ACCESS_PROVIDERS.SAFELINE}-site`,
  SSH: `${ACCESS_PROVIDERS.SSH}`,
  TENCENTCLOUD_CDN: `${ACCESS_PROVIDERS.TENCENTCLOUD}-cdn`,
  TENCENTCLOUD_CLB: `${ACCESS_PROVIDERS.TENCENTCLOUD}-clb`,
  TENCENTCLOUD_COS: `${ACCESS_PROVIDERS.TENCENTCLOUD}-cos`,
  TENCENTCLOUD_CSS: `${ACCESS_PROVIDERS.TENCENTCLOUD}-css`,
  TENCENTCLOUD_ECDN: `${ACCESS_PROVIDERS.TENCENTCLOUD}-ecdn`,
  TENCENTCLOUD_EO: `${ACCESS_PROVIDERS.TENCENTCLOUD}-eo`,
  TENCENTCLOUD_GAAP: `${ACCESS_PROVIDERS.TENCENTCLOUD}-gaap`,
  TENCENTCLOUD_SCF: `${ACCESS_PROVIDERS.TENCENTCLOUD}-scf`,
  TENCENTCLOUD_SSL: `${ACCESS_PROVIDERS.TENCENTCLOUD}-ssl`,
  TENCENTCLOUD_SSL_DEPLOY: `${ACCESS_PROVIDERS.TENCENTCLOUD}-ssldeploy`,
  TENCENTCLOUD_SSL_UPDATE: `${ACCESS_PROVIDERS.TENCENTCLOUD}-sslupdate`,
  TENCENTCLOUD_VOD: `${ACCESS_PROVIDERS.TENCENTCLOUD}-vod`,
  TENCENTCLOUD_WAF: `${ACCESS_PROVIDERS.TENCENTCLOUD}-waf`,
  UCLOUD_UCDN: `${ACCESS_PROVIDERS.UCLOUD}-ucdn`,
  UCLOUD_US3: `${ACCESS_PROVIDERS.UCLOUD}-us3`,
  UNICLOUD_WEBHOST: `${ACCESS_PROVIDERS.UNICLOUD}-webhost`,
  UPYUN_CDN: `${ACCESS_PROVIDERS.UPYUN}-cdn`,
  UPYUN_FILE: `${ACCESS_PROVIDERS.UPYUN}-file`,
  VOLCENGINE_ALB: `${ACCESS_PROVIDERS.VOLCENGINE}-alb`,
  VOLCENGINE_CDN: `${ACCESS_PROVIDERS.VOLCENGINE}-cdn`,
  VOLCENGINE_CERTCENTER: `${ACCESS_PROVIDERS.VOLCENGINE}-certcenter`,
  VOLCENGINE_CLB: `${ACCESS_PROVIDERS.VOLCENGINE}-clb`,
  VOLCENGINE_DCDN: `${ACCESS_PROVIDERS.VOLCENGINE}-dcdn`,
  VOLCENGINE_IMAGEX: `${ACCESS_PROVIDERS.VOLCENGINE}-imagex`,
  VOLCENGINE_LIVE: `${ACCESS_PROVIDERS.VOLCENGINE}-live`,
  VOLCENGINE_TOS: `${ACCESS_PROVIDERS.VOLCENGINE}-tos`,
  WANGSU_CDN: `${ACCESS_PROVIDERS.WANGSU}-cdn`,
  WANGSU_CDNPRO: `${ACCESS_PROVIDERS.WANGSU}-cdnpro`,
  WANGSU_CERTIFICATE: `${ACCESS_PROVIDERS.WANGSU}-certificate`,
  WEBHOOK: `${ACCESS_PROVIDERS.WEBHOOK}`,
} as const);

export type DeploymentProviderType = (typeof DEPLOYMENT_PROVIDERS)[keyof typeof DEPLOYMENT_PROVIDERS];

export const DEPLOYMENT_CATEGORIES = Object.freeze({
  ALL: "all",
  CDN: "cdn",
  STORAGE: "storage",
  LOADBALANCE: "loadbalance",
  FIREWALL: "firewall",
  AV: "av",
  ACCELERATOR: "accelerator",
  APIGATEWAY: "apigw",
  SERVERLESS: "serverless",
  WEBSITE: "website",
  SSL: "ssl",
  OTHER: "other",
} as const);

export type DeploymentCategoryType = (typeof DEPLOYMENT_CATEGORIES)[keyof typeof DEPLOYMENT_CATEGORIES];

export interface DeploymentProvider extends BaseProviderWithAccess<DeploymentProviderType> {
  category: DeploymentCategoryType;
}

export const deploymentProvidersMap: Map<DeploymentProvider["type"] | string, DeploymentProvider> = new Map(
  /*
     注意：此处的顺序决定显示在前端的顺序。
     NOTICE: The following order determines the order displayed at the frontend.
    */
  (
    [
      [DEPLOYMENT_PROVIDERS.LOCAL, "provider.local", DEPLOYMENT_CATEGORIES.OTHER, "builtin"],
      [DEPLOYMENT_PROVIDERS.SSH, "provider.ssh", DEPLOYMENT_CATEGORIES.OTHER],
      [DEPLOYMENT_PROVIDERS.WEBHOOK, "provider.webhook", DEPLOYMENT_CATEGORIES.OTHER],
      [DEPLOYMENT_PROVIDERS.KUBERNETES_SECRET, "provider.kubernetes_secret", DEPLOYMENT_CATEGORIES.OTHER],
      [DEPLOYMENT_PROVIDERS.ALIYUN_OSS, "provider.aliyun_oss", DEPLOYMENT_CATEGORIES.STORAGE],
      [DEPLOYMENT_PROVIDERS.ALIYUN_CDN, "provider.aliyun_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.ALIYUN_DCDN, "provider.aliyun_dcdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.ALIYUN_ESA, "provider.aliyun_esa", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.ALIYUN_CLB, "provider.aliyun_clb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
      [DEPLOYMENT_PROVIDERS.ALIYUN_ALB, "provider.aliyun_alb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
      [DEPLOYMENT_PROVIDERS.ALIYUN_NLB, "provider.aliyun_nlb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
      [DEPLOYMENT_PROVIDERS.ALIYUN_WAF, "provider.aliyun_waf", DEPLOYMENT_CATEGORIES.FIREWALL],
      [DEPLOYMENT_PROVIDERS.ALIYUN_DDOSPRO, "provider.aliyun_ddospro", DEPLOYMENT_CATEGORIES.FIREWALL],
      [DEPLOYMENT_PROVIDERS.ALIYUN_LIVE, "provider.aliyun_live", DEPLOYMENT_CATEGORIES.AV],
      [DEPLOYMENT_PROVIDERS.ALIYUN_VOD, "provider.aliyun_vod", DEPLOYMENT_CATEGORIES.AV],
      [DEPLOYMENT_PROVIDERS.ALIYUN_FC, "provider.aliyun_fc", DEPLOYMENT_CATEGORIES.SERVERLESS],
      [DEPLOYMENT_PROVIDERS.ALIYUN_APIGW, "provider.aliyun_apigw", DEPLOYMENT_CATEGORIES.APIGATEWAY],
      [DEPLOYMENT_PROVIDERS.ALIYUN_GA, "provider.aliyun_ga", DEPLOYMENT_CATEGORIES.ACCELERATOR],
      [DEPLOYMENT_PROVIDERS.ALIYUN_CAS, "provider.aliyun_casupload", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.ALIYUN_CAS_DEPLOY, "provider.aliyun_casdeploy", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_COS, "provider.tencentcloud_cos", DEPLOYMENT_CATEGORIES.STORAGE],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CDN, "provider.tencentcloud_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_ECDN, "provider.tencentcloud_ecdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_EO, "provider.tencentcloud_eo", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CLB, "provider.tencentcloud_clb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_WAF, "provider.tencentcloud_waf", DEPLOYMENT_CATEGORIES.FIREWALL],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CSS, "provider.tencentcloud_css", DEPLOYMENT_CATEGORIES.AV],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_VOD, "provider.tencentcloud_vod", DEPLOYMENT_CATEGORIES.AV],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SCF, "provider.tencentcloud_scf", DEPLOYMENT_CATEGORIES.SERVERLESS],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_GAAP, "provider.tencentcloud_gaap", DEPLOYMENT_CATEGORIES.ACCELERATOR],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL, "provider.tencentcloud_sslupload", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL_DEPLOY, "provider.tencentcloud_ssldeploy", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL_UPDATE, "provider.tencentcloud_sslupdate", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.BAIDUCLOUD_CDN, "provider.baiducloud_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.BAIDUCLOUD_BLB, "provider.baiducloud_blb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
      [DEPLOYMENT_PROVIDERS.BAIDUCLOUD_APPBLB, "provider.baiducloud_appblb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
      [DEPLOYMENT_PROVIDERS.BAIDUCLOUD_CERT, "provider.baiducloud_certupload", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_CDN, "provider.huaweicloud_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_OBS, "provider.huaweicloud_obs", DEPLOYMENT_CATEGORIES.STORAGE],
      [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_ELB, "provider.huaweicloud_elb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
      [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_WAF, "provider.huaweicloud_waf", DEPLOYMENT_CATEGORIES.FIREWALL],
      [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_SCM, "provider.huaweicloud_scmupload", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.VOLCENGINE_TOS, "provider.volcengine_tos", DEPLOYMENT_CATEGORIES.STORAGE],
      [DEPLOYMENT_PROVIDERS.VOLCENGINE_CDN, "provider.volcengine_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.VOLCENGINE_DCDN, "provider.volcengine_dcdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.VOLCENGINE_CLB, "provider.volcengine_clb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
      [DEPLOYMENT_PROVIDERS.VOLCENGINE_ALB, "provider.volcengine_alb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
      [DEPLOYMENT_PROVIDERS.VOLCENGINE_IMAGEX, "provider.volcengine_imagex", DEPLOYMENT_CATEGORIES.STORAGE],
      [DEPLOYMENT_PROVIDERS.VOLCENGINE_LIVE, "provider.volcengine_live", DEPLOYMENT_CATEGORIES.AV],
      [DEPLOYMENT_PROVIDERS.VOLCENGINE_CERTCENTER, "provider.volcengine_certcenterupload", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.JDCLOUD_ALB, "provider.jdcloud_alb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
      [DEPLOYMENT_PROVIDERS.JDCLOUD_CDN, "provider.jdcloud_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.JDCLOUD_LIVE, "provider.jdcloud_live", DEPLOYMENT_CATEGORIES.AV],
      [DEPLOYMENT_PROVIDERS.JDCLOUD_VOD, "provider.jdcloud_vod", DEPLOYMENT_CATEGORIES.AV],
      [DEPLOYMENT_PROVIDERS.QINIU_KODO, "provider.qiniu_kodo", DEPLOYMENT_CATEGORIES.STORAGE],
      [DEPLOYMENT_PROVIDERS.QINIU_CDN, "provider.qiniu_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.QINIU_PILI, "provider.qiniu_pili", DEPLOYMENT_CATEGORIES.AV],
      [DEPLOYMENT_PROVIDERS.UPYUN_FILE, "provider.upyun_file", DEPLOYMENT_CATEGORIES.STORAGE],
      [DEPLOYMENT_PROVIDERS.UPYUN_CDN, "provider.upyun_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.BAISHAN_CDN, "provider.baishan_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.WANGSU_CDN, "provider.wangsu_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.WANGSU_CDNPRO, "provider.wangsu_cdnpro", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.WANGSU_CERTIFICATE, "provider.wangsu_certificateupload", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.DOGECLOUD_CDN, "provider.dogecloud_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.KSYUN_CDN, "provider.ksyun_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.BYTEPLUS_CDN, "provider.byteplus_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.UCLOUD_US3, "provider.ucloud_us3", DEPLOYMENT_CATEGORIES.STORAGE],
      [DEPLOYMENT_PROVIDERS.UCLOUD_UCDN, "provider.ucloud_ucdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.CTCCCLOUD_CDN, "provider.ctcccloud_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.CTCCCLOUD_ICDN, "provider.ctcccloud_icdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.CTCCCLOUD_AO, "provider.ctcccloud_ao", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.CTCCCLOUD_ELB, "provider.ctcccloud_elb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
      [DEPLOYMENT_PROVIDERS.CTCCCLOUD_LVDN, "provider.ctcccloud_lvdn", DEPLOYMENT_CATEGORIES.AV],
      [DEPLOYMENT_PROVIDERS.CTCCCLOUD_CMS, "provider.ctcccloud_cmsupload", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.RAINYUN_RCDN, "provider.rainyun_rcdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.UNICLOUD_WEBHOST, "provider.unicloud_webhost", DEPLOYMENT_CATEGORIES.WEBSITE],
      [DEPLOYMENT_PROVIDERS.AWS_CLOUDFRONT, "provider.aws_cloudfront", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.AWS_ACM, "provider.aws_acm", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.AWS_IAM, "provider.aws_iam", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.AZURE_KEYVAULT, "provider.azure_keyvault", DEPLOYMENT_CATEGORIES.SSL],
      [DEPLOYMENT_PROVIDERS.BUNNY_CDN, "provider.bunny_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.CACHEFLY, "provider.cachefly", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.GCORE_CDN, "provider.gcore_cdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.MOHUA_MVH, "provider.mohua_mvh", DEPLOYMENT_CATEGORIES.WEBSITE],
      [DEPLOYMENT_PROVIDERS.NETLIFY_SITE, "provider.netlify_site", DEPLOYMENT_CATEGORIES.WEBSITE],
      [DEPLOYMENT_PROVIDERS.CDNFLY, "provider.cdnfly", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.FLEXCDN, "provider.flexcdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.GOEDGE, "provider.goedge", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS.LECDN, "provider.lecdn", DEPLOYMENT_CATEGORIES.CDN],
      [DEPLOYMENT_PROVIDERS["1PANEL_SITE"], "provider.1panel_site", DEPLOYMENT_CATEGORIES.WEBSITE],
      [DEPLOYMENT_PROVIDERS["1PANEL_CONSOLE"], "provider.1panel_console", DEPLOYMENT_CATEGORIES.OTHER],
      [DEPLOYMENT_PROVIDERS.BAOTAPANEL_SITE, "provider.baotapanel_site", DEPLOYMENT_CATEGORIES.WEBSITE],
      [DEPLOYMENT_PROVIDERS.BAOTAPANEL_CONSOLE, "provider.baotapanel_console", DEPLOYMENT_CATEGORIES.OTHER],
      [DEPLOYMENT_PROVIDERS.BAOTAPANELGO_SITE, "provider.baotapanelgo_site", DEPLOYMENT_CATEGORIES.WEBSITE],
      [DEPLOYMENT_PROVIDERS.BAOTAPANELGO_CONSOLE, "provider.baotapanelgo_console", DEPLOYMENT_CATEGORIES.OTHER],
      [DEPLOYMENT_PROVIDERS.RATPANEL_SITE, "provider.ratpanel_site", DEPLOYMENT_CATEGORIES.WEBSITE],
      [DEPLOYMENT_PROVIDERS.RATPANEL_CONSOLE, "provider.ratpanel_console", DEPLOYMENT_CATEGORIES.OTHER],
      [DEPLOYMENT_PROVIDERS.BAOTAWAF_SITE, "provider.baotawaf_site", DEPLOYMENT_CATEGORIES.FIREWALL],
      [DEPLOYMENT_PROVIDERS.BAOTAWAF_CONSOLE, "provider.baotawaf_console", DEPLOYMENT_CATEGORIES.OTHER],
      [DEPLOYMENT_PROVIDERS.SAFELINE_SITE, "provider.safeline_site", DEPLOYMENT_CATEGORIES.FIREWALL],
      [DEPLOYMENT_PROVIDERS.APISIX, "provider.apisix", DEPLOYMENT_CATEGORIES.APIGATEWAY],
      [DEPLOYMENT_PROVIDERS.KONG, "provider.kong", DEPLOYMENT_CATEGORIES.APIGATEWAY],
      [DEPLOYMENT_PROVIDERS.CPANEL_SITE, "provider.cpanel_site", DEPLOYMENT_CATEGORIES.WEBSITE],
      [DEPLOYMENT_PROVIDERS.PROXMOXVE, "provider.proxmoxve", DEPLOYMENT_CATEGORIES.OTHER],
    ] satisfies Array<[DeploymentProviderType, string, DeploymentCategoryType, "builtin"] | [DeploymentProviderType, string, DeploymentCategoryType]>
  ).map(([type, name, category, builtin]) => [
    type,
    {
      type: type,
      name: name,
      icon: accessProvidersMap.get(type.split("-")[0])!.icon,
      provider: type.split("-")[0] as AccessProviderType,
      category: category,
      builtin: builtin === "builtin",
    },
  ])
);
// #endregion

// #region NotificationProvider
/*
  注意：如果追加新的常量值，请保持以 ASCII 排序。
  NOTICE: If you add new constant, please keep ASCII order.
 */
export const NOTIFICATION_PROVIDERS = Object.freeze({
  DINGTALKBOT: `${ACCESS_PROVIDERS.DINGTALKBOT}`,
  DISCORDBOT: `${ACCESS_PROVIDERS.DISCORDBOT}`,
  EMAIL: `${ACCESS_PROVIDERS.EMAIL}`,
  LARKBOT: `${ACCESS_PROVIDERS.LARKBOT}`,
  MATTERMOST: `${ACCESS_PROVIDERS.MATTERMOST}`,
  SLACKBOT: `${ACCESS_PROVIDERS.SLACKBOT}`,
  TELEGRAMBOT: `${ACCESS_PROVIDERS.TELEGRAMBOT}`,
  WEBHOOK: `${ACCESS_PROVIDERS.WEBHOOK}`,
  WECOMBOT: `${ACCESS_PROVIDERS.WECOMBOT}`,
} as const);

export type NotificationProviderType = (typeof NOTIFICATION_PROVIDERS)[keyof typeof NOTIFICATION_PROVIDERS];

export interface NotificationProvider extends BaseProviderWithAccess<NotificationProviderType> {}

export const notificationProvidersMap: Map<NotificationProvider["type"] | string, NotificationProvider> = new Map(
  /*
    注意：此处的顺序决定显示在前端的顺序。
    NOTICE: The following order determines the order displayed at the frontend.
   */
  (
    [
      [NOTIFICATION_PROVIDERS.EMAIL],
      [NOTIFICATION_PROVIDERS.WEBHOOK],
      [NOTIFICATION_PROVIDERS.DINGTALKBOT],
      [NOTIFICATION_PROVIDERS.LARKBOT],
      [NOTIFICATION_PROVIDERS.WECOMBOT],
      [NOTIFICATION_PROVIDERS.DISCORDBOT],
      [NOTIFICATION_PROVIDERS.SLACKBOT],
      [NOTIFICATION_PROVIDERS.TELEGRAMBOT],
      [NOTIFICATION_PROVIDERS.MATTERMOST],
    ] satisfies Array<[NotificationProviderType]>
  ).map(([type]) => [
    type,
    {
      type: type,
      name: accessProvidersMap.get(type.split("-")[0])!.name,
      icon: accessProvidersMap.get(type.split("-")[0])!.icon,
      provider: type.split("-")[0] as AccessProviderType,
      builtin: false,
    },
  ])
);
// #endregion
