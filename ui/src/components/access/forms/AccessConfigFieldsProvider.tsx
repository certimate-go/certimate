import { useEffect, useState } from "react";

import { ACCESS_PROVIDERS, type AccessProviderType } from "@/domain/provider";

import AccessConfigFieldsProvider1Panel from "./AccessConfigFieldsProvider1Panel";
import AccessConfigFieldsProviderACMECA from "./AccessConfigFieldsProviderACMECA";
import AccessConfigFieldsProviderACMEDNS from "./AccessConfigFieldsProviderACMEDNS";
import AccessConfigFieldsProviderACMEHttpReq from "./AccessConfigFieldsProviderACMEHttpReq";
import AccessConfigFieldsProviderActalisSSL from "./AccessConfigFieldsProviderActalisSSL";
import AccessConfigFieldsProviderAkamai from "./AccessConfigFieldsProviderAkamai";
import AccessConfigFieldsProviderAliyun from "./AccessConfigFieldsProviderAliyun";
import AccessConfigFieldsProviderAPISIX from "./AccessConfigFieldsProviderAPISIX";
import AccessConfigFieldsProviderArvanCloud from "./AccessConfigFieldsProviderArvanCloud";
import AccessConfigFieldsProviderAWS from "./AccessConfigFieldsProviderAWS";
import AccessConfigFieldsProviderAzure from "./AccessConfigFieldsProviderAzure";
import AccessConfigFieldsProviderBaiduCloud from "./AccessConfigFieldsProviderBaiduCloud";
import AccessConfigFieldsProviderBaishan from "./AccessConfigFieldsProviderBaishan";
import AccessConfigFieldsProviderBaotaPanel from "./AccessConfigFieldsProviderBaotaPanel";
import AccessConfigFieldsProviderBaotaPanelGo from "./AccessConfigFieldsProviderBaotaPanelGo";
import AccessConfigFieldsProviderBaotaWAF from "./AccessConfigFieldsProviderBaotaWAF";
import AccessConfigFieldsProviderBookMyName from "./AccessConfigFieldsProviderBookMyName";
import AccessConfigFieldsProviderBunny from "./AccessConfigFieldsProviderBunny";
import AccessConfigFieldsProviderBytePlus from "./AccessConfigFieldsProviderBytePlus";
import AccessConfigFieldsProviderCacheFly from "./AccessConfigFieldsProviderCacheFly";
import AccessConfigFieldsProviderCdnfly from "./AccessConfigFieldsProviderCdnfly";
import AccessConfigFieldsProviderCloudflare from "./AccessConfigFieldsProviderCloudflare";
import AccessConfigFieldsProviderClouDNS from "./AccessConfigFieldsProviderClouDNS";
import AccessConfigFieldsProviderCMCCCloud from "./AccessConfigFieldsProviderCMCCCloud";
import AccessConfigFieldsProviderConstellix from "./AccessConfigFieldsProviderConstellix";
import AccessConfigFieldsProviderCTCCCloud from "./AccessConfigFieldsProviderCTCCCloud";
import AccessConfigFieldsProviderDeSEC from "./AccessConfigFieldsProviderDeSEC";
import AccessConfigFieldsProviderDigitalOcean from "./AccessConfigFieldsProviderDigitalOcean";
import AccessConfigFieldsProviderDingTalkBot from "./AccessConfigFieldsProviderDingTalkBot";
import AccessConfigFieldsProviderDiscordBot from "./AccessConfigFieldsProviderDiscordBot";
import AccessConfigFieldsProviderDNSLA from "./AccessConfigFieldsProviderDNSLA";
import AccessConfigFieldsProviderDNSMadeEasy from "./AccessConfigFieldsProviderDNSMadeEasy";
import AccessConfigFieldsProviderDogeCloud from "./AccessConfigFieldsProviderDogeCloud";
import AccessConfigFieldsProviderDuckDNS from "./AccessConfigFieldsProviderDuckDNS";
import AccessConfigFieldsProviderDynu from "./AccessConfigFieldsProviderDynu";
import AccessConfigFieldsProviderDynv6 from "./AccessConfigFieldsProviderDynv6";
import AccessConfigFieldsProviderEmail from "./AccessConfigFieldsProviderEmail";
import AccessConfigFieldsProviderFlexCDN from "./AccessConfigFieldsProviderFlexCDN";
import AccessConfigFieldsProviderGandinet from "./AccessConfigFieldsProviderGandinet";
import AccessConfigFieldsProviderGcore from "./AccessConfigFieldsProviderGcore";
import AccessConfigFieldsProviderGlobalSignAtlas from "./AccessConfigFieldsProviderGlobalSignAtlas";
import AccessConfigFieldsProviderGname from "./AccessConfigFieldsProviderGname";
import AccessConfigFieldsProviderGoDaddy from "./AccessConfigFieldsProviderGoDaddy";
import AccessConfigFieldsProviderGoEdge from "./AccessConfigFieldsProviderGoEdge";
import AccessConfigFieldsProviderGoogleTrustServices from "./AccessConfigFieldsProviderGoogleTrustServices";
import AccessConfigFieldsProviderHetzner from "./AccessConfigFieldsProviderHetzner";
import AccessConfigFieldsProviderHostingde from "./AccessConfigFieldsProviderHostingde";
import AccessConfigFieldsProviderHostinger from "./AccessConfigFieldsProviderHostinger";
import AccessConfigFieldsProviderHuaweiCloud from "./AccessConfigFieldsProviderHuaweiCloud";
import AccessConfigFieldsProviderInfomaniak from "./AccessConfigFieldsProviderInfomaniak";
import AccessConfigFieldsProviderIONOS from "./AccessConfigFieldsProviderIONOS";
import AccessConfigFieldsProviderJDCloud from "./AccessConfigFieldsProviderJDCloud";
import AccessConfigFieldsProviderKong from "./AccessConfigFieldsProviderKong";
import AccessConfigFieldsProviderKsyun from "./AccessConfigFieldsProviderKsyun";
import AccessConfigFieldsProviderKubernetes from "./AccessConfigFieldsProviderKubernetes";
import AccessConfigFieldsProviderLarkBot from "./AccessConfigFieldsProviderLarkBot";
import AccessConfigFieldsProviderLeCDN from "./AccessConfigFieldsProviderLeCDN";
import AccessConfigFieldsProviderLinode from "./AccessConfigFieldsProviderLinode";
import AccessConfigFieldsProviderMattermost from "./AccessConfigFieldsProviderMattermost";
import AccessConfigFieldsProviderNamecheap from "./AccessConfigFieldsProviderNamecheap";
import AccessConfigFieldsProviderNameDotCom from "./AccessConfigFieldsProviderNameDotCom";
import AccessConfigFieldsProviderNameSilo from "./AccessConfigFieldsProviderNameSilo";
import AccessConfigFieldsProviderNetcup from "./AccessConfigFieldsProviderNetcup";
import AccessConfigFieldsProviderNetlify from "./AccessConfigFieldsProviderNetlify";
import AccessConfigFieldsProviderNS1 from "./AccessConfigFieldsProviderNS1";
import AccessConfigFieldsProviderOVHcloud from "./AccessConfigFieldsProviderOVHcloud";
import AccessConfigFieldsProviderPorkbun from "./AccessConfigFieldsProviderPorkbun";
import AccessConfigFieldsProviderPowerDNS from "./AccessConfigFieldsProviderPowerDNS";
import AccessConfigFieldsProviderProxmoxVE from "./AccessConfigFieldsProviderProxmoxVE";
import AccessConfigFieldsProviderQiniu from "./AccessConfigFieldsProviderQiniu";
import AccessConfigFieldsProviderRainYun from "./AccessConfigFieldsProviderRainYun";
import AccessConfigFieldsProviderRatPanel from "./AccessConfigFieldsProviderRatPanel";
import AccessConfigFieldsProviderRFC2136 from "./AccessConfigFieldsProviderRFC2136";
import AccessConfigFieldsProviderSafeLine from "./AccessConfigFieldsProviderSafeLine";
import AccessConfigFieldsProviderSectigo from "./AccessConfigFieldsProviderSectigo";
import AccessConfigFieldsProviderSlackBot from "./AccessConfigFieldsProviderSlackBot";
import AccessConfigFieldsProviderSpaceship from "./AccessConfigFieldsProviderSpaceship";
import AccessConfigFieldsProviderSSH from "./AccessConfigFieldsProviderSSH";
import AccessConfigFieldsProviderSSLCom from "./AccessConfigFieldsProviderSSLCom";
import AccessConfigFieldsProviderTechnitiumDNS from "./AccessConfigFieldsProviderTechnitiumDNS";
import AccessConfigFieldsProviderTelegramBot from "./AccessConfigFieldsProviderTelegramBot";
import AccessConfigFieldsProviderTencentCloud from "./AccessConfigFieldsProviderTencentCloud";
import AccessConfigFieldsProviderUCloud from "./AccessConfigFieldsProviderUCloud";
import AccessConfigFieldsProviderUniCloud from "./AccessConfigFieldsProviderUniCloud";
import AccessConfigFieldsProviderUpyun from "./AccessConfigFieldsProviderUpyun";
import AccessConfigFieldsProviderVercel from "./AccessConfigFieldsProviderVercel";
import AccessConfigFieldsProviderVolcEngine from "./AccessConfigFieldsProviderVolcEngine";
import AccessConfigFieldsProviderVultr from "./AccessConfigFieldsProviderVultr";
import AccessConfigFieldsProviderWangsu from "./AccessConfigFieldsProviderWangsu";
import AccessConfigFieldsProviderWebhook from "./AccessConfigFieldsProviderWebhook";
import AccessConfigFieldsProviderWeComBot from "./AccessConfigFieldsProviderWeComBot";
import AccessConfigFieldsProviderWestcn from "./AccessConfigFieldsProviderWestcn";
import AccessConfigFieldsProviderZeroSSL from "./AccessConfigFieldsProviderZeroSSL";

