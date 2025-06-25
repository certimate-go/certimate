// #region AccessProvider
/*
  注意：如果追加新的常量值，请保持以 ASCII 排序。
  NOTICE: If you add new constant, please keep ASCII order.
 */
export const ACCESS_PROVIDERS = Object.freeze({
  ["1PANEL"]: "1panel",
  ACMECA: "acmeca",
  ACMEHTTPREQ: "acmehttpreq",
  ALIYUN: "aliyun",
  APISIX: "apisix",
  AWS: "aws",
  AZURE: "azure",
  BAIDUCLOUD: "baiducloud",
  BAISHAN: "baishan",
  BAOTAPANEL: "baotapanel",
  BAOTAWAF: "baotawaf",
  BUNNY: "bunny",
  BYTEPLUS: "byteplus",
  BUYPASS: "buypass",
  CACHEFLY: "cachefly",
  CDNFLY: "cdnfly",
  CLOUDFLARE: "cloudflare",
  CLOUDNS: "cloudns",
  CMCCCLOUD: "cmcccloud",
  CONSTELLIX: "constellix",
  CTCCCLOUD: "ctcccloud",
  DESEC: "desec",
  DIGITALOCEAN: "digitalocean",
  DINGTALKBOT: "dingtalkbot",
  DISCORDBOT: "discordbot",
  DNSLA: "dnsla",
  DOGECLOUD: "dogecloud",
  DUCKDNS: "duckdns",
  DYNV6: "dynv6",
  EDGIO: "edgio",
  EMAIL: "email",
  FLEXCDN: "flexcdn",
  GCORE: "gcore",
  GNAME: "gname",
  GODADDY: "godaddy",
  GOEDGE: "goedge",
  GOOGLETRUSTSERVICES: "googletrustservices",
  HETZNER: "hetzner",
  HUAWEICLOUD: "huaweicloud",
  JDCLOUD: "jdcloud",
  KONG: "kong",
  KUBERNETES: "k8s",
  LARKBOT: "larkbot",
  LECDN: "lecdn",
  LETSENCRYPT: "letsencrypt",
  LETSENCRYPTSTAGING: "letsencryptstaging",
  LOCAL: "local",
  MATTERMOST: "mattermost",
  NAMECHEAP: "namecheap",
  NAMEDOTCOM: "namedotcom",
  NAMESILO: "namesilo",
  NETCUP: "netcup",
  NETLIFY: "netlify",
  NS1: "ns1",
  PORKBUN: "porkbun",
  POWERDNS: "powerdns",
  PROXMOXVE: "proxmoxve",
  QINIU: "qiniu",
  RAINYUN: "rainyun",
  RATPANEL: "ratpanel",
  SAFELINE: "safeline",
  SLACKBOT: "slackbot",
  SPACESHIP: "spaceship",
  SSH: "ssh",
  SSLCOM: "sslcom",
  TELEGRAMBOT: "telegrambot",
  TENCENTCLOUD: "tencentcloud",
  UCLOUD: "ucloud",
  UNICLOUD: "unicloud",
  UPYUN: "upyun",
  VERCEL: "vercel",
  VOLCENGINE: "volcengine",
  WANGSU: "wangsu",
  WEBHOOK: "webhook",
  WECOMBOT: "wecombot",
  WESTCN: "westcn",
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

export type AccessProvider = {
  type: AccessProviderType;
  name: string;
  icon: string;
  usages: AccessUsageType[];
  builtin: boolean;
};

export const accessProvidersMap: Map<AccessProvider["type"] | string, AccessProvider> = new Map(
  /*
    注意：此处的顺序决定显示在前端的顺序。
    NOTICE: The following order determines the order displayed at the frontend.
  */
  [
    [ACCESS_PROVIDERS.LOCAL, "provider.local", "/imgs/providers/local.svg", [ACCESS_USAGES.HOSTING]],
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

    [ACCESS_PROVIDERS.QINIU, "provider.qiniu", "/imgs/providers/qiniu.svg", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.UPYUN, "provider.upyun", "/imgs/providers/upyun.svg", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.BAISHAN, "provider.baishan", "/imgs/providers/baishan.png", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.WANGSU, "provider.wangsu", "/imgs/providers/wangsu.svg", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.DOGECLOUD, "provider.dogecloud", "/imgs/providers/dogecloud.png", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.BYTEPLUS, "provider.byteplus", "/imgs/providers/byteplus.svg", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.UNICLOUD, "provider.unicloud", "/imgs/providers/unicloud.png", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS["1PANEL"], "provider.1panel", "/imgs/providers/1panel.svg", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.BAOTAPANEL, "provider.baotapanel", "/imgs/providers/baotapanel.svg", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.BAOTAWAF, "provider.baotawaf", "/imgs/providers/baotawaf.svg", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.RATPANEL, "provider.ratpanel", "/imgs/providers/ratpanel.png", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.SAFELINE, "provider.safeline", "/imgs/providers/safeline.svg", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.CDNFLY, "provider.cdnfly", "/imgs/providers/cdnfly.png", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.FLEXCDN, "provider.flexcdn", "/imgs/providers/flexcdn.png", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.GOEDGE, "provider.goedge", "/imgs/providers/goedge.png", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.LECDN, "provider.lecdn", "/imgs/providers/lecdn.svg", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.CACHEFLY, "provider.cachefly", "/imgs/providers/cachefly.png", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.EDGIO, "provider.edgio", "/imgs/providers/edgio.svg", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.APISIX, "provider.apisix", "/imgs/providers/apisix.svg", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.KONG, "provider.kong", "/imgs/providers/kong.png", [ACCESS_USAGES.HOSTING]],
    [ACCESS_PROVIDERS.PROXMOXVE, "provider.proxmoxve", "/imgs/providers/proxmoxve.svg", [ACCESS_USAGES.HOSTING]],

    [ACCESS_PROVIDERS.CLOUDFLARE, "provider.cloudflare", "/imgs/providers/cloudflare.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.CLOUDNS, "provider.cloudns", "/imgs/providers/cloudns.png", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.CONSTELLIX, "provider.constellix", "/imgs/providers/constellix.png", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.DESEC, "provider.desec", "/imgs/providers/desec.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.DIGITALOCEAN, "provider.digitalocean", "/imgs/providers/digitalocean.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.DNSLA, "provider.dnsla", "/imgs/providers/dnsla.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.DUCKDNS, "provider.duckdns", "/imgs/providers/duckdns.png", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.DYNV6, "provider.dynv6", "/imgs/providers/dynv6.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.GNAME, "provider.gname", "/imgs/providers/gname.png", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.GODADDY, "provider.godaddy", "/imgs/providers/godaddy.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.HETZNER, "provider.hetzner", "/imgs/providers/hetzner.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.NAMECHEAP, "provider.namecheap", "/imgs/providers/namecheap.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.NAMEDOTCOM, "provider.namedotcom", "/imgs/providers/namedotcom.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.NAMESILO, "provider.namesilo", "/imgs/providers/namesilo.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.NETCUP, "provider.netcup", "/imgs/providers/netcup.png", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.NS1, "provider.ns1", "/imgs/providers/ns1.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.PORKBUN, "provider.porkbun", "/imgs/providers/porkbun.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.SPACESHIP, "provider.spaceship", "/imgs/providers/spaceship.png", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.VERCEL, "provider.vercel", "/imgs/providers/vercel.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.CMCCCLOUD, "provider.cmcccloud", "/imgs/providers/cmcccloud.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.WESTCN, "provider.westcn", "/imgs/providers/westcn.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.POWERDNS, "provider.powerdns", "/imgs/providers/powerdns.svg", [ACCESS_USAGES.DNS]],
    [ACCESS_PROVIDERS.ACMEHTTPREQ, "provider.acmehttpreq", "/imgs/providers/acmehttpreq.svg", [ACCESS_USAGES.DNS]],

    [ACCESS_PROVIDERS.LETSENCRYPT, "provider.letsencrypt", "/imgs/providers/letsencrypt.svg", [ACCESS_USAGES.CA]],
    [ACCESS_PROVIDERS.LETSENCRYPTSTAGING, "provider.letsencryptstaging", "/imgs/providers/letsencrypt.svg", [ACCESS_USAGES.CA]],
    [ACCESS_PROVIDERS.BUYPASS, "provider.buypass", "/imgs/providers/buypass.png", [ACCESS_USAGES.CA]],
    [ACCESS_PROVIDERS.GOOGLETRUSTSERVICES, "provider.googletrustservices", "/imgs/providers/google.svg", [ACCESS_USAGES.CA]],
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
  ].map((e) => [
    e[0] as string,
    {
      type: e[0] as AccessProviderType,
      name: e[1] as string,
      icon: e[2] as string,
      usages: e[3] as AccessUsageType[],
      builtin: ([ACCESS_PROVIDERS.LOCAL, ACCESS_PROVIDERS.LETSENCRYPT, ACCESS_PROVIDERS.LETSENCRYPTSTAGING] as string[]).includes(e[0] as string),
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
  BUYPASS: `${ACCESS_PROVIDERS.BUYPASS}`,
  GOOGLETRUSTSERVICES: `${ACCESS_PROVIDERS.GOOGLETRUSTSERVICES}`,
  LETSENCRYPT: `${ACCESS_PROVIDERS.LETSENCRYPT}`,
  LETSENCRYPTSTAGING: `${ACCESS_PROVIDERS.LETSENCRYPTSTAGING}`,
  SSLCOM: `${ACCESS_PROVIDERS.SSLCOM}`,
  ZEROSSL: `${ACCESS_PROVIDERS.ZEROSSL}`,
} as const);

export type CAProviderType = (typeof CA_PROVIDERS)[keyof typeof CA_PROVIDERS];

export type CAProvider = {
  type: CAProviderType;
  name: string;
  icon: string;
  provider: AccessProviderType;
  builtin: boolean;
};

export const caProvidersMap: Map<CAProvider["type"] | string, CAProvider> = new Map(
  /*
    注意：此处的顺序决定显示在前端的顺序。
    NOTICE: The following order determines the order displayed at the frontend.
  */
  [
    [CA_PROVIDERS.LETSENCRYPT, "builtin"],
    [CA_PROVIDERS.LETSENCRYPTSTAGING, "builtin"],
    [CA_PROVIDERS.BUYPASS],
    [CA_PROVIDERS.GOOGLETRUSTSERVICES],
    [CA_PROVIDERS.SSLCOM],
    [CA_PROVIDERS.ZEROSSL],
    [CA_PROVIDERS.ACMECA],
  ].map(([type, builtin]) => [
    type,
    {
      type: type as CAProviderType,
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
  ACMEHTTPREQ: `${ACCESS_PROVIDERS.ACMEHTTPREQ}`,
  ALIYUN: `${ACCESS_PROVIDERS.ALIYUN}`, // 兼容旧值，等同于 `ALIYUN_DNS`
  ALIYUN_DNS: `${ACCESS_PROVIDERS.ALIYUN}-dns`,
  ALIYUN_ESA: `${ACCESS_PROVIDERS.ALIYUN}-esa`,
  AWS: `${ACCESS_PROVIDERS.AWS}`, // 兼容旧值，等同于 `AWS_ROUTE53`
  AWS_ROUTE53: `${ACCESS_PROVIDERS.AWS}-route53`,
  AZURE: `${ACCESS_PROVIDERS.AZURE}`, // 兼容旧值，等同于 `AZURE_DNS`
  AZURE_DNS: `${ACCESS_PROVIDERS.AZURE}-dns`,
  BAIDUCLOUD: `${ACCESS_PROVIDERS.BAIDUCLOUD}`, // 兼容旧值，等同于 `BAIDUCLOUD_DNS`
  BAIDUCLOUD_DNS: `${ACCESS_PROVIDERS.BAIDUCLOUD}-dns`,
  BUNNY: `${ACCESS_PROVIDERS.BUNNY}`,
  CLOUDFLARE: `${ACCESS_PROVIDERS.CLOUDFLARE}`,
  CLOUDNS: `${ACCESS_PROVIDERS.CLOUDNS}`,
  CMCCCLOUD: `${ACCESS_PROVIDERS.CMCCCLOUD}`, // 兼容旧值，等同于 `CMCCCLOUD_DNS`
  CMCCCLOUD_DNS: `${ACCESS_PROVIDERS.CMCCCLOUD}-dns`,
  CONSTELLIX: `${ACCESS_PROVIDERS.CONSTELLIX}`,
  CTCCCLOUD: `${ACCESS_PROVIDERS.CTCCCLOUD}`, // 兼容旧值，等同于 `CTCCCLOUD_SMARTDNS`
  CTCCCLOUD_SMARTDNS: `${ACCESS_PROVIDERS.CTCCCLOUD}-smartdns`,
  DESEC: `${ACCESS_PROVIDERS.DESEC}`,
  DIGITALOCEAN: `${ACCESS_PROVIDERS.DIGITALOCEAN}`,
  DNSLA: `${ACCESS_PROVIDERS.DNSLA}`,
  DUCKDNS: `${ACCESS_PROVIDERS.DUCKDNS}`,
  DYNV6: `${ACCESS_PROVIDERS.DYNV6}`,
  GCORE: `${ACCESS_PROVIDERS.GCORE}`,
  GNAME: `${ACCESS_PROVIDERS.GNAME}`,
  GODADDY: `${ACCESS_PROVIDERS.GODADDY}`,
  HETZNER: `${ACCESS_PROVIDERS.HETZNER}`,
  HUAWEICLOUD: `${ACCESS_PROVIDERS.HUAWEICLOUD}`, // 兼容旧值，等同于 `HUAWEICLOUD_DNS`
  HUAWEICLOUD_DNS: `${ACCESS_PROVIDERS.HUAWEICLOUD}-dns`,
  JDCLOUD: `${ACCESS_PROVIDERS.JDCLOUD}`, // 兼容旧值，等同于 `JDCLOUD_DNS`
  JDCLOUD_DNS: `${ACCESS_PROVIDERS.JDCLOUD}-dns`,
  NAMECHEAP: `${ACCESS_PROVIDERS.NAMECHEAP}`,
  NAMEDOTCOM: `${ACCESS_PROVIDERS.NAMEDOTCOM}`,
  NAMESILO: `${ACCESS_PROVIDERS.NAMESILO}`,
  NETCUP: `${ACCESS_PROVIDERS.NETCUP}`,
  NETLIFY: `${ACCESS_PROVIDERS.NETLIFY}`,
  NS1: `${ACCESS_PROVIDERS.NS1}`,
  PORKBUN: `${ACCESS_PROVIDERS.PORKBUN}`,
  POWERDNS: `${ACCESS_PROVIDERS.POWERDNS}`,
  RAINYUN: `${ACCESS_PROVIDERS.RAINYUN}`,
  SPACESHIP: `${ACCESS_PROVIDERS.SPACESHIP}`,
  UCLOUD_UDNR: `${ACCESS_PROVIDERS.UCLOUD}-udnr`,
  TENCENTCLOUD: `${ACCESS_PROVIDERS.TENCENTCLOUD}`, // 兼容旧值，等同于 `TENCENTCLOUD_DNS`
  TENCENTCLOUD_DNS: `${ACCESS_PROVIDERS.TENCENTCLOUD}-dns`,
  TENCENTCLOUD_EO: `${ACCESS_PROVIDERS.TENCENTCLOUD}-eo`,
  VERCEL: `${ACCESS_PROVIDERS.VERCEL}`,
  VOLCENGINE: `${ACCESS_PROVIDERS.VOLCENGINE}`, // 兼容旧值，等同于 `VOLCENGINE_DNS`
  VOLCENGINE_DNS: `${ACCESS_PROVIDERS.VOLCENGINE}-dns`,
  WESTCN: `${ACCESS_PROVIDERS.WESTCN}`,
} as const);

export type ACMEDns01ProviderType = (typeof ACME_DNS01_PROVIDERS)[keyof typeof ACME_DNS01_PROVIDERS];

export type ACMEDns01Provider = {
  type: ACMEDns01ProviderType;
  name: string;
  icon: string;
  provider: AccessProviderType;
};

export const acmeDns01ProvidersMap: Map<ACMEDns01Provider["type"] | string, ACMEDns01Provider> = new Map(
  /*
    注意：此处的顺序决定显示在前端的顺序。
    NOTICE: The following order determines the order displayed at the frontend.
   */
  [
    [ACME_DNS01_PROVIDERS.ALIYUN_DNS, "provider.aliyun.dns"],
    [ACME_DNS01_PROVIDERS.ALIYUN_ESA, "provider.aliyun.esa"],
    [ACME_DNS01_PROVIDERS.TENCENTCLOUD_DNS, "provider.tencentcloud.dns"],
    [ACME_DNS01_PROVIDERS.TENCENTCLOUD_EO, "provider.tencentcloud.eo"],
    [ACME_DNS01_PROVIDERS.BAIDUCLOUD_DNS, "provider.baiducloud.dns"],
    [ACME_DNS01_PROVIDERS.HUAWEICLOUD_DNS, "provider.huaweicloud.dns"],
    [ACME_DNS01_PROVIDERS.VOLCENGINE_DNS, "provider.volcengine.dns"],
    [ACME_DNS01_PROVIDERS.JDCLOUD_DNS, "provider.jdcloud.dns"],
    [ACME_DNS01_PROVIDERS.AWS_ROUTE53, "provider.aws.route53"],
    [ACME_DNS01_PROVIDERS.AZURE_DNS, "provider.azure.dns"],
    [ACME_DNS01_PROVIDERS.BUNNY, "provider.bunny"],
    [ACME_DNS01_PROVIDERS.CLOUDFLARE, "provider.cloudflare"],
    [ACME_DNS01_PROVIDERS.CLOUDNS, "provider.cloudns"],
    [ACME_DNS01_PROVIDERS.CONSTELLIX, "provider.constellix"],
    [ACME_DNS01_PROVIDERS.DESEC, "provider.desec"],
    [ACME_DNS01_PROVIDERS.DIGITALOCEAN, "provider.digitalocean"],
    [ACME_DNS01_PROVIDERS.DNSLA, "provider.dnsla"],
    [ACME_DNS01_PROVIDERS.DUCKDNS, "provider.duckdns"],
    [ACME_DNS01_PROVIDERS.DYNV6, "provider.dynv6"],
    [ACME_DNS01_PROVIDERS.GCORE, "provider.gcore"],
    [ACME_DNS01_PROVIDERS.GNAME, "provider.gname"],
    [ACME_DNS01_PROVIDERS.GODADDY, "provider.godaddy"],
    [ACME_DNS01_PROVIDERS.HETZNER, "provider.hetzner"],
    [ACME_DNS01_PROVIDERS.NAMECHEAP, "provider.namecheap"],
    [ACME_DNS01_PROVIDERS.NAMEDOTCOM, "provider.namedotcom"],
    [ACME_DNS01_PROVIDERS.NAMESILO, "provider.namesilo"],
    [ACME_DNS01_PROVIDERS.NETCUP, "provider.netcup"],
    [ACME_DNS01_PROVIDERS.NETLIFY, "provider.netlify"],
    [ACME_DNS01_PROVIDERS.NS1, "provider.ns1"],
    [ACME_DNS01_PROVIDERS.PORKBUN, "provider.porkbun"],
    [ACME_DNS01_PROVIDERS.SPACESHIP, "provider.spaceship"],
    [ACME_DNS01_PROVIDERS.VERCEL, "provider.vercel"],
    [ACME_DNS01_PROVIDERS.CMCCCLOUD_DNS, "provider.cmcccloud.dns"],
    [ACME_DNS01_PROVIDERS.CTCCCLOUD_SMARTDNS, "provider.ctcccloud.smartdns"],
    [ACME_DNS01_PROVIDERS.RAINYUN, "provider.rainyun"],
    [ACME_DNS01_PROVIDERS.UCLOUD_UDNR, "provider.ucloud.udnr"],
    [ACME_DNS01_PROVIDERS.WESTCN, "provider.westcn"],
    [ACME_DNS01_PROVIDERS.POWERDNS, "provider.powerdns"],
    [ACME_DNS01_PROVIDERS.ACMEHTTPREQ, "provider.acmehttpreq"],
  ].map(([type, name]) => [
    type,
    {
      type: type as ACMEDns01ProviderType,
      name: name,
      icon: accessProvidersMap.get(type.split("-")[0])!.icon,
      provider: type.split("-")[0] as AccessProviderType,
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
  ALIYUN_DDOS: `${ACCESS_PROVIDERS.ALIYUN}-ddospro`,
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
  BAOTAWAF_CONSOLE: `${ACCESS_PROVIDERS.BAOTAWAF}-console`,
  BAOTAWAF_SITE: `${ACCESS_PROVIDERS.BAOTAWAF}-site`,
  BUNNY_CDN: `${ACCESS_PROVIDERS.BUNNY}-cdn`,
  BYTEPLUS_CDN: `${ACCESS_PROVIDERS.BYTEPLUS}-cdn`,
  CACHEFLY: `${ACCESS_PROVIDERS.CACHEFLY}`,
  CDNFLY: `${ACCESS_PROVIDERS.CDNFLY}`,
  CTCCCLOUD_AO: `${ACCESS_PROVIDERS.CTCCCLOUD}-ao`,
  CTCCCLOUD_CDN: `${ACCESS_PROVIDERS.CTCCCLOUD}-cdn`,
  CTCCCLOUD_CMS: `${ACCESS_PROVIDERS.CTCCCLOUD}-cms`,
  CTCCCLOUD_ELB: `${ACCESS_PROVIDERS.CTCCCLOUD}-elb`,
  CTCCCLOUD_ICDN: `${ACCESS_PROVIDERS.CTCCCLOUD}-icdn`,
  CTCCCLOUD_LVDN: `${ACCESS_PROVIDERS.CTCCCLOUD}-lvdn`,
  DOGECLOUD_CDN: `${ACCESS_PROVIDERS.DOGECLOUD}-cdn`,
  EDGIO_APPLICATIONS: `${ACCESS_PROVIDERS.EDGIO}-applications`,
  FLEXCDN: `${ACCESS_PROVIDERS.FLEXCDN}`,
  GCORE_CDN: `${ACCESS_PROVIDERS.GCORE}-cdn`,
  GOEDGE: `${ACCESS_PROVIDERS.GOEDGE}`,
  HUAWEICLOUD_CDN: `${ACCESS_PROVIDERS.HUAWEICLOUD}-cdn`,
  HUAWEICLOUD_ELB: `${ACCESS_PROVIDERS.HUAWEICLOUD}-elb`,
  HUAWEICLOUD_SCM: `${ACCESS_PROVIDERS.HUAWEICLOUD}-scm`,
  HUAWEICLOUD_WAF: `${ACCESS_PROVIDERS.HUAWEICLOUD}-waf`,
  JDCLOUD_ALB: `${ACCESS_PROVIDERS.JDCLOUD}-alb`,
  JDCLOUD_CDN: `${ACCESS_PROVIDERS.JDCLOUD}-cdn`,
  JDCLOUD_LIVE: `${ACCESS_PROVIDERS.JDCLOUD}-live`,
  JDCLOUD_VOD: `${ACCESS_PROVIDERS.JDCLOUD}-vod`,
  KONG: `${ACCESS_PROVIDERS.KONG}`,
  KUBERNETES_SECRET: `${ACCESS_PROVIDERS.KUBERNETES}-secret`,
  LECDN: `${ACCESS_PROVIDERS.LECDN}`,
  LOCAL: `${ACCESS_PROVIDERS.LOCAL}`,
  NETLIFY_SITE: `${ACCESS_PROVIDERS.NETLIFY}-site`,
  PROXMOXVE: `${ACCESS_PROVIDERS.PROXMOXVE}`,
  QINIU_CDN: `${ACCESS_PROVIDERS.QINIU}-cdn`,
  QINIU_KODO: `${ACCESS_PROVIDERS.QINIU}-kodo`,
  QINIU_PILI: `${ACCESS_PROVIDERS.QINIU}-pili`,
  RAINYUN_RCDN: `${ACCESS_PROVIDERS.RAINYUN}-rcdn`,
  RATPANEL_CONSOLE: `${ACCESS_PROVIDERS.RATPANEL}-console`,
  RATPANEL_SITE: `${ACCESS_PROVIDERS.RATPANEL}-site`,
  SAFELINE: `${ACCESS_PROVIDERS.SAFELINE}`,
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
  NAS: "nas",
  OTHER: "other",
} as const);

export type DeploymentCategoryType = (typeof DEPLOYMENT_CATEGORIES)[keyof typeof DEPLOYMENT_CATEGORIES];

export type DeploymentProvider = {
  type: DeploymentProviderType;
  name: string;
  icon: string;
  provider: AccessProviderType;
  category: DeploymentCategoryType;
  builtin: boolean;
};

export const deploymentProvidersMap: Map<DeploymentProvider["type"] | string, DeploymentProvider> = new Map(
  /*
     注意：此处的顺序决定显示在前端的顺序。
     NOTICE: The following order determines the order displayed at the frontend.
    */
  [
    [DEPLOYMENT_PROVIDERS.LOCAL, "provider.local", DEPLOYMENT_CATEGORIES.OTHER, "builtin"],
    [DEPLOYMENT_PROVIDERS.SSH, "provider.ssh", DEPLOYMENT_CATEGORIES.OTHER],
    [DEPLOYMENT_PROVIDERS.WEBHOOK, "provider.webhook", DEPLOYMENT_CATEGORIES.OTHER],
    [DEPLOYMENT_PROVIDERS.KUBERNETES_SECRET, "provider.kubernetes.secret", DEPLOYMENT_CATEGORIES.OTHER],
    [DEPLOYMENT_PROVIDERS.ALIYUN_OSS, "provider.aliyun.oss", DEPLOYMENT_CATEGORIES.STORAGE],
    [DEPLOYMENT_PROVIDERS.ALIYUN_CDN, "provider.aliyun.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.ALIYUN_DCDN, "provider.aliyun.dcdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.ALIYUN_ESA, "provider.aliyun.esa", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.ALIYUN_CLB, "provider.aliyun.clb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
    [DEPLOYMENT_PROVIDERS.ALIYUN_ALB, "provider.aliyun.alb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
    [DEPLOYMENT_PROVIDERS.ALIYUN_NLB, "provider.aliyun.nlb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
    [DEPLOYMENT_PROVIDERS.ALIYUN_WAF, "provider.aliyun.waf", DEPLOYMENT_CATEGORIES.FIREWALL],
    [DEPLOYMENT_PROVIDERS.ALIYUN_DDOS, "provider.aliyun.ddos", DEPLOYMENT_CATEGORIES.FIREWALL],
    [DEPLOYMENT_PROVIDERS.ALIYUN_LIVE, "provider.aliyun.live", DEPLOYMENT_CATEGORIES.AV],
    [DEPLOYMENT_PROVIDERS.ALIYUN_VOD, "provider.aliyun.vod", DEPLOYMENT_CATEGORIES.AV],
    [DEPLOYMENT_PROVIDERS.ALIYUN_FC, "provider.aliyun.fc", DEPLOYMENT_CATEGORIES.SERVERLESS],
    [DEPLOYMENT_PROVIDERS.ALIYUN_APIGW, "provider.aliyun.apigw", DEPLOYMENT_CATEGORIES.APIGATEWAY],
    [DEPLOYMENT_PROVIDERS.ALIYUN_GA, "provider.aliyun.ga", DEPLOYMENT_CATEGORIES.ACCELERATOR],
    [DEPLOYMENT_PROVIDERS.ALIYUN_CAS, "provider.aliyun.cas_upload", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.ALIYUN_CAS_DEPLOY, "provider.aliyun.cas_deploy", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_COS, "provider.tencentcloud.cos", DEPLOYMENT_CATEGORIES.STORAGE],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CDN, "provider.tencentcloud.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_ECDN, "provider.tencentcloud.ecdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_EO, "provider.tencentcloud.eo", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CLB, "provider.tencentcloud.clb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_WAF, "provider.tencentcloud.waf", DEPLOYMENT_CATEGORIES.FIREWALL],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CSS, "provider.tencentcloud.css", DEPLOYMENT_CATEGORIES.AV],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_VOD, "provider.tencentcloud.vod", DEPLOYMENT_CATEGORIES.AV],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SCF, "provider.tencentcloud.scf", DEPLOYMENT_CATEGORIES.SERVERLESS],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_GAAP, "provider.tencentcloud.gaap", DEPLOYMENT_CATEGORIES.ACCELERATOR],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL, "provider.tencentcloud.ssl_upload", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL_DEPLOY, "provider.tencentcloud.ssl_deploy", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL_UPDATE, "provider.tencentcloud.ssl_update", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.BAIDUCLOUD_CDN, "provider.baiducloud.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.BAIDUCLOUD_BLB, "provider.baiducloud.blb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
    [DEPLOYMENT_PROVIDERS.BAIDUCLOUD_APPBLB, "provider.baiducloud.appblb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
    [DEPLOYMENT_PROVIDERS.BAIDUCLOUD_CERT, "provider.baiducloud.cert_upload", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_CDN, "provider.huaweicloud.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_ELB, "provider.huaweicloud.elb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
    [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_WAF, "provider.huaweicloud.waf", DEPLOYMENT_CATEGORIES.FIREWALL],
    [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_SCM, "provider.huaweicloud.scm_upload", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.VOLCENGINE_TOS, "provider.volcengine.tos", DEPLOYMENT_CATEGORIES.STORAGE],
    [DEPLOYMENT_PROVIDERS.VOLCENGINE_CDN, "provider.volcengine.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.VOLCENGINE_DCDN, "provider.volcengine.dcdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.VOLCENGINE_CLB, "provider.volcengine.clb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
    [DEPLOYMENT_PROVIDERS.VOLCENGINE_ALB, "provider.volcengine.alb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
    [DEPLOYMENT_PROVIDERS.VOLCENGINE_IMAGEX, "provider.volcengine.imagex", DEPLOYMENT_CATEGORIES.STORAGE],
    [DEPLOYMENT_PROVIDERS.VOLCENGINE_LIVE, "provider.volcengine.live", DEPLOYMENT_CATEGORIES.AV],
    [DEPLOYMENT_PROVIDERS.VOLCENGINE_CERTCENTER, "provider.volcengine.certcenter_upload", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.JDCLOUD_ALB, "provider.jdcloud.alb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
    [DEPLOYMENT_PROVIDERS.JDCLOUD_CDN, "provider.jdcloud.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.JDCLOUD_LIVE, "provider.jdcloud.live", DEPLOYMENT_CATEGORIES.AV],
    [DEPLOYMENT_PROVIDERS.JDCLOUD_VOD, "provider.jdcloud.vod", DEPLOYMENT_CATEGORIES.AV],
    [DEPLOYMENT_PROVIDERS.QINIU_KODO, "provider.qiniu.kodo", DEPLOYMENT_CATEGORIES.STORAGE],
    [DEPLOYMENT_PROVIDERS.QINIU_CDN, "provider.qiniu.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.QINIU_PILI, "provider.qiniu.pili", DEPLOYMENT_CATEGORIES.AV],
    [DEPLOYMENT_PROVIDERS.UPYUN_FILE, "provider.upyun.file", DEPLOYMENT_CATEGORIES.STORAGE],
    [DEPLOYMENT_PROVIDERS.UPYUN_CDN, "provider.upyun.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.BAISHAN_CDN, "provider.baishan.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.WANGSU_CDN, "provider.wangsu.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.WANGSU_CDNPRO, "provider.wangsu.cdnpro", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.WANGSU_CERTIFICATE, "provider.wangsu.certificate_upload", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.DOGECLOUD_CDN, "provider.dogecloud.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.BYTEPLUS_CDN, "provider.byteplus.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.UCLOUD_US3, "provider.ucloud.us3", DEPLOYMENT_CATEGORIES.STORAGE],
    [DEPLOYMENT_PROVIDERS.UCLOUD_UCDN, "provider.ucloud.ucdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.CTCCCLOUD_CDN, "provider.ctcccloud.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.CTCCCLOUD_ICDN, "provider.ctcccloud.icdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.CTCCCLOUD_AO, "provider.ctcccloud.ao", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.CTCCCLOUD_ELB, "provider.ctcccloud.elb", DEPLOYMENT_CATEGORIES.LOADBALANCE],
    [DEPLOYMENT_PROVIDERS.CTCCCLOUD_LVDN, "provider.ctcccloud.lvdn", DEPLOYMENT_CATEGORIES.AV],
    [DEPLOYMENT_PROVIDERS.CTCCCLOUD_CMS, "provider.ctcccloud.cms_upload", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.RAINYUN_RCDN, "provider.rainyun.rcdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.UNICLOUD_WEBHOST, "provider.unicloud.webhost", DEPLOYMENT_CATEGORIES.WEBSITE],
    [DEPLOYMENT_PROVIDERS.AWS_CLOUDFRONT, "provider.aws.cloudfront", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.AWS_ACM, "provider.aws.acm", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.AWS_IAM, "provider.aws.iam", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.AZURE_KEYVAULT, "provider.azure.keyvault", DEPLOYMENT_CATEGORIES.SSL],
    [DEPLOYMENT_PROVIDERS.BUNNY_CDN, "provider.bunny.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.CACHEFLY, "provider.cachefly", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.EDGIO_APPLICATIONS, "provider.edgio.applications", DEPLOYMENT_CATEGORIES.WEBSITE],
    [DEPLOYMENT_PROVIDERS.GCORE_CDN, "provider.gcore.cdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.NETLIFY_SITE, "provider.netlify.site", DEPLOYMENT_CATEGORIES.WEBSITE],
    [DEPLOYMENT_PROVIDERS.CDNFLY, "provider.cdnfly", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.FLEXCDN, "provider.flexcdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.GOEDGE, "provider.goedge", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS.LECDN, "provider.lecdn", DEPLOYMENT_CATEGORIES.CDN],
    [DEPLOYMENT_PROVIDERS["1PANEL_SITE"], "provider.1panel.site", DEPLOYMENT_CATEGORIES.WEBSITE],
    [DEPLOYMENT_PROVIDERS["1PANEL_CONSOLE"], "provider.1panel.console", DEPLOYMENT_CATEGORIES.OTHER],
    [DEPLOYMENT_PROVIDERS.BAOTAPANEL_SITE, "provider.baotapanel.site", DEPLOYMENT_CATEGORIES.WEBSITE],
    [DEPLOYMENT_PROVIDERS.BAOTAPANEL_CONSOLE, "provider.baotapanel.console", DEPLOYMENT_CATEGORIES.OTHER],
    [DEPLOYMENT_PROVIDERS.RATPANEL_SITE, "provider.ratpanel.site", DEPLOYMENT_CATEGORIES.WEBSITE],
    [DEPLOYMENT_PROVIDERS.RATPANEL_CONSOLE, "provider.ratpanel.console", DEPLOYMENT_CATEGORIES.OTHER],
    [DEPLOYMENT_PROVIDERS.BAOTAWAF_SITE, "provider.baotawaf.site", DEPLOYMENT_CATEGORIES.FIREWALL],
    [DEPLOYMENT_PROVIDERS.BAOTAWAF_CONSOLE, "provider.baotawaf.console", DEPLOYMENT_CATEGORIES.OTHER],
    [DEPLOYMENT_PROVIDERS.SAFELINE, "provider.safeline", DEPLOYMENT_CATEGORIES.FIREWALL],
    [DEPLOYMENT_PROVIDERS.APISIX, "provider.apisix", DEPLOYMENT_CATEGORIES.APIGATEWAY],
    [DEPLOYMENT_PROVIDERS.KONG, "provider.kong", DEPLOYMENT_CATEGORIES.APIGATEWAY],
    [DEPLOYMENT_PROVIDERS.PROXMOXVE, "provider.proxmoxve", DEPLOYMENT_CATEGORIES.NAS],
  ].map(([type, name, category, builtin]) => [
    type,
    {
      type: type as DeploymentProviderType,
      name: name,
      icon: accessProvidersMap.get(type.split("-")[0])!.icon,
      provider: type.split("-")[0] as AccessProviderType,
      category: category as DeploymentCategoryType,
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

export type NotificationProviderType = (typeof CA_PROVIDERS)[keyof typeof CA_PROVIDERS];

export type NotificationProvider = {
  type: NotificationProviderType;
  name: string;
  icon: string;
  provider: AccessProviderType;
};

export const notificationProvidersMap: Map<NotificationProvider["type"] | string, NotificationProvider> = new Map(
  /*
    注意：此处的顺序决定显示在前端的顺序。
    NOTICE: The following order determines the order displayed at the frontend.
   */
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
  ].map(([type]) => [
    type,
    {
      type: type as CAProviderType,
      name: accessProvidersMap.get(type.split("-")[0])!.name,
      icon: accessProvidersMap.get(type.split("-")[0])!.icon,
      provider: type.split("-")[0] as AccessProviderType,
    },
  ])
);
// #endregion
