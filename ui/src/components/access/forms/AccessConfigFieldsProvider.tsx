import { useEffect, useState } from "react";

import { type ProviderSchemaEnvelope, getProviderSchema } from "@/api/providerschema";
import SchemaConfigFields from "@/components/workflow/designer/forms/SchemaConfigFields";
import { ACCESS_PROVIDERS, type AccessProviderType } from "@/domain/provider";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";

import AccessConfigFieldsProviderACMECA from "./AccessConfigFieldsProviderACMECA";
import AccessConfigFieldsProviderActalisSSL from "./AccessConfigFieldsProviderActalisSSL";
import AccessConfigFieldsProviderAWS from "./AccessConfigFieldsProviderAWS";
import AccessConfigFieldsProviderAzure from "./AccessConfigFieldsProviderAzure";
import AccessConfigFieldsProviderBaiduCloud from "./AccessConfigFieldsProviderBaiduCloud";
import AccessConfigFieldsProviderBunny from "./AccessConfigFieldsProviderBunny";
import AccessConfigFieldsProviderBytePlus from "./AccessConfigFieldsProviderBytePlus";
import AccessConfigFieldsProviderCloudflare from "./AccessConfigFieldsProviderCloudflare";
import AccessConfigFieldsProviderCMCCCloud from "./AccessConfigFieldsProviderCMCCCloud";
import AccessConfigFieldsProviderCPanel from "./AccessConfigFieldsProviderCPanel";
import AccessConfigFieldsProviderCTCCCloud from "./AccessConfigFieldsProviderCTCCCloud";
import AccessConfigFieldsProviderDigiCert from "./AccessConfigFieldsProviderDigiCert";
import AccessConfigFieldsProviderDigitalOcean from "./AccessConfigFieldsProviderDigitalOcean";
import AccessConfigFieldsProviderEmail from "./AccessConfigFieldsProviderEmail";
import AccessConfigFieldsProviderFTP from "./AccessConfigFieldsProviderFTP";
import AccessConfigFieldsProviderGcore from "./AccessConfigFieldsProviderGcore";
import AccessConfigFieldsProviderGlobalSignAtlas from "./AccessConfigFieldsProviderGlobalSignAtlas";
import AccessConfigFieldsProviderGoogleCloud from "./AccessConfigFieldsProviderGoogleCloud";
import AccessConfigFieldsProviderGoogleTrustServices from "./AccessConfigFieldsProviderGoogleTrustServices";
import AccessConfigFieldsProviderHuaweiCloud from "./AccessConfigFieldsProviderHuaweiCloud";
import AccessConfigFieldsProviderJDCloud from "./AccessConfigFieldsProviderJDCloud";
import AccessConfigFieldsProviderLinode from "./AccessConfigFieldsProviderLinode";
import AccessConfigFieldsProviderLiteSSL from "./AccessConfigFieldsProviderLiteSSL";
import AccessConfigFieldsProviderNetlify from "./AccessConfigFieldsProviderNetlify";
import AccessConfigFieldsProviderOracleCloud from "./AccessConfigFieldsProviderOracleCloud";
import AccessConfigFieldsProviderPowerDNS from "./AccessConfigFieldsProviderPowerDNS";
import AccessConfigFieldsProviderQingCloud from "./AccessConfigFieldsProviderQingCloud";
import AccessConfigFieldsProviderRainYun from "./AccessConfigFieldsProviderRainYun";
import AccessConfigFieldsProviderRFC2136 from "./AccessConfigFieldsProviderRFC2136";
import AccessConfigFieldsProviderS3 from "./AccessConfigFieldsProviderS3";
import AccessConfigFieldsProviderSectigo from "./AccessConfigFieldsProviderSectigo";
import AccessConfigFieldsProviderSSH from "./AccessConfigFieldsProviderSSH";
import AccessConfigFieldsProviderSSLCom from "./AccessConfigFieldsProviderSSLCom";
import AccessConfigFieldsProviderTencentCloud from "./AccessConfigFieldsProviderTencentCloud";
import AccessConfigFieldsProviderUCloud from "./AccessConfigFieldsProviderUCloud";
import AccessConfigFieldsProviderVercel from "./AccessConfigFieldsProviderVercel";
import AccessConfigFieldsProviderVolcEngine from "./AccessConfigFieldsProviderVolcEngine";
import AccessConfigFieldsProviderWebhook from "./AccessConfigFieldsProviderWebhook";
import AccessConfigFieldsProviderYandexCloud from "./AccessConfigFieldsProviderYandexCloud";
import AccessConfigFieldsProviderZeroSSL from "./AccessConfigFieldsProviderZeroSSL";

