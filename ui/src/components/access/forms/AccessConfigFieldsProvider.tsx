import { useEffect, useState } from "react";

import { type ProviderSchemaEnvelope, getProviderSchema } from "@/api/providerschema";
import SchemaConfigFields from "@/components/workflow/designer/forms/SchemaConfigFields";
import { ACCESS_PROVIDERS, type AccessProviderType } from "@/domain/provider";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";

import AccessConfigFieldsProvider35cn from "./AccessConfigFieldsProvider35cn";
import AccessConfigFieldsProvider51DNScom from "./AccessConfigFieldsProvider51DNScom";
import AccessConfigFieldsProviderACMECA from "./AccessConfigFieldsProviderACMECA";
import AccessConfigFieldsProviderACMEDNS from "./AccessConfigFieldsProviderACMEDNS";
import AccessConfigFieldsProviderACMEHttpReq from "./AccessConfigFieldsProviderACMEHttpReq";
import AccessConfigFieldsProviderActalisSSL from "./AccessConfigFieldsProviderActalisSSL";
import AccessConfigFieldsProviderAkamai from "./AccessConfigFieldsProviderAkamai";
import AccessConfigFieldsProviderArvanCloud from "./AccessConfigFieldsProviderArvanCloud";
import AccessConfigFieldsProviderAWS from "./AccessConfigFieldsProviderAWS";
import AccessConfigFieldsProviderAzure from "./AccessConfigFieldsProviderAzure";
import AccessConfigFieldsProviderBaiduCloud from "./AccessConfigFieldsProviderBaiduCloud";
import AccessConfigFieldsProviderBeget from "./AccessConfigFieldsProviderBeget";
import AccessConfigFieldsProviderBookMyName from "./AccessConfigFieldsProviderBookMyName";
import AccessConfigFieldsProviderBunny from "./AccessConfigFieldsProviderBunny";
import AccessConfigFieldsProviderBytePlus from "./AccessConfigFieldsProviderBytePlus";
import AccessConfigFieldsProviderCloudflare from "./AccessConfigFieldsProviderCloudflare";
import AccessConfigFieldsProviderClouDNS from "./AccessConfigFieldsProviderClouDNS";
import AccessConfigFieldsProviderCMCCCloud from "./AccessConfigFieldsProviderCMCCCloud";
import AccessConfigFieldsProviderConoHaVPS from "./AccessConfigFieldsProviderConoHaVPS";
import AccessConfigFieldsProviderConstellix from "./AccessConfigFieldsProviderConstellix";
import AccessConfigFieldsProviderCPanel from "./AccessConfigFieldsProviderCPanel";
import AccessConfigFieldsProviderCTCCCloud from "./AccessConfigFieldsProviderCTCCCloud";
import AccessConfigFieldsProviderDeSEC from "./AccessConfigFieldsProviderDeSEC";
import AccessConfigFieldsProviderDigiCert from "./AccessConfigFieldsProviderDigiCert";
import AccessConfigFieldsProviderDigitalOcean from "./AccessConfigFieldsProviderDigitalOcean";
import AccessConfigFieldsProviderDNSExit from "./AccessConfigFieldsProviderDNSExit";
import AccessConfigFieldsProviderDNSLA from "./AccessConfigFieldsProviderDNSLA";
import AccessConfigFieldsProviderDNSMadeEasy from "./AccessConfigFieldsProviderDNSMadeEasy";
import AccessConfigFieldsProviderDuckDNS from "./AccessConfigFieldsProviderDuckDNS";
import AccessConfigFieldsProviderDynadot from "./AccessConfigFieldsProviderDynadot";
import AccessConfigFieldsProviderDynu from "./AccessConfigFieldsProviderDynu";
import AccessConfigFieldsProviderDynv6 from "./AccessConfigFieldsProviderDynv6";
import AccessConfigFieldsProviderEmail from "./AccessConfigFieldsProviderEmail";
import AccessConfigFieldsProviderFTP from "./AccessConfigFieldsProviderFTP";
import AccessConfigFieldsProviderGandinet from "./AccessConfigFieldsProviderGandinet";
import AccessConfigFieldsProviderGcore from "./AccessConfigFieldsProviderGcore";
import AccessConfigFieldsProviderGlobalSignAtlas from "./AccessConfigFieldsProviderGlobalSignAtlas";
import AccessConfigFieldsProviderGname from "./AccessConfigFieldsProviderGname";
import AccessConfigFieldsProviderGoDaddy from "./AccessConfigFieldsProviderGoDaddy";
import AccessConfigFieldsProviderGoogleCloud from "./AccessConfigFieldsProviderGoogleCloud";
import AccessConfigFieldsProviderGoogleTrustServices from "./AccessConfigFieldsProviderGoogleTrustServices";
import AccessConfigFieldsProviderHetzner from "./AccessConfigFieldsProviderHetzner";
import AccessConfigFieldsProviderHostingde from "./AccessConfigFieldsProviderHostingde";
import AccessConfigFieldsProviderHostinger from "./AccessConfigFieldsProviderHostinger";
import AccessConfigFieldsProviderHuaweiCloud from "./AccessConfigFieldsProviderHuaweiCloud";
import AccessConfigFieldsProviderInfomaniak from "./AccessConfigFieldsProviderInfomaniak";
import AccessConfigFieldsProviderIONOS from "./AccessConfigFieldsProviderIONOS";
import AccessConfigFieldsProviderJDCloud from "./AccessConfigFieldsProviderJDCloud";
import AccessConfigFieldsProviderLinode from "./AccessConfigFieldsProviderLinode";
import AccessConfigFieldsProviderLiteSSL from "./AccessConfigFieldsProviderLiteSSL";
import AccessConfigFieldsProviderLuaDNS from "./AccessConfigFieldsProviderLuaDNS";
import AccessConfigFieldsProviderNamecheap from "./AccessConfigFieldsProviderNamecheap";
import AccessConfigFieldsProviderNameDotCom from "./AccessConfigFieldsProviderNameDotCom";
import AccessConfigFieldsProviderNameSilo from "./AccessConfigFieldsProviderNameSilo";
import AccessConfigFieldsProviderNetcup from "./AccessConfigFieldsProviderNetcup";
import AccessConfigFieldsProviderNetlify from "./AccessConfigFieldsProviderNetlify";
import AccessConfigFieldsProviderNS1 from "./AccessConfigFieldsProviderNS1";
import AccessConfigFieldsProviderOracleCloud from "./AccessConfigFieldsProviderOracleCloud";
import AccessConfigFieldsProviderOVHcloud from "./AccessConfigFieldsProviderOVHcloud";
import AccessConfigFieldsProviderPorkbun from "./AccessConfigFieldsProviderPorkbun";
import AccessConfigFieldsProviderPowerDNS from "./AccessConfigFieldsProviderPowerDNS";
import AccessConfigFieldsProviderQingCloud from "./AccessConfigFieldsProviderQingCloud";
import AccessConfigFieldsProviderRainYun from "./AccessConfigFieldsProviderRainYun";
import AccessConfigFieldsProviderRegru from "./AccessConfigFieldsProviderRegru";
import AccessConfigFieldsProviderRFC2136 from "./AccessConfigFieldsProviderRFC2136";
import AccessConfigFieldsProviderRuCenter from "./AccessConfigFieldsProviderRuCenter";
import AccessConfigFieldsProviderS3 from "./AccessConfigFieldsProviderS3";
import AccessConfigFieldsProviderSectigo from "./AccessConfigFieldsProviderSectigo";
import AccessConfigFieldsProviderSimplyCom from "./AccessConfigFieldsProviderSimplyCom";
import AccessConfigFieldsProviderSpaceship from "./AccessConfigFieldsProviderSpaceship";
import AccessConfigFieldsProviderSSH from "./AccessConfigFieldsProviderSSH";
import AccessConfigFieldsProviderSSLCom from "./AccessConfigFieldsProviderSSLCom";
import AccessConfigFieldsProviderTechnitiumDNS from "./AccessConfigFieldsProviderTechnitiumDNS";
import AccessConfigFieldsProviderTencentCloud from "./AccessConfigFieldsProviderTencentCloud";
import AccessConfigFieldsProviderTodayNIC from "./AccessConfigFieldsProviderTodayNIC";
import AccessConfigFieldsProviderUCloud from "./AccessConfigFieldsProviderUCloud";
import AccessConfigFieldsProviderVercel from "./AccessConfigFieldsProviderVercel";
import AccessConfigFieldsProviderVolcEngine from "./AccessConfigFieldsProviderVolcEngine";
import AccessConfigFieldsProviderVultr from "./AccessConfigFieldsProviderVultr";
import AccessConfigFieldsProviderWebhook from "./AccessConfigFieldsProviderWebhook";
import AccessConfigFieldsProviderWestcn from "./AccessConfigFieldsProviderWestcn";
import AccessConfigFieldsProviderXinnet from "./AccessConfigFieldsProviderXinnet";
import AccessConfigFieldsProviderYandexCloud from "./AccessConfigFieldsProviderYandexCloud";
import AccessConfigFieldsProviderZeroSSL from "./AccessConfigFieldsProviderZeroSSL";

