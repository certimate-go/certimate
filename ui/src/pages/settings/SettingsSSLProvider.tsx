import { createContext, useContext, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMount } from "ahooks";
import { App, Button, Card, Form, Input, Skeleton } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { produce } from "immer";
import { z } from "zod";

import Show from "@/components/Show";
import { type SSLProviderSettingsContent } from "@/domain/settings";
import { useAntdForm, useZustandShallowSelector } from "@/hooks";
import { useSSLProviderSettingsStore } from "@/stores/settings";
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
      <div>
        <SettingsSSLProviderOthers />
      </div>
    </InternalSettingsContext.Provider>
  );
};

const SettingsSSLProviderOthers = () => {
  const { t } = useTranslation();

  const { loading, settings, updateSettings } = useContext(InternalSettingsContext);

  const formSchema = z.object({
    timeout: z.union([z.string(), z.int().positive()]).nullish(),
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
    <Card
      title={
        <div className="mt-2 mb-0 overflow-hidden">
          <h2 className="truncate text-base font-semibold">{t("settings.sslprovider.others.title")}</h2>
        </div>
      }
    >
      <Show when={!loading} fallback={<Skeleton active />}>
        <Form className="md:max-w-160" {...formProps} form={formInst} disabled={formPending} layout="vertical" onValuesChange={handleFormChange}>
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

          <Form.Item noStyle>
            <Button type="primary" htmlType="submit" disabled={!formChanged} loading={formPending}>
              {t("common.button.save")}
            </Button>
          </Form.Item>
        </Form>
      </Show>
    </Card>
  );
};

const InternalSettingsContext = createContext(
  {} as {
    loading: boolean;
    settings: SSLProviderSettingsContent;
    updateSettings: (settings: SSLProviderSettingsContent) => Promise<void>;
  }
);

export default SettingsSSLProvider;
