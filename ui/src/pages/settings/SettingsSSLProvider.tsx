import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMount } from "ahooks";
import { App, Button, Card, Divider, Form, Input, Select, Skeleton } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { produce } from "immer";
import { z } from "zod";

import Show from "@/components/Show";
import Tips from "@/components/Tips";
import { type CAProviderType, CA_PROVIDERS } from "@/domain/provider";
import { type SSLProviderSettingsContent } from "@/domain/settings";
import { useAntdForm, useZustandShallowSelector } from "@/hooks";
import { useSSLProviderSettingsStore } from "@/stores/settings";
import { mergeCls } from "@/utils/css";
import { unwrapErrMsg } from "@/utils/error";

const SettingsSSLProvider = () => {
  const { t } = useTranslation();

  const { message, notification } = App.useApp();

  const { settings, loading, loadSettings, saveSettings } = useSSLProviderSettingsStore(
    useZustandShallowSelector(["settings", "loading", "loadSettings", "saveSettings"])
  );
  useMount(() => loadSettings());

  const updateContextSettings = async (settings: SSLProviderSettingsContent) => {
    try {
      await saveSettings(settings);

      message.success(t("common.text.operation_succeeded"));
    } catch (err) {
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
    }
  };

  return (
    <InternalSettingsContext.Provider
      value={{
        loading: loading,
        settings: settings!,
        updateSettings: updateContextSettings,
      }}
    >
      <h2>{t("settings.sslprovider.ca.title")}</h2>
      <SettingsSSLProviderCA />

      <Divider />

      <h2>{t("settings.sslprovider.others.title")}</h2>
      <SettingsSSLProviderOthers className="md:max-w-160" />
    </InternalSettingsContext.Provider>
  );
};