const providerComponentMap: Partial<Record<AccessProviderType, React.ComponentType<any>>> = {
  /*
    注意：如果追加新的子组件，请保持以 ASCII 排序。
    NOTICE: If you add new child component, please keep ASCII order.
    */
  [ACCESS_PROVIDERS["35CN"]]: AccessConfigFieldsProvider35cn,
  [ACCESS_PROVIDERS["51DNSCOM"]]: AccessConfigFieldsProvider51DNScom,
  [ACCESS_PROVIDERS.ACMECA]: AccessConfigFieldsProviderACMECA,
  [ACCESS_PROVIDERS.ACMEDNS]: AccessConfigFieldsProviderACMEDNS,
  [ACCESS_PROVIDERS.ACMEHTTPREQ]: AccessConfigFieldsProviderACMEHttpReq,
  [ACCESS_PROVIDERS.ACTALISSSL]: AccessConfigFieldsProviderActalisSSL,
  [ACCESS_PROVIDERS.AKAMAI]: AccessConfigFieldsProviderAkamai,
  [ACCESS_PROVIDERS.ARVANCLOUD]: AccessConfigFieldsProviderArvanCloud,
  [ACCESS_PROVIDERS.AWS]: AccessConfigFieldsProviderAWS,
  [ACCESS_PROVIDERS.AZURE]: AccessConfigFieldsProviderAzure,
  [ACCESS_PROVIDERS.BAIDUCLOUD]: AccessConfigFieldsProviderBaiduCloud,
  [ACCESS_PROVIDERS.BEGET]: AccessConfigFieldsProviderBeget,
  [ACCESS_PROVIDERS.BOOKMYNAME]: AccessConfigFieldsProviderBookMyName,
  [ACCESS_PROVIDERS.BUNNY]: AccessConfigFieldsProviderBunny,
  [ACCESS_PROVIDERS.BYTEPLUS]: AccessConfigFieldsProviderBytePlus,
  [ACCESS_PROVIDERS.CLOUDFLARE]: AccessConfigFieldsProviderCloudflare,
  [ACCESS_PROVIDERS.CLOUDNS]: AccessConfigFieldsProviderClouDNS,
  [ACCESS_PROVIDERS.CMCCCLOUD]: AccessConfigFieldsProviderCMCCCloud,
  [ACCESS_PROVIDERS.CONOHAVPS]: AccessConfigFieldsProviderConoHaVPS,
  [ACCESS_PROVIDERS.CONSTELLIX]: AccessConfigFieldsProviderConstellix,
  [ACCESS_PROVIDERS.CPANEL]: AccessConfigFieldsProviderCPanel,
  [ACCESS_PROVIDERS.CTCCCLOUD]: AccessConfigFieldsProviderCTCCCloud,
  [ACCESS_PROVIDERS.DESEC]: AccessConfigFieldsProviderDeSEC,
  [ACCESS_PROVIDERS.DIGICERT]: AccessConfigFieldsProviderDigiCert,
  [ACCESS_PROVIDERS.DIGITALOCEAN]: AccessConfigFieldsProviderDigitalOcean,
  [ACCESS_PROVIDERS.DNSEXIT]: AccessConfigFieldsProviderDNSExit,
  [ACCESS_PROVIDERS.DNSLA]: AccessConfigFieldsProviderDNSLA,
  [ACCESS_PROVIDERS.DNSMADEEASY]: AccessConfigFieldsProviderDNSMadeEasy,
  [ACCESS_PROVIDERS.DUCKDNS]: AccessConfigFieldsProviderDuckDNS,
  [ACCESS_PROVIDERS.DYNADOT]: AccessConfigFieldsProviderDynadot,
  [ACCESS_PROVIDERS.DYNU]: AccessConfigFieldsProviderDynu,
  [ACCESS_PROVIDERS.DYNV6]: AccessConfigFieldsProviderDynv6,
  [ACCESS_PROVIDERS.EMAIL]: AccessConfigFieldsProviderEmail,
  [ACCESS_PROVIDERS.FTP]: AccessConfigFieldsProviderFTP,
  [ACCESS_PROVIDERS.GANDINET]: AccessConfigFieldsProviderGandinet,
  [ACCESS_PROVIDERS.GCORE]: AccessConfigFieldsProviderGcore,
  [ACCESS_PROVIDERS.GLOBALSIGNATLAS]: AccessConfigFieldsProviderGlobalSignAtlas,
  [ACCESS_PROVIDERS.GNAME]: AccessConfigFieldsProviderGname,
  [ACCESS_PROVIDERS.GODADDY]: AccessConfigFieldsProviderGoDaddy,
  [ACCESS_PROVIDERS.GOOGLECLOUD]: AccessConfigFieldsProviderGoogleCloud,
  [ACCESS_PROVIDERS.GOOGLETRUSTSERVICES]: AccessConfigFieldsProviderGoogleTrustServices,
  [ACCESS_PROVIDERS.HETZNER]: AccessConfigFieldsProviderHetzner,
  [ACCESS_PROVIDERS.HOSTINGDE]: AccessConfigFieldsProviderHostingde,
  [ACCESS_PROVIDERS.HOSTINGER]: AccessConfigFieldsProviderHostinger,
  [ACCESS_PROVIDERS.HUAWEICLOUD]: AccessConfigFieldsProviderHuaweiCloud,
  [ACCESS_PROVIDERS.IONOS]: AccessConfigFieldsProviderIONOS,
  [ACCESS_PROVIDERS.JDCLOUD]: AccessConfigFieldsProviderJDCloud,
  [ACCESS_PROVIDERS.INFOMANIAK]: AccessConfigFieldsProviderInfomaniak,
  [ACCESS_PROVIDERS.LINODE]: AccessConfigFieldsProviderLinode,
  [ACCESS_PROVIDERS.LITESSL]: AccessConfigFieldsProviderLiteSSL,
  [ACCESS_PROVIDERS.LUADNS]: AccessConfigFieldsProviderLuaDNS,
  [ACCESS_PROVIDERS.NAMECHEAP]: AccessConfigFieldsProviderNamecheap,
  [ACCESS_PROVIDERS.NAMEDOTCOM]: AccessConfigFieldsProviderNameDotCom,
  [ACCESS_PROVIDERS.NAMESILO]: AccessConfigFieldsProviderNameSilo,
  [ACCESS_PROVIDERS.NETCUP]: AccessConfigFieldsProviderNetcup,
  [ACCESS_PROVIDERS.NETLIFY]: AccessConfigFieldsProviderNetlify,
  [ACCESS_PROVIDERS.NS1]: AccessConfigFieldsProviderNS1,
  [ACCESS_PROVIDERS.ORACLECLOUD]: AccessConfigFieldsProviderOracleCloud,
  [ACCESS_PROVIDERS.OVHCLOUD]: AccessConfigFieldsProviderOVHcloud,
  [ACCESS_PROVIDERS.PORKBUN]: AccessConfigFieldsProviderPorkbun,
  [ACCESS_PROVIDERS.POWERDNS]: AccessConfigFieldsProviderPowerDNS,
  [ACCESS_PROVIDERS.QINGCLOUD]: AccessConfigFieldsProviderQingCloud,
  [ACCESS_PROVIDERS.RAINYUN]: AccessConfigFieldsProviderRainYun,
  [ACCESS_PROVIDERS.REGRU]: AccessConfigFieldsProviderRegru,
  [ACCESS_PROVIDERS.RFC2136]: AccessConfigFieldsProviderRFC2136,
  [ACCESS_PROVIDERS.RUCENTER]: AccessConfigFieldsProviderRuCenter,
  [ACCESS_PROVIDERS.S3]: AccessConfigFieldsProviderS3,
  [ACCESS_PROVIDERS.SECTIGO]: AccessConfigFieldsProviderSectigo,
  [ACCESS_PROVIDERS.SIMPLYCOM]: AccessConfigFieldsProviderSimplyCom,
  [ACCESS_PROVIDERS.SPACESHIP]: AccessConfigFieldsProviderSpaceship,
  [ACCESS_PROVIDERS.SSLCOM]: AccessConfigFieldsProviderSSLCom,
  [ACCESS_PROVIDERS.SSH]: AccessConfigFieldsProviderSSH,
  [ACCESS_PROVIDERS.TECHNITIUMDNS]: AccessConfigFieldsProviderTechnitiumDNS,
  [ACCESS_PROVIDERS.TENCENTCLOUD]: AccessConfigFieldsProviderTencentCloud,
  [ACCESS_PROVIDERS.TODAYNIC]: AccessConfigFieldsProviderTodayNIC,
  [ACCESS_PROVIDERS.UCLOUD]: AccessConfigFieldsProviderUCloud,
  [ACCESS_PROVIDERS.VERCEL]: AccessConfigFieldsProviderVercel,
  [ACCESS_PROVIDERS.VOLCENGINE]: AccessConfigFieldsProviderVolcEngine,
  [ACCESS_PROVIDERS.VULTR]: AccessConfigFieldsProviderVultr,
  [ACCESS_PROVIDERS.WEBHOOK]: AccessConfigFieldsProviderWebhook,
  [ACCESS_PROVIDERS.WESTCN]: AccessConfigFieldsProviderWestcn,
  [ACCESS_PROVIDERS.XINNET]: AccessConfigFieldsProviderXinnet,
  [ACCESS_PROVIDERS.YANDEXCLOUD]: AccessConfigFieldsProviderYandexCloud,
  [ACCESS_PROVIDERS.ZEROSSL]: AccessConfigFieldsProviderZeroSSL,
};

