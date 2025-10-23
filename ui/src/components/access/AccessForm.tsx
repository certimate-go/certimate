import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import AccessProviderSelect from "@/components/provider/AccessProviderSelect";
import { type AccessModel } from "@/domain/access";
import { ACCESS_PROVIDERS, ACCESS_USAGES } from "@/domain/provider";
import { useAntdForm } from "@/hooks";

import { FormNestedFieldsContextProvider } from "./forms/_context";
import { useProviderFilterByUsage } from "./forms/_hooks";
import AccessConfigFieldsProvider1Panel from "./forms/AccessConfigFieldsProvider1Panel";
import AccessConfigFieldsProviderACMECA from "./forms/AccessConfigFieldsProviderACMECA";
import AccessConfigFieldsProviderACMEDNS from "./forms/AccessConfigFieldsProviderACMEDNS";
import AccessConfigFieldsProviderACMEHttpReq from "./forms/AccessConfigFieldsProviderACMEHttpReq";
import AccessConfigFieldsProviderActalisSSL from "./forms/AccessConfigFieldsProviderActalisSSL";
import AccessConfigFieldsProviderAliyun from "./forms/AccessConfigFieldsProviderAliyun";
import AccessConfigFieldsProviderAPISIX from "./forms/AccessConfigFieldsProviderAPISIX";
import AccessConfigFieldsProviderAWS from "./forms/AccessConfigFieldsProviderAWS";
import AccessConfigFieldsProviderAzure from "./forms/AccessConfigFieldsProviderAzure";
import AccessConfigFieldsProviderBaiduCloud from "./forms/AccessConfigFieldsProviderBaiduCloud";
import AccessConfigFieldsProviderBaishan from "./forms/AccessConfigFieldsProviderBaishan";
import AccessConfigFieldsProviderBaotaPanel from "./forms/AccessConfigFieldsProviderBaotaPanel";
import AccessConfigFieldsProviderBaotaPanelGo from "./forms/AccessConfigFieldsProviderBaotaPanelGo";
import AccessConfigFieldsProviderBaotaWAF from "./forms/AccessConfigFieldsProviderBaotaWAF";
import AccessConfigFieldsProviderBookMyName from "./forms/AccessConfigFieldsProviderBookMyName";
import AccessConfigFieldsProviderBunny from "./forms/AccessConfigFieldsProviderBunny";
import AccessConfigFieldsProviderBytePlus from "./forms/AccessConfigFieldsProviderBytePlus";
import AccessConfigFieldsProviderCacheFly from "./forms/AccessConfigFieldsProviderCacheFly";
import AccessConfigFieldsProviderCdnfly from "./forms/AccessConfigFieldsProviderCdnfly";
import AccessConfigFieldsProviderCloudflare from "./forms/AccessConfigFieldsProviderCloudflare";
import AccessConfigFieldsProviderClouDNS from "./forms/AccessConfigFieldsProviderClouDNS";
import AccessConfigFieldsProviderCMCCCloud from "./forms/AccessConfigFieldsProviderCMCCCloud";
import AccessConfigFieldsProviderConstellix from "./forms/AccessConfigFieldsProviderConstellix";
import AccessConfigFieldsProviderCTCCCloud from "./forms/AccessConfigFieldsProviderCTCCCloud";
import AccessConfigFieldsProviderDeSEC from "./forms/AccessConfigFieldsProviderDeSEC";
import AccessConfigFieldsProviderDigitalOcean from "./forms/AccessConfigFieldsProviderDigitalOcean";
import AccessConfigFieldsProviderDingTalkBot from "./forms/AccessConfigFieldsProviderDingTalkBot";
import AccessConfigFieldsProviderDiscordBot from "./forms/AccessConfigFieldsProviderDiscordBot";
import AccessConfigFieldsProviderDNSLA from "./forms/AccessConfigFieldsProviderDNSLA";
import AccessConfigFieldsProviderDogeCloud from "./forms/AccessConfigFieldsProviderDogeCloud";
import AccessConfigFieldsProviderDuckDNS from "./forms/AccessConfigFieldsProviderDuckDNS";
import AccessConfigFieldsProviderDynv6 from "./forms/AccessConfigFieldsProviderDynv6";
import AccessConfigFieldsProviderEmail from "./forms/AccessConfigFieldsProviderEmail";
import AccessConfigFieldsProviderFlexCDN from "./forms/AccessConfigFieldsProviderFlexCDN";
import AccessConfigFieldsProviderGandinet from "./forms/AccessConfigFieldsProviderGandinet";
import AccessConfigFieldsProviderGcore from "./forms/AccessConfigFieldsProviderGcore";
import AccessConfigFieldsProviderGlobalSignAtlas from "./forms/AccessConfigFieldsProviderGlobalSignAtlas";
import AccessConfigFieldsProviderGname from "./forms/AccessConfigFieldsProviderGname";
import AccessConfigFieldsProviderGoDaddy from "./forms/AccessConfigFieldsProviderGoDaddy";
import AccessConfigFieldsProviderGoEdge from "./forms/AccessConfigFieldsProviderGoEdge";
import AccessConfigFieldsProviderGoogleTrustServices from "./forms/AccessConfigFieldsProviderGoogleTrustServices";
import AccessConfigFieldsProviderHetzner from "./forms/AccessConfigFieldsProviderHetzner";
import AccessConfigFieldsProviderHostinger from "./forms/AccessConfigFieldsProviderHostinger";
import AccessConfigFieldsProviderHuaweiCloud from "./forms/AccessConfigFieldsProviderHuaweiCloud";
import AccessConfigFieldsProviderIONOS from "./forms/AccessConfigFieldsProviderIONOS";
import AccessConfigFieldsProviderJDCloud from "./forms/AccessConfigFieldsProviderJDCloud";
import AccessConfigFieldsProviderKong from "./forms/AccessConfigFieldsProviderKong";
import AccessConfigFieldsProviderKubernetes from "./forms/AccessConfigFieldsProviderKubernetes";
import AccessConfigFieldsProviderLarkBot from "./forms/AccessConfigFieldsProviderLarkBot";
import AccessConfigFieldsProviderLeCDN from "./forms/AccessConfigFieldsProviderLeCDN";
import AccessConfigFieldsProviderLinode from "./forms/AccessConfigFieldsProviderLinode";
import AccessConfigFieldsProviderMattermost from "./forms/AccessConfigFieldsProviderMattermost";
import AccessConfigFieldsProviderNamecheap from "./forms/AccessConfigFieldsProviderNamecheap";
import AccessConfigFieldsProviderNameDotCom from "./forms/AccessConfigFieldsProviderNameDotCom";
import AccessConfigFieldsProviderNameSilo from "./forms/AccessConfigFieldsProviderNameSilo";
import AccessConfigFieldsProviderNetcup from "./forms/AccessConfigFieldsProviderNetcup";
import AccessConfigFieldsProviderNetlify from "./forms/AccessConfigFieldsProviderNetlify";
import AccessConfigFieldsProviderNS1 from "./forms/AccessConfigFieldsProviderNS1";
import AccessConfigFieldsProviderPorkbun from "./forms/AccessConfigFieldsProviderPorkbun";
import AccessConfigFieldsProviderPowerDNS from "./forms/AccessConfigFieldsProviderPowerDNS";
import AccessConfigFieldsProviderProxmoxVE from "./forms/AccessConfigFieldsProviderProxmoxVE";
import AccessConfigFieldsProviderQiniu from "./forms/AccessConfigFieldsProviderQiniu";
import AccessConfigFieldsProviderRainYun from "./forms/AccessConfigFieldsProviderRainYun";
import AccessConfigFieldsProviderRatPanel from "./forms/AccessConfigFieldsProviderRatPanel";
import AccessConfigFieldsProviderRFC2136 from "./forms/AccessConfigFieldsProviderRFC2136";
import AccessConfigFieldsProviderSafeLine from "./forms/AccessConfigFieldsProviderSafeLine";
import AccessConfigFieldsProviderSectigo from "./forms/AccessConfigFieldsProviderSectigo";
import AccessConfigFieldsProviderSlackBot from "./forms/AccessConfigFieldsProviderSlackBot";
import AccessConfigFieldsProviderSpaceship from "./forms/AccessConfigFieldsProviderSpaceship";
import AccessConfigFieldsProviderSSH from "./forms/AccessConfigFieldsProviderSSH";
import AccessConfigFieldsProviderSSLCom from "./forms/AccessConfigFieldsProviderSSLCom";
import AccessConfigFieldsProviderTechnitiumDNS from "./forms/AccessConfigFieldsProviderTechnitiumDNS";
import AccessConfigFieldsProviderTelegramBot from "./forms/AccessConfigFieldsProviderTelegramBot";
import AccessConfigFieldsProviderTencentCloud from "./forms/AccessConfigFieldsProviderTencentCloud";
import AccessConfigFieldsProviderUCloud from "./forms/AccessConfigFieldsProviderUCloud";
import AccessConfigFieldsProviderUniCloud from "./forms/AccessConfigFieldsProviderUniCloud";
import AccessConfigFieldsProviderUpyun from "./forms/AccessConfigFieldsProviderUpyun";
import AccessConfigFieldsProviderVercel from "./forms/AccessConfigFieldsProviderVercel";
import AccessConfigFieldsProviderVolcEngine from "./forms/AccessConfigFieldsProviderVolcEngine";
import AccessConfigFieldsProviderVultr from "./forms/AccessConfigFieldsProviderVultr";
import AccessConfigFieldsProviderWangsu from "./forms/AccessConfigFieldsProviderWangsu";
import AccessConfigFieldsProviderWebhook from "./forms/AccessConfigFieldsProviderWebhook";
import AccessConfigFieldsProviderWeComBot from "./forms/AccessConfigFieldsProviderWeComBot";
import AccessConfigFieldsProviderWestcn from "./forms/AccessConfigFieldsProviderWestcn";
import AccessConfigFieldsProviderZeroSSL from "./forms/AccessConfigFieldsProviderZeroSSL";