const SettingsSSLProviderCA = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const { t } = useTranslation();

  const { loading, settings } = useContext(InternalSettingsContext);

  const formSchema = z.object({
    provider: z.string().nonempty(),
    configs: z.object().nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const [formInst] = Form.useForm<z.infer<typeof formSchema>>();

  const providers = [
    [CA_PROVIDERS.LETSENCRYPT, "provider.letsencrypt", "letsencrypt.org", "/imgs/providers/letsencrypt.svg"],
    [CA_PROVIDERS.LETSENCRYPTSTAGING, "provider.letsencryptstaging", "letsencrypt.org", "/imgs/providers/letsencrypt.svg"],
    [CA_PROVIDERS.ACTALISSSL, "provider.actalisssl", "actalis.com", "/imgs/providers/actalisssl.png"],
    [CA_PROVIDERS.GLOBALSIGNATLAS, "provider.globalsignatlas", "atlas.globalsign.com", "/imgs/providers/globalsignatlas.png"],
    [CA_PROVIDERS.GOOGLETRUSTSERVICES, "provider.googletrustservices", "pki.goog", "/imgs/providers/google.svg"],
    [CA_PROVIDERS.SECTIGO, "provider.sectigo", "sectigo.com", "/imgs/providers/sectigo.svg"],
    [CA_PROVIDERS.SSLCOM, "provider.sslcom", "ssl.com", "/imgs/providers/sslcom.svg"],
    [CA_PROVIDERS.ZEROSSL, "provider.zerossl", "zerossl.com", "/imgs/providers/zerossl.svg"],
    [CA_PROVIDERS.LITESSL, "provider.litessl", "litessl.cn (freessl.cn)", "/imgs/providers/litessl.svg"],
    [CA_PROVIDERS.ACMECA, "provider.acmeca", "ACME v2 (RFC 8555)", "/imgs/providers/acmeca.svg"],
  ].map(([value, name, description, icon]) => {
    return {
      value: value as CAProviderType,
      name: t(name),
      description,
      icon,
    };
  });
  const [providerValue, setProviderValue] = useState(settings.provider);

  const renderSiblingFieldProviderComponent = useMemo(() => {
    switch (providerValue) {
      case CA_PROVIDERS.LETSENCRYPT:
        return <InternalCASettingsFormProviderLetsEncrypt />;
      case CA_PROVIDERS.LETSENCRYPTSTAGING:
        return <InternalCASettingsFormProviderLetsEncryptStaging />;
      case CA_PROVIDERS.ACTALISSSL:
        return <InternalCASettingsFormProviderActalisSSL />;
      case CA_PROVIDERS.GLOBALSIGNATLAS:
        return <InternalCASettingsFormProviderGlobalSignAtlas />;
      case CA_PROVIDERS.GOOGLETRUSTSERVICES:
        return <InternalCASettingsFormProviderGoogleTrustServices />;
      case CA_PROVIDERS.LITESSL:
        return <InternalCASettingsFormProviderLiteSSL />;
      case CA_PROVIDERS.SECTIGO:
        return <InternalCASettingsFormProviderSectigo />;
      case CA_PROVIDERS.SSLCOM:
        return <InternalCASettingsFormProviderSSLCom />;
      case CA_PROVIDERS.ZEROSSL:
        return <InternalCASettingsFormProviderZeroSSL />;
      case CA_PROVIDERS.ACMECA:
        return <InternalCASettingsFormProviderACMECA />;
    }
  }, [providerValue]);

  useEffect(() => {
    setProviderValue(settings.provider);
  }, [settings.provider]);

  return (
    <div className={className} style={style}>
      <Show when={!loading} fallback={<Skeleton active />}>
        <Form form={formInst} layout="vertical" initialValues={{ provider: providerValue }}>
          <Form.Item>
            <Tips message={<span dangerouslySetInnerHTML={{ __html: t("settings.sslprovider.ca.tips") }}></span>} />
          </Form.Item>

          <Form.Item
            name="provider"
            label={t("settings.sslprovider.ca.form.provider.label")}
            extra={t("settings.sslprovider.ca.form.provider.help")}
            rules={[formRule]}
          >
            <div className="flex w-full flex-wrap items-center gap-4">
              {providers.map((provider) => (
                <Card
                  key={provider.value}
                  className={mergeCls("relative overflow-hidden", { ["border-primary"]: providerValue === provider.value })}
                  style={{ width: 280 }}
                  styles={{
                    body: { padding: 0 },
                  }}
                  hoverable
                  onClick={() => setProviderValue(provider.value)}
                >
                  <div className="relative z-1 px-3 py-4">
                    <div className="flex items-center justify-between gap-3">
                      <div>
                        <img src={provider.icon} className="size-8" />
                      </div>
                      <div className="flex-1 overflow-hidden">
                        <div className="truncate">{provider.name}</div>
                        <div className="mt-1 truncate text-xs">{provider.description}</div>
                      </div>
                    </div>
                  </div>
                  {providerValue === provider.value && <div className="absolute top-0 left-0 size-full bg-primary opacity-20"></div>}
                </Card>
              ))}
            </div>
          </Form.Item>
        </Form>

        <div className="md:max-w-160">{renderSiblingFieldProviderComponent}</div>
      </Show>
    </div>
  );
};

const SettingsSSLProviderOthers = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const { t } = useTranslation();

  const { loading, settings, updateSettings } = useContext(InternalSettingsContext);

  const formSchema = z.object({
    timeout: z.union([z.string(), z.number().int().positive()]).nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    initialValues: { timeout: settings.timeout },
    onSubmit: async (values) => {
      setFormPending(true);

      try {
        const newSettings = produce(settings, (draft) => {
          draft.timeout = +values.timeout!;
        });
        await updateSettings(newSettings);
      } finally {
        setFormPending(false);
      }

      setFormChanged(false);
    },
  });
  const [formPending, setFormPending] = useState(false);
  const [formChanged, setFormChanged] = useState(false);

  const handleFormChange = () => {
    setFormChanged(true);
  };

  return (
    <div className={className} style={style}>
      <Show when={!loading} fallback={<Skeleton active />}>
        <Form {...formProps} form={formInst} disabled={formPending} layout="vertical" onValuesChange={handleFormChange}>
          <Form.Item
            name="timeout"
            label={t("settings.sslprovider.others.form.timeout.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.sslprovider.others.form.timeout.tooltip") }}></span>}
          >
            <Input
              type="number"
              allowClear
              min={0}
              max={3600}
              placeholder={t("settings.sslprovider.others.form.timeout.placeholder")}
              suffix={t("settings.sslprovider.others.form.timeout.unit")}
            />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" disabled={!formChanged} loading={formPending}>
              {t("common.button.save")}
            </Button>
          </Form.Item>
        </Form>
      </Show>
    </div>
  );
};