const providerComponentMap: Partial<Record<AccessProviderType, React.ComponentType<any>>> = {
  /*
    注意：如果追加新的子组件，请保持以 ASCII 排序。
    NOTICE: If you add new child component, please keep ASCII order.
    */
  [ACCESS_PROVIDERS["1PANEL"]]: AccessConfigFieldsProvider1Panel,
  [ACCESS_PROVIDERS.ACMECA]: AccessConfigFieldsProviderACMECA,
  [ACCESS_PROVIDERS.ACMEDNS]: AccessConfigFieldsProviderACMEDNS,
  [ACCESS_PROVIDERS.ACMEHTTPREQ]: AccessConfigFieldsProviderACMEHttpReq,
  [ACCESS_PROVIDERS.ACTALISSSL]: AccessConfigFieldsProviderActalisSSL,
  [ACCESS_PROVIDERS.AKAMAI]: AccessConfigFieldsProviderAkamai,
  [ACCESS_PROVIDERS.ALIYUN]: AccessConfigFieldsProviderAliyun,
  [ACCESS_PROVIDERS.APISIX]: AccessConfigFieldsProviderAPISIX,
  [ACCESS_PROVIDERS.ARVANCLOUD]: AccessConfigFieldsProviderArvanCloud,
  [ACCESS_PROVIDERS.AWS]: AccessConfigFieldsProviderAWS,
  [ACCESS_PROVIDERS.AZURE]: AccessConfigFieldsProviderAzure,
  [ACCESS_PROVIDERS.BAIDUCLOUD]: AccessConfigFieldsProviderBaiduCloud,
  [ACCESS_PROVIDERS.BAISHAN]: AccessConfigFieldsProviderBaishan,
  [ACCESS_PROVIDERS.BAOTAPANEL]: AccessConfigFieldsProviderBaotaPanel,
  [ACCESS_PROVIDERS.BAOTAPANELGO]: AccessConfigFieldsProviderBaotaPanelGo,
  [ACCESS_PROVIDERS.BAOTAWAF]: AccessConfigFieldsProviderBaotaWAF,
  [ACCESS_PROVIDERS.BOOKMYNAME]: AccessConfigFieldsProviderBookMyName,
  [ACCESS_PROVIDERS.BUNNY]: AccessConfigFieldsProviderBunny,
  [ACCESS_PROVIDERS.BYTEPLUS]: AccessConfigFieldsProviderBytePlus,
  [ACCESS_PROVIDERS.CACHEFLY]: AccessConfigFieldsProviderCacheFly,
  [ACCESS_PROVIDERS.CDNFLY]: AccessConfigFieldsProviderCdnfly,
  [ACCESS_PROVIDERS.CLOUDFLARE]: AccessConfigFieldsProviderCloudflare,
  [ACCESS_PROVIDERS.CLOUDNS]: AccessConfigFieldsProviderClouDNS,
  [ACCESS_PROVIDERS.CMCCCLOUD]: AccessConfigFieldsProviderCMCCCloud,
  [ACCESS_PROVIDERS.CONSTELLIX]: AccessConfigFieldsProviderConstellix,
  [ACCESS_PROVIDERS.CTCCCLOUD]: AccessConfigFieldsProviderCTCCCloud,
  [ACCESS_PROVIDERS.DESEC]: AccessConfigFieldsProviderDeSEC,
  [ACCESS_PROVIDERS.DIGITALOCEAN]: AccessConfigFieldsProviderDigitalOcean,
  [ACCESS_PROVIDERS.DINGTALKBOT]: AccessConfigFieldsProviderDingTalkBot,
  [ACCESS_PROVIDERS.DISCORDBOT]: AccessConfigFieldsProviderDiscordBot,
  [ACCESS_PROVIDERS.DNSLA]: AccessConfigFieldsProviderDNSLA,
  [ACCESS_PROVIDERS.DNSMADEEASY]: AccessConfigFieldsProviderDNSMadeEasy,
  [ACCESS_PROVIDERS.DOGECLOUD]: AccessConfigFieldsProviderDogeCloud,
  [ACCESS_PROVIDERS.DUCKDNS]: AccessConfigFieldsProviderDuckDNS,
  [ACCESS_PROVIDERS.DYNU]: AccessConfigFieldsProviderDynu,
  [ACCESS_PROVIDERS.DYNV6]: AccessConfigFieldsProviderDynv6,
  [ACCESS_PROVIDERS.EMAIL]: AccessConfigFieldsProviderEmail,
  [ACCESS_PROVIDERS.FLEXCDN]: AccessConfigFieldsProviderFlexCDN,
  [ACCESS_PROVIDERS.GANDINET]: AccessConfigFieldsProviderGandinet,
  [ACCESS_PROVIDERS.GCORE]: AccessConfigFieldsProviderGcore,
  [ACCESS_PROVIDERS.GNAME]: AccessConfigFieldsProviderGname,
  [ACCESS_PROVIDERS.GODADDY]: AccessConfigFieldsProviderGoDaddy,
  [ACCESS_PROVIDERS.GOEDGE]: AccessConfigFieldsProviderGoEdge,
  [ACCESS_PROVIDERS.GLOBALSIGNATLAS]: AccessConfigFieldsProviderGlobalSignAtlas,
  [ACCESS_PROVIDERS.GOOGLETRUSTSERVICES]: AccessConfigFieldsProviderGoogleTrustServices,
  [ACCESS_PROVIDERS.HETZNER]: AccessConfigFieldsProviderHetzner,
  [ACCESS_PROVIDERS.HOSTINGDE]: AccessConfigFieldsProviderHostingde,
  [ACCESS_PROVIDERS.HOSTINGER]: AccessConfigFieldsProviderHostinger,
  [ACCESS_PROVIDERS.HUAWEICLOUD]: AccessConfigFieldsProviderHuaweiCloud,
  [ACCESS_PROVIDERS.IONOS]: AccessConfigFieldsProviderIONOS,
  [ACCESS_PROVIDERS.JDCLOUD]: AccessConfigFieldsProviderJDCloud,
  [ACCESS_PROVIDERS.KONG]: AccessConfigFieldsProviderKong,
  [ACCESS_PROVIDERS.KUBERNETES]: AccessConfigFieldsProviderKubernetes,
  [ACCESS_PROVIDERS.KSYUN]: AccessConfigFieldsProviderKsyun,
  [ACCESS_PROVIDERS.LARKBOT]: AccessConfigFieldsProviderLarkBot,
  [ACCESS_PROVIDERS.LECDN]: AccessConfigFieldsProviderLeCDN,
  [ACCESS_PROVIDERS.INFOMANIAK]: AccessConfigFieldsProviderInfomaniak,
  [ACCESS_PROVIDERS.LINODE]: AccessConfigFieldsProviderLinode,
  [ACCESS_PROVIDERS.MATTERMOST]: AccessConfigFieldsProviderMattermost,
  [ACCESS_PROVIDERS.NAMECHEAP]: AccessConfigFieldsProviderNamecheap,
  [ACCESS_PROVIDERS.NAMEDOTCOM]: AccessConfigFieldsProviderNameDotCom,
  [ACCESS_PROVIDERS.NAMESILO]: AccessConfigFieldsProviderNameSilo,
  [ACCESS_PROVIDERS.NETCUP]: AccessConfigFieldsProviderNetcup,
  [ACCESS_PROVIDERS.NETLIFY]: AccessConfigFieldsProviderNetlify,
  [ACCESS_PROVIDERS.NS1]: AccessConfigFieldsProviderNS1,
  [ACCESS_PROVIDERS.OVHCLOUD]: AccessConfigFieldsProviderOVHcloud,
  [ACCESS_PROVIDERS.PORKBUN]: AccessConfigFieldsProviderPorkbun,
  [ACCESS_PROVIDERS.POWERDNS]: AccessConfigFieldsProviderPowerDNS,
  [ACCESS_PROVIDERS.PROXMOXVE]: AccessConfigFieldsProviderProxmoxVE,
  [ACCESS_PROVIDERS.QINIU]: AccessConfigFieldsProviderQiniu,
  [ACCESS_PROVIDERS.RAINYUN]: AccessConfigFieldsProviderRainYun,
  [ACCESS_PROVIDERS.RATPANEL]: AccessConfigFieldsProviderRatPanel,
  [ACCESS_PROVIDERS.RFC2136]: AccessConfigFieldsProviderRFC2136,
  [ACCESS_PROVIDERS.SAFELINE]: AccessConfigFieldsProviderSafeLine,
  [ACCESS_PROVIDERS.SECTIGO]: AccessConfigFieldsProviderSectigo,
  [ACCESS_PROVIDERS.SLACKBOT]: AccessConfigFieldsProviderSlackBot,
  [ACCESS_PROVIDERS.SPACESHIP]: AccessConfigFieldsProviderSpaceship,
  [ACCESS_PROVIDERS.SSLCOM]: AccessConfigFieldsProviderSSLCom,
  [ACCESS_PROVIDERS.SSH]: AccessConfigFieldsProviderSSH,
  [ACCESS_PROVIDERS.TECHNITIUMDNS]: AccessConfigFieldsProviderTechnitiumDNS,
  [ACCESS_PROVIDERS.TELEGRAMBOT]: AccessConfigFieldsProviderTelegramBot,
  [ACCESS_PROVIDERS.TENCENTCLOUD]: AccessConfigFieldsProviderTencentCloud,
  [ACCESS_PROVIDERS.UCLOUD]: AccessConfigFieldsProviderUCloud,
  [ACCESS_PROVIDERS.UNICLOUD]: AccessConfigFieldsProviderUniCloud,
  [ACCESS_PROVIDERS.UPYUN]: AccessConfigFieldsProviderUpyun,
  [ACCESS_PROVIDERS.VERCEL]: AccessConfigFieldsProviderVercel,
  [ACCESS_PROVIDERS.VOLCENGINE]: AccessConfigFieldsProviderVolcEngine,
  [ACCESS_PROVIDERS.VULTR]: AccessConfigFieldsProviderVultr,
  [ACCESS_PROVIDERS.WANGSU]: AccessConfigFieldsProviderWangsu,
  [ACCESS_PROVIDERS.WEBHOOK]: AccessConfigFieldsProviderWebhook,
  [ACCESS_PROVIDERS.WECOMBOT]: AccessConfigFieldsProviderWeComBot,
  [ACCESS_PROVIDERS.WESTCN]: AccessConfigFieldsProviderWestcn,
  [ACCESS_PROVIDERS.ZEROSSL]: AccessConfigFieldsProviderZeroSSL,
};

const useComponent = (provider: string, { initProps, deps = [] }: { initProps?: (provider: string) => any; deps?: unknown[] }) => {
  const initComponent = () => {
    const Component = providerComponentMap[provider as AccessProviderType];
    if (!Component) return null;

    const props = initProps?.(provider);
    if (props) {
      return <Component {...props} />;
    }

    return <Component />;
  };

  const [component, setComponent] = useState(() => initComponent());

  useEffect(() => setComponent(initComponent()), [provider]);
  useEffect(() => setComponent(initComponent()), deps);

  return component;
};

const _default = {
  useComponent,
};

export default _default;