export type AccessFormModes = "create" | "modify";
export type AccessFormUsages = "dns" | "hosting" | "dns-hosting" | "ca" | "notification";

export interface AccessFormProps {
  className?: string;
  style?: React.CSSProperties;
  disabled?: boolean;
  initialValues?: Nullish<MaybeModelRecord<AccessModel>>;
  form: FormInstance;
  mode: AccessFormModes;
  usage?: AccessFormUsages;
  onFormValuesChange?: (changedValues: Nullish<MaybeModelRecord<AccessModel>>, values: Nullish<MaybeModelRecord<AccessModel>>) => void;
}

const AccessForm = ({ className, style, disabled, initialValues, mode, usage, onFormValuesChange, ...props }: AccessFormProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    name: z
      .string(t("access.form.name.placeholder"))
      .min(1, t("access.form.name.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    provider: z.enum(ACCESS_PROVIDERS, t("access.form.provider.placeholder")),
    config: z.any(),
    reserve: z.string().nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "accessForm",
    initialValues: initialValues,
  });

  const providerFilter = useProviderFilterByUsage(usage);

  const fieldProvider = Form.useWatch<string>("provider", { form: formInst, preserve: true });

  const nestedProviderConfigFields = useMemo(() => {
    /*
        注意：如果追加新的子组件，请保持以 ASCII 排序。
        NOTICE: If you add new child component, please keep ASCII order.
       */
    switch (fieldProvider) {
      case ACCESS_PROVIDERS["1PANEL"]: {
        return <AccessConfigFieldsProvider1Panel />;
      }
      case ACCESS_PROVIDERS.ACMECA: {
        return <AccessConfigFieldsProviderACMECA />;
      }
      case ACCESS_PROVIDERS.ACMEDNS: {
        return <AccessConfigFieldsProviderACMEDNS />;
      }
      case ACCESS_PROVIDERS.ACMEHTTPREQ: {
        return <AccessConfigFieldsProviderACMEHttpReq />;
      }
      case ACCESS_PROVIDERS.ACTALISSSL: {
        return <AccessConfigFieldsProviderActalisSSL />;
      }
      case ACCESS_PROVIDERS.ALIYUN: {
        return <AccessConfigFieldsProviderAliyun />;
      }
      case ACCESS_PROVIDERS.APISIX: {
        return <AccessConfigFieldsProviderAPISIX />;
      }
      case ACCESS_PROVIDERS.AWS: {
        return <AccessConfigFieldsProviderAWS />;
      }
      case ACCESS_PROVIDERS.AZURE: {
        return <AccessConfigFieldsProviderAzure />;
      }
      case ACCESS_PROVIDERS.BAIDUCLOUD: {
        return <AccessConfigFieldsProviderBaiduCloud />;
      }
      case ACCESS_PROVIDERS.BAISHAN: {
        return <AccessConfigFieldsProviderBaishan />;
      }
      case ACCESS_PROVIDERS.BAOTAPANEL: {
        return <AccessConfigFieldsProviderBaotaPanel />;
      }
      case ACCESS_PROVIDERS.BAOTAPANELGO: {
        return <AccessConfigFieldsProviderBaotaPanelGo />;
      }
      case ACCESS_PROVIDERS.BAOTAWAF: {
        return <AccessConfigFieldsProviderBaotaWAF />;
      }
      case ACCESS_PROVIDERS.BOOKMYNAME: {
        return <AccessConfigFieldsProviderBookMyName />;
      }
      case ACCESS_PROVIDERS.BUNNY: {
        return <AccessConfigFieldsProviderBunny />;
      }
      case ACCESS_PROVIDERS.BYTEPLUS: {
        return <AccessConfigFieldsProviderBytePlus />;
      }
      case ACCESS_PROVIDERS.CACHEFLY: {
        return <AccessConfigFieldsProviderCacheFly />;
      }
      case ACCESS_PROVIDERS.CDNFLY: {
        return <AccessConfigFieldsProviderCdnfly />;
      }
      case ACCESS_PROVIDERS.CLOUDFLARE: {
        return <AccessConfigFieldsProviderCloudflare />;
      }
      case ACCESS_PROVIDERS.CLOUDNS: {
        return <AccessConfigFieldsProviderClouDNS />;
      }
      case ACCESS_PROVIDERS.CMCCCLOUD: {
        return <AccessConfigFieldsProviderCMCCCloud />;
      }
      case ACCESS_PROVIDERS.CONSTELLIX: {
        return <AccessConfigFieldsProviderConstellix />;
      }
      case ACCESS_PROVIDERS.CTCCCLOUD: {
        return <AccessConfigFieldsProviderCTCCCloud />;
      }
      case ACCESS_PROVIDERS.DESEC: {
        return <AccessConfigFieldsProviderDeSEC />;
      }
      case ACCESS_PROVIDERS.DIGITALOCEAN: {
        return <AccessConfigFieldsProviderDigitalOcean />;
      }
      case ACCESS_PROVIDERS.DINGTALKBOT: {
        return <AccessConfigFieldsProviderDingTalkBot />;
      }
      case ACCESS_PROVIDERS.DISCORDBOT: {
        return <AccessConfigFieldsProviderDiscordBot />;
      }
      case ACCESS_PROVIDERS.DNSLA: {
        return <AccessConfigFieldsProviderDNSLA />;
      }
      case ACCESS_PROVIDERS.DOGECLOUD: {
        return <AccessConfigFieldsProviderDogeCloud />;
      }
      case ACCESS_PROVIDERS.DUCKDNS: {
        return <AccessConfigFieldsProviderDuckDNS />;
      }
      case ACCESS_PROVIDERS.DYNV6: {
        return <AccessConfigFieldsProviderDynv6 />;
      }
      case ACCESS_PROVIDERS.EMAIL: {
        return <AccessConfigFieldsProviderEmail />;
      }
      case ACCESS_PROVIDERS.FLEXCDN: {
        return <AccessConfigFieldsProviderFlexCDN />;
      }
      case ACCESS_PROVIDERS.GANDINET: {
        return <AccessConfigFieldsProviderGandinet />;
      }
      case ACCESS_PROVIDERS.GCORE: {
        return <AccessConfigFieldsProviderGcore />;
      }
      case ACCESS_PROVIDERS.GNAME: {
        return <AccessConfigFieldsProviderGname />;
      }
      case ACCESS_PROVIDERS.GODADDY: {
        return <AccessConfigFieldsProviderGoDaddy />;
      }
      case ACCESS_PROVIDERS.GOEDGE: {
        return <AccessConfigFieldsProviderGoEdge />;
      }
      case ACCESS_PROVIDERS.GLOBALSIGNATLAS: {
        return <AccessConfigFieldsProviderGlobalSignAtlas />;
      }
      case ACCESS_PROVIDERS.GOOGLETRUSTSERVICES: {
        return <AccessConfigFieldsProviderGoogleTrustServices />;
      }
      case ACCESS_PROVIDERS.HETZNER: {
        return <AccessConfigFieldsProviderHetzner />;
      }
      case ACCESS_PROVIDERS.HOSTINGER: {
        return <AccessConfigFieldsProviderHostinger />;
      }
      case ACCESS_PROVIDERS.HUAWEICLOUD: {
        return <AccessConfigFieldsProviderHuaweiCloud />;
      }
      case ACCESS_PROVIDERS.IONOS: {
        return <AccessConfigFieldsProviderIONOS />;
      }
      case ACCESS_PROVIDERS.JDCLOUD: {
        return <AccessConfigFieldsProviderJDCloud />;
      }
      case ACCESS_PROVIDERS.KONG: {
        return <AccessConfigFieldsProviderKong />;
      }
      case ACCESS_PROVIDERS.KUBERNETES: {
        return <AccessConfigFieldsProviderKubernetes />;
      }
      case ACCESS_PROVIDERS.LARKBOT: {
        return <AccessConfigFieldsProviderLarkBot />;
      }
      case ACCESS_PROVIDERS.LECDN: {
        return <AccessConfigFieldsProviderLeCDN />;
      }
      case ACCESS_PROVIDERS.LINODE: {
        return <AccessConfigFieldsProviderLinode />;
      }
      case ACCESS_PROVIDERS.MATTERMOST: {
        return <AccessConfigFieldsProviderMattermost />;
      }
      case ACCESS_PROVIDERS.NAMECHEAP: {
        return <AccessConfigFieldsProviderNamecheap />;
      }
      case ACCESS_PROVIDERS.NAMEDOTCOM: {
        return <AccessConfigFieldsProviderNameDotCom />;
      }
      case ACCESS_PROVIDERS.NAMESILO: {
        return <AccessConfigFieldsProviderNameSilo />;
      }
      case ACCESS_PROVIDERS.NETCUP: {
        return <AccessConfigFieldsProviderNetcup />;
      }
      case ACCESS_PROVIDERS.NETLIFY: {
        return <AccessConfigFieldsProviderNetlify />;
      }
      case ACCESS_PROVIDERS.NS1: {
        return <AccessConfigFieldsProviderNS1 />;
      }
      case ACCESS_PROVIDERS.PORKBUN: {
        return <AccessConfigFieldsProviderPorkbun />;
      }
      case ACCESS_PROVIDERS.POWERDNS: {
        return <AccessConfigFieldsProviderPowerDNS />;
      }
      case ACCESS_PROVIDERS.PROXMOXVE: {
        return <AccessConfigFieldsProviderProxmoxVE />;
      }
      case ACCESS_PROVIDERS.QINIU: {
        return <AccessConfigFieldsProviderQiniu />;
      }
      case ACCESS_PROVIDERS.RAINYUN: {
        return <AccessConfigFieldsProviderRainYun />;
      }
      case ACCESS_PROVIDERS.RATPANEL: {
        return <AccessConfigFieldsProviderRatPanel />;
      }
      case ACCESS_PROVIDERS.RFC2136: {
        return <AccessConfigFieldsProviderRFC2136 />;
      }
      case ACCESS_PROVIDERS.SAFELINE: {
        return <AccessConfigFieldsProviderSafeLine />;
      }
      case ACCESS_PROVIDERS.SECTIGO: {
        return <AccessConfigFieldsProviderSectigo />;
      }
      case ACCESS_PROVIDERS.SLACKBOT: {
        return <AccessConfigFieldsProviderSlackBot />;
      }
      case ACCESS_PROVIDERS.SPACESHIP: {
        return <AccessConfigFieldsProviderSpaceship />;
      }
      case ACCESS_PROVIDERS.SSLCOM: {
        return <AccessConfigFieldsProviderSSLCom />;
      }
      case ACCESS_PROVIDERS.SSH: {
        return <AccessConfigFieldsProviderSSH disabled={disabled} />;
      }
      case ACCESS_PROVIDERS.TECHNITIUMDNS: {
        return <AccessConfigFieldsProviderTechnitiumDNS />;
      }
      case ACCESS_PROVIDERS.TELEGRAMBOT: {
        return <AccessConfigFieldsProviderTelegramBot />;
      }
      case ACCESS_PROVIDERS.TENCENTCLOUD: {
        return <AccessConfigFieldsProviderTencentCloud />;
      }
      case ACCESS_PROVIDERS.UCLOUD: {
        return <AccessConfigFieldsProviderUCloud />;
      }
      case ACCESS_PROVIDERS.UNICLOUD: {
        return <AccessConfigFieldsProviderUniCloud />;
      }
      case ACCESS_PROVIDERS.UPYUN: {
        return <AccessConfigFieldsProviderUpyun />;
      }
      case ACCESS_PROVIDERS.VERCEL: {
        return <AccessConfigFieldsProviderVercel />;
      }
      case ACCESS_PROVIDERS.VOLCENGINE: {
        return <AccessConfigFieldsProviderVolcEngine />;
      }
      case ACCESS_PROVIDERS.VULTR: {
        return <AccessConfigFieldsProviderVultr />;
      }
      case ACCESS_PROVIDERS.WANGSU: {
        return <AccessConfigFieldsProviderWangsu />;
      }
      case ACCESS_PROVIDERS.WEBHOOK: {
        const webhookUsage = usage === "notification" ? "notification" : usage === "hosting" || usage === "dns-hosting" ? "deployment" : "none";
        return <AccessConfigFieldsProviderWebhook usage={webhookUsage} />;
      }
      case ACCESS_PROVIDERS.WECOMBOT: {
        return <AccessConfigFieldsProviderWeComBot />;
      }
      case ACCESS_PROVIDERS.WESTCN: {
        return <AccessConfigFieldsProviderWestcn />;
      }
      case ACCESS_PROVIDERS.ZEROSSL: {
        return <AccessConfigFieldsProviderZeroSSL />;
      }
    }
  }, [disabled, usage, fieldProvider]);

  return (
    <Form
      className={className}
      style={style}
      {...formProps}
      clearOnDestroy={true}
      disabled={disabled}
      form={formInst}
      layout="vertical"
      preserve={false}
      scrollToFirstError
      onValuesChange={onFormValuesChange}
    >
      <Form.Item name="name" label={t("access.form.name.label")} rules={[formRule]}>
        <Input placeholder={t("access.form.name.placeholder")} />
      </Form.Item>

      <Form.Item
        name="provider"
        label={t("access.form.provider.label")}
        extra={usage === "dns-hosting" ? <span dangerouslySetInnerHTML={{ __html: t("access.form.provider.help") }}></span> : null}
        rules={[formRule]}
      >
        <AccessProviderSelect
          disabled={mode !== "create"}
          placeholder={t("access.form.provider.placeholder")}
          showOptionTags={
            usage == null || (usage === "dns-hosting" ? { ["builtin"]: true, [ACCESS_USAGES.DNS]: true, [ACCESS_USAGES.HOSTING]: true } : { ["builtin"]: true })
          }
          showSearch={!disabled}
          onFilter={providerFilter}
        />
      </Form.Item>

      <FormNestedFieldsContextProvider value={{ parentNamePath: "config" }}>
        <>{nestedProviderConfigFields}</>
      </FormNestedFieldsContextProvider>
    </Form>
  );
};

const _default = Object.assign(AccessForm, {
  useProviderFilterByUsage,
});

export default _default;