const InternalSettingsContext = createContext(
  {} as {
    loading: boolean;
    settings: SSLProviderSettingsContent;
    updateSettings: (settings: SSLProviderSettingsContent) => Promise<void>;
  }
);

const InternalCASharedForm = ({ children, provider }: { children?: React.ReactNode; provider: CAProviderType }) => {
  const { t } = useTranslation();

  const { settings, updateSettings } = useContext(InternalSettingsContext);

  const { form: formInst, formProps } = useAntdForm<NonNullable<unknown>>({
    initialValues: settings?.configs?.[provider],
    onSubmit: async (values) => {
      setFormPending(true);

      try {
        const newSettings = produce(settings, (draft) => {
          draft.provider = provider;
          draft.configs ??= {} as SSLProviderSettingsContent["configs"];
          draft.configs[provider] = values;
        });
        await updateSettings(newSettings);
      } finally {
        setFormPending(false);
      }

      setFormChanged(false);
    },
  });
  const [formPending, setFormPending] = useState(false);
  const [formChanged, setFormChanged] = useState(false);

  useEffect(() => {
    setFormChanged(provider !== settings?.provider);
  }, [provider, settings?.provider]);

  const handleFormChange = () => {
    setFormChanged(true);
  };

  return (
    <Form {...formProps} form={formInst} disabled={formPending} layout="vertical" onValuesChange={handleFormChange}>
      {children}

      <Form.Item>
        <Button type="primary" htmlType="submit" disabled={!formChanged} loading={formPending}>
          {t("common.button.save")}
        </Button>
      </Form.Item>
    </Form>
  );
};

const InternalCASharedFormEabFields = ({ i18nKey }: { i18nKey: string }) => {
  const { t, i18n } = useTranslation();

  const hasGuide = i18n.exists(`access.form.${i18nKey}_eab.guide`);

  const formSchema = z.object({
    endpoint: z.url(t("common.errmsg.url_invalid")),
    eabKid: z.string(t("access.form.shared_acme_eab_kid.label")).nonempty(t("access.form.shared_acme_eab_kid.placeholder")),
    eabHmacKey: z.string(t("access.form.shared_acme_eab_hmac_key.label")).nonempty(t("access.form.shared_acme_eab_hmac_key.placeholder")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  return (
    <>
      <Form.Item name="eabKid" label={t("access.form.shared_acme_eab_kid.label")} rules={[formRule]}>
        <Input autoComplete="new-password" placeholder={t("access.form.shared_acme_eab_kid.placeholder")} />
      </Form.Item>

      <Form.Item name="eabHmacKey" label={t("access.form.shared_acme_eab_hmac_key.label")} rules={[formRule]}>
        <Input.Password autoComplete="new-password" placeholder={t("access.form.shared_acme_eab_hmac_key.placeholder")} />
      </Form.Item>

      <Form.Item hidden={!hasGuide}>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t(`access.form.${i18nKey}_eab.guide`) }}></span>} />
      </Form.Item>
    </>
  );
};

const InternalCASettingsFormProviderLetsEncrypt = () => {
  return <InternalCASharedForm provider={CA_PROVIDERS.LETSENCRYPT} />;
};

