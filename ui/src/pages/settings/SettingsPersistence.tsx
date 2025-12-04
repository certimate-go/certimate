import { createContext, useContext, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMount } from "ahooks";
import { App, Button, Divider, Form, InputNumber, Skeleton } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { produce } from "immer";
import { z } from "zod";

import Show from "@/components/Show";
import { type PersistenceSettingsContent } from "@/domain/settings";
import { useAntdForm, useZustandShallowSelector } from "@/hooks";
import { usePersistenceSettingsStore } from "@/stores/settings";
import { getErrMsg } from "@/utils/error";

const SettingsPersistence = () => {
  const { t } = useTranslation();

  const { message, notification } = App.useApp();

  const { settings, loading, loadSettings, saveSettings } = usePersistenceSettingsStore(
    useZustandShallowSelector(["settings", "loading", "loadSettings", "saveSettings"])
  );
  useMount(() => loadSettings());

  const updateContextSettings = async (settings: PersistenceSettingsContent) => {
    try {
      await saveSettings(settings);

      message.success(t("common.text.operation_succeeded"));
    } catch (err) {
      notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
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
      <h2>{t("settings.persistence.alerting.title")}</h2>
      <SettingsPersistenceAlerting className="md:max-w-160" />

      <Divider />

      <h2>{t("settings.persistence.data_retention.title")}</h2>
      <SettingsPersistenceDataRetention className="md:max-w-160" />
    </InternalSettingsContext.Provider>
  );
};

const SettingsPersistenceAlerting = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const { t } = useTranslation();

  const { loading, settings, updateSettings } = useContext(InternalSettingsContext);

  const formSchema = z.object({
    certificatesWarningDaysBeforeExpire: z.number().int().positive(),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const {
    form: formInst,
    formPending,
    formProps,
  } = useAntdForm<z.infer<typeof formSchema>>({
    initialValues: {
      certificatesWarningDaysBeforeExpire: settings?.certificatesWarningDaysBeforeExpire,
    },
    onSubmit: async (values) => {
      updateSettings(
        produce(settings!, (draft) => {
          draft.certificatesWarningDaysBeforeExpire = values.certificatesWarningDaysBeforeExpire;
        })
      );
    },
  });
  const [formChanged, setFormChanged] = useState(false);

  const handleInputChange = () => {
    const changed = formInst.getFieldValue("certificatesWarningDaysBeforeExpire") !== formProps.initialValues?.certificatesWarningDaysBeforeExpire;
    setFormChanged(changed);
  };

  return (
    <>
      <div className={className} style={style}>
        <Show when={!loading} fallback={<Skeleton active />}>
          <Form {...formProps} form={formInst} disabled={formPending} layout="vertical">
            <Form.Item
              name="certificatesWarningDaysBeforeExpire"
              label={t("settings.persistence.alerting.form.certificates_warning_days_before_expire.label")}
              extra={<span dangerouslySetInnerHTML={{ __html: t("settings.persistence.alerting.form.certificates_warning_days_before_expire.help") }}></span>}
              rules={[formRule]}
            >
              <InputNumber
                style={{ width: "100%" }}
                min={1}
                max={365}
                placeholder={t("settings.persistence.alerting.form.certificates_warning_days_before_expire.placeholder")}
                suffix={t("settings.persistence.alerting.form.certificates_warning_days_before_expire.unit")}
                onChange={handleInputChange}
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
    </>
  );
};

const SettingsPersistenceDataRetention = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const { t } = useTranslation();

  const { loading, settings, updateSettings } = useContext(InternalSettingsContext);

  const formSchema = z.object({
    certificatesRetentionMaxDays: z.number().int().nonnegative(),
    workflowRunsRetentionMaxDays: z.number().int().nonnegative(),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const {
    form: formInst,
    formPending,
    formProps,
  } = useAntdForm<z.infer<typeof formSchema>>({
    initialValues: {
      certificatesRetentionMaxDays: settings?.certificatesRetentionMaxDays,
      workflowRunsRetentionMaxDays: settings?.workflowRunsRetentionMaxDays,
    },
    onSubmit: async (values) => {
      updateSettings(
        produce(settings!, (draft) => {
          draft.certificatesRetentionMaxDays = values.certificatesRetentionMaxDays;
          draft.workflowRunsRetentionMaxDays = values.workflowRunsRetentionMaxDays;
        })
      );
    },
  });
  const [formChanged, setFormChanged] = useState(false);

  const handleInputChange = () => {
    const changed =
      formInst.getFieldValue("certificatesRetentionMaxDays") !== formProps.initialValues?.certificatesRetentionMaxDays ||
      formInst.getFieldValue("workflowRunsRetentionMaxDays") !== formProps.initialValues?.workflowRunsRetentionMaxDays;
    setFormChanged(changed);
  };

  return (
    <>
      <div className={className} style={style}>
        <Show when={!loading} fallback={<Skeleton active />}>
          <Form {...formProps} form={formInst} disabled={formPending} layout="vertical">
            <Form.Item
              name="certificatesRetentionMaxDays"
              label={t("settings.persistence.data_retention.form.certificates_retention_max_days.label")}
              extra={<span dangerouslySetInnerHTML={{ __html: t("settings.persistence.data_retention.form.certificates_retention_max_days.help") }}></span>}
              rules={[formRule]}
            >
              <InputNumber
                style={{ width: "100%" }}
                min={0}
                max={36500}
                placeholder={t("settings.persistence.data_retention.form.certificates_retention_max_days.placeholder")}
                suffix={t("settings.persistence.data_retention.form.certificates_retention_max_days.unit")}
                onChange={handleInputChange}
              />
            </Form.Item>

            <Form.Item
              name="workflowRunsRetentionMaxDays"
              label={t("settings.persistence.data_retention.form.workflow_runs_retention_max_days.label")}
              extra={<span dangerouslySetInnerHTML={{ __html: t("settings.persistence.data_retention.form.workflow_runs_retention_max_days.help") }}></span>}
              rules={[formRule]}
            >
              <InputNumber
                style={{ width: "100%" }}
                min={0}
                max={36500}
                placeholder={t("settings.persistence.data_retention.form.workflow_runs_retention_max_days.placeholder")}
                suffix={t("settings.persistence.data_retention.form.workflow_runs_retention_max_days.unit")}
                onChange={handleInputChange}
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
    </>
  );
};

const InternalSettingsContext = createContext(
  {} as {
    loading: boolean;
    settings: PersistenceSettingsContent;
    updateSettings: (settings: PersistenceSettingsContent) => Promise<void>;
  }
);

export default SettingsPersistence;