const useComponent = (provider: string | undefined, { initProps }: { initProps?: (provider: string) => object }) => {
  const [result, setResult] = useState<{
    provider: string;
    catalogVersion: number;
    status: "ready" | "unavailable";
    envelope?: ProviderSchemaEnvelope;
  }>();
  const catalogVersion = usePluginCatalogStore((s) => s.version);

  useEffect(() => {
    if (!provider) return;

    let cancelled = false;
    void getProviderSchema(provider, "access").then((envelope) => {
      if (cancelled) return;
      setResult((previous) => ({
        provider,
        catalogVersion,
        status: envelope ? "ready" : "unavailable",
        envelope: envelope ?? (previous?.provider === provider ? previous.envelope : undefined),
      }));
    });
    return () => {
      cancelled = true;
    };
  }, [provider, catalogVersion]);

  const matched = result?.provider === provider ? result : undefined;
  const envelope = matched?.envelope;
  const schemaStatus = matched?.catalogVersion === catalogVersion ? matched.status : "loading";

  if (envelope) return { schemaStatus, fields: <SchemaConfigFields key={provider} envelope={envelope} /> };
  const Component = provider ? providerComponentMap[provider as AccessProviderType] : undefined;
  return { schemaStatus, fields: Component && provider ? <Component {...initProps?.(provider)} /> : null };
};

const _default = {
  useComponent,
};

export default _default;