const InternalCASettingsFormProviderLetsEncryptStaging = () => {
  const { t } = useTranslation();

  return (
    <InternalCASharedForm provider={CA_PROVIDERS.LETSENCRYPTSTAGING}>
      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("settings.sslprovider.ca.form.letsencryptstaging_alert") }}></span>} />
      </Form.Item>
    </InternalCASharedForm>
  );
};

const InternalCASettingsFormProviderActalisSSL = () => {
  return (
    <InternalCASharedForm provider={CA_PROVIDERS.ACTALISSSL}>
      <InternalCASharedFormEabFields i18nKey="actalisssl" />
    </InternalCASharedForm>
  );
};

const InternalCASettingsFormProviderGlobalSignAtlas = () => {
  return (
    <InternalCASharedForm provider={CA_PROVIDERS.GLOBALSIGNATLAS}>
      <InternalCASharedFormEabFields i18nKey="globalsignatlas" />
    </InternalCASharedForm>
  );
};

const InternalCASettingsFormProviderGoogleTrustServices = () => {
  return (
    <InternalCASharedForm provider={CA_PROVIDERS.GOOGLETRUSTSERVICES}>
      <InternalCASharedFormEabFields i18nKey="googletrustservices" />
    </InternalCASharedForm>
  );
};

const InternalCASettingsFormProviderLiteSSL = () => {
  return (
    <InternalCASharedForm provider={CA_PROVIDERS.LITESSL}>
      <InternalCASharedFormEabFields i18nKey="litessl" />
    </InternalCASharedForm>
  );
};

const InternalCASettingsFormProviderSectigo = () => {
  const { t } = useTranslation();

  const formSchema = z.object({
    validationType: z.string().nonempty(t("access.form.sectigo_validation_type.placeholder")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  return (
    <InternalCASharedForm provider={CA_PROVIDERS.SECTIGO}>
      <Form.Item name="validationType" initialValue="dv" label={t("access.form.sectigo_validation_type.label")} rules={[formRule]}>
        <Select
          options={["dv", "ov", "ev"].map((s) => ({
            key: s,
            label: t(`access.form.sectigo_validation_type.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("access.form.sectigo_validation_type.placeholder")}
        />
      </Form.Item>

      <InternalCASharedFormEabFields i18nKey="sectigo" />
    </InternalCASharedForm>
  );
};

const InternalCASettingsFormProviderSSLCom = () => {
  return (
    <InternalCASharedForm provider={CA_PROVIDERS.SSLCOM}>
      <InternalCASharedFormEabFields i18nKey="sslcom" />
    </InternalCASharedForm>
  );
};

const InternalCASettingsFormProviderZeroSSL = () => {
  return (
    <InternalCASharedForm provider={CA_PROVIDERS.ZEROSSL}>
      <InternalCASharedFormEabFields i18nKey="zerossl" />
    </InternalCASharedForm>
  );
};

const InternalCASettingsFormProviderACMECA = () => {
  const { t } = useTranslation();

  const formSchema = z.object({
    endpoint: z.url(t("common.errmsg.url_invalid")),
    eabKid: z.string(t("access.form.acmeca_eab_kid.placeholder")).nullish(),
    eabHmacKey: z.string(t("access.form.acmeca_eab_hmac_key.placeholder")).nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);

  return (
    <InternalCASharedForm provider={CA_PROVIDERS.ACMECA}>
      <Form.Item
        name="endpoint"
        label={t("access.form.acmeca_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmeca_endpoint.tooltip") }}></span>}
      >
        <Input placeholder={t("access.form.acmeca_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name="eabKid"
        label={t("access.form.acmeca_eab_kid.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmeca_eab_kid.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.acmeca_eab_kid.placeholder")} />
      </Form.Item>

      <Form.Item
        name="eabHmacKey"
        label={t("access.form.acmeca_eab_hmac_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmeca_eab_hmac_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.acmeca_eab_hmac_key.placeholder")} />
      </Form.Item>
    </InternalCASharedForm>
  );
};

export default SettingsSSLProvider;