const providerComponentMap: Partial<Record<AccessProviderType, React.ComponentType<any>>> = {
  /*
    注意：如果追加新的子组件，请保持以 ASCII 排序。
    NOTICE: If you add new child component, please keep ASCII order.
    */
  [ACCESS_PROVIDERS.ACMECA]: AccessConfigFieldsProviderACMECA,
  [ACCESS_PROVIDERS.ACTALISSSL]: AccessConfigFieldsProviderActalisSSL,
  [ACCESS_PROVIDERS.AWS]: AccessConfigFieldsProviderAWS,
  [ACCESS_PROVIDERS.AZURE]: AccessConfigFieldsProviderAzure,
  [ACCESS_PROVIDERS.BAIDUCLOUD]: AccessConfigFieldsProviderBaiduCloud,
  [ACCESS_PROVIDERS.BUNNY]: AccessConfigFieldsProviderBunny,
  [ACCESS_PROVIDERS.BYTEPLUS]: AccessConfigFieldsProviderBytePlus,
  [ACCESS_PROVIDERS.CLOUDFLARE]: AccessConfigFieldsProviderCloudflare,
  [ACCESS_PROVIDERS.CMCCCLOUD]: AccessConfigFieldsProviderCMCCCloud,
  [ACCESS_PROVIDERS.CPANEL]: AccessConfigFieldsProviderCPanel,
  [ACCESS_PROVIDERS.CTCCCLOUD]: AccessConfigFieldsProviderCTCCCloud,
  [ACCESS_PROVIDERS.DIGICERT]: AccessConfigFieldsProviderDigiCert,
  [ACCESS_PROVIDERS.DIGITALOCEAN]: AccessConfigFieldsProviderDigitalOcean,
  [ACCESS_PROVIDERS.EMAIL]: AccessConfigFieldsProviderEmail,
  [ACCESS_PROVIDERS.FTP]: AccessConfigFieldsProviderFTP,
  [ACCESS_PROVIDERS.GCORE]: AccessConfigFieldsProviderGcore,
  [ACCESS_PROVIDERS.GLOBALSIGNATLAS]: AccessConfigFieldsProviderGlobalSignAtlas,
  [ACCESS_PROVIDERS.GOOGLECLOUD]: AccessConfigFieldsProviderGoogleCloud,
  [ACCESS_PROVIDERS.GOOGLETRUSTSERVICES]: AccessConfigFieldsProviderGoogleTrustServices,
  [ACCESS_PROVIDERS.HUAWEICLOUD]: AccessConfigFieldsProviderHuaweiCloud,
  [ACCESS_PROVIDERS.JDCLOUD]: AccessConfigFieldsProviderJDCloud,
  [ACCESS_PROVIDERS.LINODE]: AccessConfigFieldsProviderLinode,
  [ACCESS_PROVIDERS.LITESSL]: AccessConfigFieldsProviderLiteSSL,
  [ACCESS_PROVIDERS.NETLIFY]: AccessConfigFieldsProviderNetlify,
  [ACCESS_PROVIDERS.ORACLECLOUD]: AccessConfigFieldsProviderOracleCloud,
  [ACCESS_PROVIDERS.POWERDNS]: AccessConfigFieldsProviderPowerDNS,
  [ACCESS_PROVIDERS.QINGCLOUD]: AccessConfigFieldsProviderQingCloud,
  [ACCESS_PROVIDERS.RAINYUN]: AccessConfigFieldsProviderRainYun,
  [ACCESS_PROVIDERS.RFC2136]: AccessConfigFieldsProviderRFC2136,
  [ACCESS_PROVIDERS.S3]: AccessConfigFieldsProviderS3,
  [ACCESS_PROVIDERS.SECTIGO]: AccessConfigFieldsProviderSectigo,
  [ACCESS_PROVIDERS.SSLCOM]: AccessConfigFieldsProviderSSLCom,
  [ACCESS_PROVIDERS.SSH]: AccessConfigFieldsProviderSSH,
  [ACCESS_PROVIDERS.TENCENTCLOUD]: AccessConfigFieldsProviderTencentCloud,
  [ACCESS_PROVIDERS.UCLOUD]: AccessConfigFieldsProviderUCloud,
  [ACCESS_PROVIDERS.VERCEL]: AccessConfigFieldsProviderVercel,
  [ACCESS_PROVIDERS.VOLCENGINE]: AccessConfigFieldsProviderVolcEngine,
  [ACCESS_PROVIDERS.WEBHOOK]: AccessConfigFieldsProviderWebhook,
  [ACCESS_PROVIDERS.YANDEXCLOUD]: AccessConfigFieldsProviderYandexCloud,
  [ACCESS_PROVIDERS.ZEROSSL]: AccessConfigFieldsProviderZeroSSL,
};

const useComponent = (
  provider: string | undefined,
  { initProps, preferStaticFields }: { initProps?: (provider: string) => object; preferStaticFields?: boolean }
) => {
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
  const envelope = preferStaticFields ? undefined : matched?.envelope;
  const schemaStatus = matched?.catalogVersion === catalogVersion ? matched.status : "loading";

  if (envelope) return { schemaStatus, fields: <SchemaConfigFields key={provider} envelope={envelope} /> };
  const Component = provider ? providerComponentMap[provider as AccessProviderType] : undefined;
  return { schemaStatus, fields: Component && provider ? <Component {...initProps?.(provider)} /> : null };
};

const _default = {
  useComponent,
};

export default _default;
