import { forwardRef, useImperativeHandle, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import AccessProviderPicker from "@/components/provider/AccessProviderPicker";
import AccessProviderSelect from "@/components/provider/AccessProviderSelect";
import Show from "@/components/Show";
import { type AccessModel } from "@/domain/access";
import { ACCESS_PROVIDERS, ACCESS_USAGES, type AccessProvider } from "@/domain/provider";
import { useAntdForm, useAntdFormName } from "@/hooks";

import AccessForm1PanelConfig from "./AccessForm1PanelConfig";
import AccessFormACMECAConfig from "./AccessFormACMECAConfig";
import AccessFormACMEHttpReqConfig from "./AccessFormACMEHttpReqConfig";
import AccessFormAliyunConfig from "./AccessFormAliyunConfig";
import AccessFormAPISIXConfig from "./AccessFormAPISIXConfig";
import AccessFormAWSConfig from "./AccessFormAWSConfig";
import AccessFormAzureConfig from "./AccessFormAzureConfig";
import AccessFormBaiduCloudConfig from "./AccessFormBaiduCloudConfig";
import AccessFormBaishanConfig from "./AccessFormBaishanConfig";
import AccessFormBaotaPanelConfig from "./AccessFormBaotaPanelConfig";
import AccessFormBaotaWAFConfig from "./AccessFormBaotaWAFConfig";
import AccessFormBunnyConfig from "./AccessFormBunnyConfig";
import AccessFormBytePlusConfig from "./AccessFormBytePlusConfig";
import AccessFormCacheFlyConfig from "./AccessFormCacheFlyConfig";
import AccessFormCdnflyConfig from "./AccessFormCdnflyConfig";
import AccessFormCloudflareConfig from "./AccessFormCloudflareConfig";
import AccessFormClouDNSConfig from "./AccessFormClouDNSConfig";
import AccessFormCMCCCloudConfig from "./AccessFormCMCCCloudConfig";
import AccessFormConstellixConfig from "./AccessFormConstellixConfig";
import AccessFormCTCCCloudConfig from "./AccessFormCTCCCloudConfig";
import AccessFormDeSECConfig from "./AccessFormDeSECConfig";
import AccessFormDigitalOceanConfig from "./AccessFormDigitalOceanConfig";
import AccessFormDingTalkBotConfig from "./AccessFormDingTalkBotConfig";
import AccessFormDiscordBotConfig from "./AccessFormDiscordBotConfig";
import AccessFormDNSLAConfig from "./AccessFormDNSLAConfig";
import AccessFormDogeCloudConfig from "./AccessFormDogeCloudConfig";
import AccessFormDuckDNSConfig from "./AccessFormDuckDNSConfig";
import AccessFormDynv6Config from "./AccessFormDynv6Config";
import AccessFormEdgioConfig from "./AccessFormEdgioConfig";
import AccessFormEmailConfig from "./AccessFormEmailConfig";
import AccessFormFlexCDNConfig from "./AccessFormFlexCDNConfig";
import AccessFormGcoreConfig from "./AccessFormGcoreConfig";
import AccessFormGnameConfig from "./AccessFormGnameConfig";
import AccessFormGoDaddyConfig from "./AccessFormGoDaddyConfig";
import AccessFormGoEdgeConfig from "./AccessFormGoEdgeConfig";
import AccessFormGoogleTrustServicesConfig from "./AccessFormGoogleTrustServicesConfig";
import AccessFormHetznerConfig from "./AccessFormHetznerConfig";
import AccessFormHuaweiCloudConfig from "./AccessFormHuaweiCloudConfig";
import AccessFormJDCloudConfig from "./AccessFormJDCloudConfig";
import AccessFormKongConfig from "./AccessFormKongConfig";
import AccessFormKubernetesConfig from "./AccessFormKubernetesConfig";
import AccessFormLarkBotConfig from "./AccessFormLarkBotConfig";
import AccessFormLeCDNConfig from "./AccessFormLeCDNConfig";
import AccessFormMattermostConfig from "./AccessFormMattermostConfig";
import AccessFormNamecheapConfig from "./AccessFormNamecheapConfig";
import AccessFormNameDotComConfig from "./AccessFormNameDotComConfig";
import AccessFormNameSiloConfig from "./AccessFormNameSiloConfig";
import AccessFormNetcupConfig from "./AccessFormNetcupConfig";
import AccessFormNetlifyConfig from "./AccessFormNetlifyConfig";
import AccessFormNS1Config from "./AccessFormNS1Config";
import AccessFormPorkbunConfig from "./AccessFormPorkbunConfig";
import AccessFormPowerDNSConfig from "./AccessFormPowerDNSConfig";
import AccessFormProxmoxVEConfig from "./AccessFormProxmoxVEConfig";
import AccessFormQiniuConfig from "./AccessFormQiniuConfig";
import AccessFormRainYunConfig from "./AccessFormRainYunConfig";
import AccessFormRatPanelConfig from "./AccessFormRatPanelConfig";
import AccessFormSafeLineConfig from "./AccessFormSafeLineConfig";
import AccessFormSlackBotConfig from "./AccessFormSlackBotConfig";
import AccessFormSpaceshipConfig from "./AccessFormSpaceshipConfig";
import AccessFormSSHConfig from "./AccessFormSSHConfig";
import AccessFormSSLComConfig from "./AccessFormSSLComConfig";
import AccessFormTelegramBotConfig from "./AccessFormTelegramBotConfig";
import AccessFormTencentCloudConfig from "./AccessFormTencentCloudConfig";
import AccessFormUCloudConfig from "./AccessFormUCloudConfig";
import AccessFormUniCloudConfig from "./AccessFormUniCloudConfig";
import AccessFormUpyunConfig from "./AccessFormUpyunConfig";
import AccessFormVercelConfig from "./AccessFormVercelConfig";
import AccessFormVolcEngineConfig from "./AccessFormVolcEngineConfig";
import AccessFormWangsuConfig from "./AccessFormWangsuConfig";
import AccessFormWebhookConfig from "./AccessFormWebhookConfig";
import AccessFormWeComBotConfig from "./AccessFormWeComBotConfig";
import AccessFormWestcnConfig from "./AccessFormWestcnConfig";
import AccessFormZeroSSLConfig from "./AccessFormZeroSSLConfig";

type AccessFormFieldValues = Partial<MaybeModelRecord<AccessModel>>;
type AccessFormScenes = "add" | "edit";
type AccessFormUsages = "both-dns-hosting" | "ca-only" | "notification-only";

export type AccessFormProps = {
  className?: string;
  style?: React.CSSProperties;
  disabled?: boolean;
  initialValues?: AccessFormFieldValues;
  scene: AccessFormScenes;
  usage?: AccessFormUsages;
  onValuesChange?: (values: AccessFormFieldValues) => void;
};

export type AccessFormInstance = {
  getFieldsValue: () => ReturnType<FormInstance<AccessFormFieldValues>["getFieldsValue"]>;
  resetFields: FormInstance<AccessFormFieldValues>["resetFields"];
  validateFields: FormInstance<AccessFormFieldValues>["validateFields"];
};

const AccessForm = forwardRef<AccessFormInstance, AccessFormProps>(({ className, style, disabled, initialValues, usage, scene, onValuesChange }, ref) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    name: z
      .string(t("access.form.name.placeholder"))
      .min(1, t("access.form.name.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    provider: z.nativeEnum(ACCESS_PROVIDERS, {
      message:
        usage === "ca-only"
          ? t("access.form.certificate_authority.placeholder")
          : usage === "notification-only"
            ? t("access.form.notification_channel.placeholder")
            : t("access.form.provider.placeholder"),
    }),
    config: z.any(),
    reserve: z.string().nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm({
    name: "accessForm",
    initialValues: initialValues,
  });

  const providerFilter = useMemo(() => {
    switch (usage) {
      case "both-dns-hosting":
        return (record: AccessProvider) => record.usages.includes(ACCESS_USAGES.DNS) || record.usages.includes(ACCESS_USAGES.HOSTING);
      case "ca-only":
        return (record: AccessProvider) => record.usages.includes(ACCESS_USAGES.CA);
      case "notification-only":
        return (record: AccessProvider) => record.usages.includes(ACCESS_USAGES.NOTIFICATION);
    }

    return undefined;
  }, [usage]);
  const providerLabel = useMemo(() => {
    switch (usage) {
      case "ca-only":
        return t("access.form.certificate_authority.label");
      case "notification-only":
        return t("access.form.notification_channel.label");
    }

    return t("access.form.provider.label");
  }, [usage]);
  const providerPlaceholder = useMemo(() => {
    switch (usage) {
      case "ca-only":
        return t("access.form.certificate_authority.placeholder");
      case "notification-only":
        return t("access.form.notification_channel.placeholder");
    }

    return t("access.form.provider.placeholder");
  }, [usage]);
  const providerTooltip = useMemo(() => {
    switch (usage) {
      case "both-dns-hosting":
        return <span dangerouslySetInnerHTML={{ __html: t("access.form.provider.tooltip") }}></span>;
    }

    return undefined;
  }, [usage]);

  const fieldProvider = Form.useWatch<z.infer<typeof formSchema>["provider"]>("provider", formInst);
  const [fieldProviderPicked, setFieldProviderPicked] = useState<string>(initialValues?.provider); // bugfix: Form.useWatch 在条件渲染下不生效，这里用单独的变量存放 Picker 组件选择的值

  const [nestedFormInst] = Form.useForm();
  const nestedFormName = useAntdFormName({ form: nestedFormInst, name: "accessConfigForm" });
  const nestedFormEl = useMemo(() => {
    const nestedFormProps = {
      form: nestedFormInst,
      formName: nestedFormName,
      disabled: disabled,
      initialValues: initialValues?.config,
    };

    /*
      注意：如果追加新的子组件，请保持以 ASCII 排序。
      NOTICE: If you add new child component, please keep ASCII order.
     */
    switch (fieldProvider) {
      case ACCESS_PROVIDERS["1PANEL"]:
        return <AccessForm1PanelConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.ACMECA:
        return <AccessFormACMECAConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.ACMEHTTPREQ:
        return <AccessFormACMEHttpReqConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.ALIYUN:
        return <AccessFormAliyunConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.APISIX:
        return <AccessFormAPISIXConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.AWS:
        return <AccessFormAWSConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.AZURE:
        return <AccessFormAzureConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.BAIDUCLOUD:
        return <AccessFormBaiduCloudConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.BAISHAN:
        return <AccessFormBaishanConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.BAOTAPANEL:
        return <AccessFormBaotaPanelConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.BAOTAWAF:
        return <AccessFormBaotaWAFConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.BUNNY:
        return <AccessFormBunnyConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.BYTEPLUS:
        return <AccessFormBytePlusConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.CACHEFLY:
        return <AccessFormCacheFlyConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.CDNFLY:
        return <AccessFormCdnflyConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.CLOUDFLARE:
        return <AccessFormCloudflareConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.CLOUDNS:
        return <AccessFormClouDNSConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.CMCCCLOUD:
        return <AccessFormCMCCCloudConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.CONSTELLIX:
        return <AccessFormConstellixConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.CTCCCLOUD:
        return <AccessFormCTCCCloudConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.DESEC:
        return <AccessFormDeSECConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.DIGITALOCEAN:
        return <AccessFormDigitalOceanConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.DINGTALKBOT:
        return <AccessFormDingTalkBotConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.DISCORDBOT:
        return <AccessFormDiscordBotConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.DNSLA:
        return <AccessFormDNSLAConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.DOGECLOUD:
        return <AccessFormDogeCloudConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.DUCKDNS:
        return <AccessFormDuckDNSConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.DYNV6:
        return <AccessFormDynv6Config {...nestedFormProps} />;
      case ACCESS_PROVIDERS.EDGIO:
        return <AccessFormEdgioConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.EMAIL:
        return <AccessFormEmailConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.FLEXCDN:
        return <AccessFormFlexCDNConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.GCORE:
        return <AccessFormGcoreConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.GNAME:
        return <AccessFormGnameConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.GODADDY:
        return <AccessFormGoDaddyConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.GOEDGE:
        return <AccessFormGoEdgeConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.GOOGLETRUSTSERVICES:
        return <AccessFormGoogleTrustServicesConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.HETZNER:
        return <AccessFormHetznerConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.HUAWEICLOUD:
        return <AccessFormHuaweiCloudConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.JDCLOUD:
        return <AccessFormJDCloudConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.KONG:
        return <AccessFormKongConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.KUBERNETES:
        return <AccessFormKubernetesConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.LARKBOT:
        return <AccessFormLarkBotConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.LECDN:
        return <AccessFormLeCDNConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.MATTERMOST:
        return <AccessFormMattermostConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.NAMECHEAP:
        return <AccessFormNamecheapConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.NAMEDOTCOM:
        return <AccessFormNameDotComConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.NAMESILO:
        return <AccessFormNameSiloConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.NETCUP:
        return <AccessFormNetcupConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.NETLIFY:
        return <AccessFormNetlifyConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.NS1:
        return <AccessFormNS1Config {...nestedFormProps} />;
      case ACCESS_PROVIDERS.PORKBUN:
        return <AccessFormPorkbunConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.POWERDNS:
        return <AccessFormPowerDNSConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.PROXMOXVE:
        return <AccessFormProxmoxVEConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.QINIU:
        return <AccessFormQiniuConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.RAINYUN:
        return <AccessFormRainYunConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.RATPANEL:
        return <AccessFormRatPanelConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.SAFELINE:
        return <AccessFormSafeLineConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.SLACKBOT:
        return <AccessFormSlackBotConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.SPACESHIP:
        return <AccessFormSpaceshipConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.SSH:
        return <AccessFormSSHConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.TELEGRAMBOT:
        return <AccessFormTelegramBotConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.SSLCOM:
        return <AccessFormSSLComConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.TENCENTCLOUD:
        return <AccessFormTencentCloudConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.UCLOUD:
        return <AccessFormUCloudConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.UNICLOUD:
        return <AccessFormUniCloudConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.UPYUN:
        return <AccessFormUpyunConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.VERCEL:
        return <AccessFormVercelConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.VOLCENGINE:
        return <AccessFormVolcEngineConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.WANGSU:
        return <AccessFormWangsuConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.WEBHOOK:
        return (
          <AccessFormWebhookConfig
            usage={usage === "notification-only" ? "notification" : usage === "both-dns-hosting" ? "deployment" : "none"}
            {...nestedFormProps}
          />
        );
      case ACCESS_PROVIDERS.WECOMBOT:
        return <AccessFormWeComBotConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.WESTCN:
        return <AccessFormWestcnConfig {...nestedFormProps} />;
      case ACCESS_PROVIDERS.ZEROSSL:
        return <AccessFormZeroSSLConfig {...nestedFormProps} />;
    }
  }, [usage, disabled, initialValues?.config, fieldProvider, nestedFormInst, nestedFormName]);

  const handleProviderPick = (value: string) => {
    setFieldProviderPicked(value);
    formInst.setFieldValue("provider", value);
    onValuesChange?.(formInst.getFieldsValue(true));
  };

  const handleFormProviderChange = (name: string) => {
    if (name === nestedFormName) {
      formInst.setFieldValue("config", nestedFormInst.getFieldsValue());
      onValuesChange?.(formInst.getFieldsValue(true));
    }
  };

  const handleFormChange = (_: unknown, values: AccessFormFieldValues) => {
    onValuesChange?.(values);
  };

  useImperativeHandle(ref, () => {
    return {
      getFieldsValue: () => {
        const values = formInst.getFieldsValue(true);
        values.config = nestedFormInst.getFieldsValue();
        values.reserve = usage === "ca-only" ? "ca" : usage === "notification-only" ? "notification" : undefined;
        return values;
      },
      resetFields: (fields) => {
        formInst.resetFields(fields);

        if (!!fields && fields.includes("config")) {
          nestedFormInst.resetFields(fields);
        }
      },
      validateFields: (nameList, config) => {
        const t1 = formInst.validateFields(nameList, config);
        const t2 = nestedFormInst.validateFields(undefined, config);
        return Promise.all([t1, t2]).then(() => t1);
      },
    } as AccessFormInstance;
  });

  return (
    <Form.Provider onFormChange={handleFormProviderChange}>
      <div className={className} style={style}>
        <Form {...formProps} disabled={disabled} layout="vertical" scrollToFirstError onValuesChange={handleFormChange}>
          <Show
            when={!!fieldProvider || !!fieldProviderPicked}
            fallback={
              <AccessProviderPicker
                autoFocus
                filter={providerFilter}
                placeholder={t("access.form.provider.search.placeholder")}
                showOptionTags={usage == null || (usage === "both-dns-hosting" ? { [ACCESS_USAGES.DNS]: true, [ACCESS_USAGES.HOSTING]: true } : false)}
                onSelect={handleProviderPick}
              />
            }
          >
            <Form.Item name="name" label={t("access.form.name.label")} rules={[formRule]}>
              <Input placeholder={t("access.form.name.placeholder")} />
            </Form.Item>

            <Form.Item name="provider" label={providerLabel} rules={[formRule]} tooltip={providerTooltip}>
              <AccessProviderSelect
                filter={providerFilter}
                disabled={scene !== "add"}
                placeholder={providerPlaceholder}
                showOptionTags={usage == null || (usage === "both-dns-hosting" ? { [ACCESS_USAGES.DNS]: true, [ACCESS_USAGES.HOSTING]: true } : false)}
                showSearch={!disabled}
              />
            </Form.Item>
          </Show>
        </Form>

        {nestedFormEl}
      </div>
    </Form.Provider>
  );
});

export default AccessForm;
